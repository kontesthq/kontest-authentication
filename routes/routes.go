package routes

import (
	"kontest-authentication/handlers"
	"net/http"
)

func RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/login", handlers.LoginHandler)
	router.HandleFunc("/logout", handlers.LogoutHandler)
	router.HandleFunc("/register", handlers.RegisterHandler)
}
