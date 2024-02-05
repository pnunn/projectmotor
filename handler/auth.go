package handler

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/sessions"
	"github.com/pnunn/projectmotor/auth"
	"github.com/pnunn/projectmotor/github"
	"github.com/pnunn/projectmotor/template"
	"golang.org/x/net/context"
	"net/http"
)

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	template.Login.ExecuteWriter(pongo2.Context{}, w)
}

func generateCSRFToken(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h Handler) OAuthGitHubLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateCSRFToken(16)
	if err != nil {
		fail(w, err, http.StatusInternalServerError)
		return
	}
	session, err := h.GetSessionStore(r)
	if err != nil {
		fail(w, err, http.StatusInternalServerError)
		return
	}
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		fail(w, err, http.StatusInternalServerError)
		return
	}
	url := github.Config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h Handler) OAuthGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Get session store
	session, err := h.GetSessionStore(r)
	if err != nil {
		fail(w, err, http.StatusInternalServerError)
		return
	}
	// Get state and code from url query (?state=foo&code=bar)
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	// Check if state matches between query and session
	if stateMatches(state, session) {
		// Exchange code for token
		token, err := github.Config.Exchange(context.Background(), code)
		if err != nil {
			fail(w, err, http.StatusInternalServerError)
			return
		}
		gh := github.New(token.AccessToken)
		data, err := gh.GetData()
		if err != nil {
			fail(w, err, http.StatusInternalServerError)
			return
		}
		// Check if account with ID exists
		account, exists, err := h.AccountService.GetAccountByID(data.ID)
		if err != nil {
			fail(w, err, http.StatusInternalServerError)
			return
		}
		if !exists {
			// Begin transaction
			tx, err := h.BeginTx(r.Context())
			defer tx.Rollback()
			if err != nil {
				fail(w, err, http.StatusInternalServerError)
				return
			}
			// Create user within transaction
			user, err := h.UserService.CreateUser(tx, data.PrimaryEmail)
			if err != nil {
				fail(w, err, http.StatusInternalServerError)
				return
			}
			// Create account within transaction
			_, err = h.AccountService.CreateAccount(tx, data.ID, user.ID, token.AccessToken)
			if err != nil {
				fail(w, err, http.StatusInternalServerError)
				return
			}
			// Commit transaction
			err = tx.Commit()
			if err != nil {
				fail(w, err, http.StatusInternalServerError)
				return
			}
			err = auth.SetUserSession(w, r, user.ID, session)
			if err != nil {
				fail(w, err, http.StatusInternalServerError)
				return
			}
		}
		err = auth.SetUserSession(w, r, account.UserID, session)
		if err != nil {
			fail(w, err, http.StatusInternalServerError)
			return
		}
		return
	}
	fail(w, err, http.StatusUnauthorized)
}

func stateMatches(s string, session *sessions.Session) bool {
	state := session.Values["state"]
	if state == nil {
		return false
	}
	return s == state
}
