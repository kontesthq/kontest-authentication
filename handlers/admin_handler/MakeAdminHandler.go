package admin_handler

import (
	"encoding/json"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"net/http"
)

func MakeAdminHandler(w http.ResponseWriter, r *http.Request) {
	var updateUserRoleRequest model.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserRoleRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	uid := updateUserRoleRequest.UID

	userService := service.NewUserService()

	hasMade, err := userService.MakeUserAdmin(uid)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !hasMade {
		http.Error(w, "not able to make user admin", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Made user admin")
}
