package handler

import (
	"github.com/pnunn/projectmotor/template"
	"golang.org/x/net/context"
	"net/http"
)

func (h Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	//user := h.GetUserFromContext(r.Context())
	//template.Dashboard.ExecuteWriter(pongo2.Context{"message": fmt.Sprintf("Welcome back, %s!", user.Email)}, w)
	component := template.Hello("Peter")
	component.Render(context.Background(), w)
}
