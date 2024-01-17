package main

import (
	"context"
	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/pnunn/projectmotor/auth"
	"github.com/pnunn/projectmotor/database"
	"github.com/pnunn/projectmotor/handler"
	"github.com/pnunn/projectmotor/template"
	"log"
	"net/http"
)

var store = sessions.NewCookieStore([]byte("5560b5496286320578ec68b78c64161a19d86264d047271bd1e63675cc00d04a"))

func main() {
	db, err := database.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	err = database.SetupDB()
	if err != nil {
		log.Fatal(err)
	}
	h := handler.NewHandler(handler.HandlerOptions{
		DB:    db,
		Store: store,
	})
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/login", h.Login)
	r.Get("/oauth/github/login", h.OAuthGitHubLogin)
	r.Get("/oauth/github/callback", h.OAuthGitHubCallback)

	// group protected routes into one function to run middleware
	r.Group(protectedRouter(h))
	http.ListenAndServe("localhost:3000", r)
}

// Redirect to public auth route
//
// Use this when session user doesn't exist
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://localhost:3000/login", http.StatusSeeOther)
}

// Protected context
//
// Middleware checks if user exists within current session
func ProtectedCtx(h *handler.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			session, err := h.GetSessionStore(r)
			// redirect in case of error
			if err != nil {
				redirectToLogin(w, r)
				return
			}
			user := session.Values["user"]
			// redirect in case of missing user
			if user == nil {
				redirectToLogin(w, r)
				return
			}
			// check if user is db.User
			ctx := r.Context()
			if user, ok := user.(database.User); ok {
				ctx = context.WithValue(ctx, auth.UserIDKey{}, user.ID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// Router with user ensured
//
// Add routes here where user has to be logged in
func protectedRouter(h *handler.Handler) func(chi.Router) {
	return func(r chi.Router) {
		r.Use(ProtectedCtx(h))
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			//template.Dashboard.ExecuteWriter(pongo2.Context{}, w)
			// NOTE ->> only for testing, remove after actual interactions with database
			message, err := database.GetMessage()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			// <<- NOTE
			template.Dashboard.ExecuteWriter(pongo2.Context{
				"message": message,
			}, w)
		})
	}
}