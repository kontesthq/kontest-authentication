package admin_handler

import (
	"encoding/json"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"kontest-authentication/utils/spicedb_utils"
	"net/http"
)

func MakeMemberHandler(w http.ResponseWriter, r *http.Request) {
	var updateUserRoleRequest model.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserRoleRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	loggedInUserUID, err := GetLoggedInUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	uid := updateUserRoleRequest.UID

	isUserAllowed := spicedb_utils.HasPermissionForUserAction(loggedInUserUID.String(), updateUserRoleRequest.UID.String(), spicedb_utils.PermissionMakeMember)
	if !isUserAllowed {
		http.Error(w, "Error: admin only endpoint", http.StatusUnauthorized)
		return
	}

	userService := service.NewUserService()

	hasMade, err := userService.MakeUserMember(uid)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !hasMade {
		http.Error(w, "not able to make user member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Made user member")
}
