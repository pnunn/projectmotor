package handler

import (
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/pnunn/projectmotor/database"
	"log"
	"net/http"
)

type Handler struct {
	userService    *database.UserService
	accountService *database.AccountService
	store          *sessions.CookieStore
}

type HandlerOptions struct {
	DB    *sqlx.DB
	Store *sessions.CookieStore
}

func NewHandler(options HandlerOptions) *Handler {
	userService := database.NewUserService(options.DB)
	accountService := database.NewAccountService(options.DB)
	return &Handler{
		userService:    userService,
		accountService: accountService,
		store:          options.Store,
	}
}

func (h Handler) GetSessionStore(r *http.Request) (*sessions.Session, error) {
	return h.store.Get(r, "_projectmotor_session")
}

func fail(w http.ResponseWriter, err error, code int) {
	http.Error(w, http.StatusText(code), code)
	log.Println("error:", err)
}
