package admin_handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"kontest-authentication/utils"
	"log/slog"
	"net/http"
)

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// ID of user to be deleted
	var deleteUserRequest model.DeleteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&deleteUserRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if user is logged in or not
	loggedInUserID := r.Header.Get(utils.UserIdRequestHeader)

	if loggedInUserID == "" {
		http.Error(w, "Error: admin only endpoint", http.StatusUnauthorized)
		return
	}

	uid, err := utils.IsValidUUID(loggedInUserID)

	if err != nil {
		http.Error(w, "Error: admin only endpoint", http.StatusUnauthorized)
		return
	}

	isUserAllowed := checkIfUserIsAdmin(uid)

	if !isUserAllowed {
		http.Error(w, "Error: admin only endpoint", http.StatusUnauthorized)
		return
	}

	userService := service.NewUserService()

	isDeleted, err := userService.DeleteUser(deleteUserRequest.UID)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if !isDeleted {
		slog.Error("Unable to delete user")
		http.Error(w, "Error: Unable to delete user", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}

func checkIfUserIsAdmin(userID uuid.UUID) bool {
	return true
}
