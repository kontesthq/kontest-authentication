package admin_handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"kontest-authentication/utils"
	"kontest-authentication/utils/spicedb_utils"
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

	loggedInUserUID, err := GetLoggedInUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	slog.Info(fmt.Sprintf("loggedInUserUID: %s/n", loggedInUserUID))

	isUserAllowed := spicedb_utils.HasPermissionForUserAction(loggedInUserUID.String(), deleteUserRequest.UID.String(), spicedb_utils.PermissionDelete)
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

func GetLoggedInUserID(r *http.Request) (uuid.UUID, error) {
	// Check if user is logged in or not
	loggedInUserID := r.Header.Get(utils.UserIdRequestHeader)
	if loggedInUserID == "" {
		return uuid.Nil, &error2.AdminOnlyEndpointError{}
	}

	loggedInUserUID, err := utils.IsValidUUID(loggedInUserID)
	if err != nil {
		return uuid.Nil, &error2.AdminOnlyEndpointError{}
	}

	return loggedInUserUID, nil
}
