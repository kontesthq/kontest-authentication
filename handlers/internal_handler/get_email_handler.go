package internal_handler

import (
	"encoding/json"
	"kontest-authentication/service"
	"kontest-authentication/utils"
	"net/http"
)

func GetEmail(w http.ResponseWriter, r *http.Request) {
	if !checkIsRequestInternal(r) {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		http.Error(w, "uuid not present", http.StatusBadRequest)
		return
	}

	uid, err := utils.IsValidUUID(userID)

	if err != nil {
		http.Error(w, "uuid not valid", http.StatusBadRequest)
		return
	}

	userService := service.NewUserService()

	user, err := userService.GetUserByUserID(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.Email)
}

func checkIsRequestInternal(r *http.Request) bool {
	return true
}
