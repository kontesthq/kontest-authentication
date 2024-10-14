package routes

import (
	"kontest-authentication/handlers"
	"net/http"
)

func RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("PUT /login", handlers.LoginHandler)
	router.HandleFunc("POST /logout", handlers.LogoutHandler)
	router.HandleFunc("POST /register", handlers.RegisterHandler)
}
