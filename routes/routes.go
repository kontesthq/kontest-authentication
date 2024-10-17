package routes

import (
	"kontest-authentication/handlers/admin_handler"
	"kontest-authentication/handlers/user_handler"
	"net/http"
)

func RegisterRoutes(router *http.ServeMux) {
	registerUserRoutes(router)
	registerAdminRoutes(router)
}

func registerUserRoutes(router *http.ServeMux) {
	router.HandleFunc("PUT /user/login", user_handler.LoginHandler)
	router.HandleFunc("POST /user/logout", user_handler.LogoutHandler)
	router.HandleFunc("POST /user/register", user_handler.RegisterHandler)
	router.HandleFunc("PUT /user/refresh", user_handler.RefreshHandler)
}

func registerAdminRoutes(router *http.ServeMux) {
	router.HandleFunc("PUT /admin/makeAdmin", admin_handler.MakeAdminHandler)
	router.HandleFunc("PUT /admin/makeNormal", admin_handler.MakeNormalHandler)
	router.HandleFunc("DELETE /admin/deleteUser", admin_handler.DeleteUserHandler)
}
