package routes

import (
	"kontest-authentication/handlers"
	"net/http"
)

func RegisterRoutes(router *http.ServeMux) {
	RegisterUserRoutes(router)
}

func RegisterUserRoutes(router *http.ServeMux) {
	router.HandleFunc("PUT /user/login", handlers.LoginHandler)
	router.HandleFunc("POST /user/logout", handlers.LogoutHandler)
	router.HandleFunc("POST /user/register", handlers.RegisterHandler)
	router.HandleFunc("PUT /user/refresh", handlers.RefreshHandler)
}
