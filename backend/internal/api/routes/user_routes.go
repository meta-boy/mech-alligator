package routes

import (
	"net/http"

	"github.com/meta-boy/mech-alligator/internal/api/handlers"
)

func SetupUserRoutes(mux *http.ServeMux, userHandler *handlers.UserHandler) {
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			userHandler.Login(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
