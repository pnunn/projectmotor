package handler

import (
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/pnunn/projectmotor/auth"
	"github.com/pnunn/projectmotor/database"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

type Handler struct {
	UserService    *database.UserService
	AccountService *database.AccountService
	Store          *sessions.CookieStore
	Db             *sqlx.DB
}

type HandlerOptions struct {
	DB    *sqlx.DB
	Store *sessions.CookieStore
}

func NewHandler(options HandlerOptions) *Handler {
	userService := database.NewUserService(options.DB)
	accountService := database.NewAccountService(options.DB)
	return &Handler{
		Store:          options.Store,
		Db:             options.DB,
		UserService:    userService,
		AccountService: accountService,
	}
}

func (h Handler) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := h.Db.BeginTxx(ctx, nil)
	if err != nil {
		return &sqlx.Tx{}, err
	}
	return tx, nil
}

func (h Handler) GetSessionStore(r *http.Request) (*sessions.Session, error) {
	return h.Store.Get(r, "_projectmotor_session")
}

func (h Handler) GetUserFromContext(ctx context.Context) database.User {
	user := ctx.Value(auth.UserKey{})
	if user, ok := user.(database.User); ok {

		println("User: ")
		println(user.Name.String)
		return user
	}
	println("NOT OK")
	return database.User{}
}

func fail(w http.ResponseWriter, err error, code int) {
	http.Error(w, http.StatusText(code), code)
	log.Println("error:", err)
}
