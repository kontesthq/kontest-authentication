package admin_handler

import (
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
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

	// Check if user is logged in or not
	loggedInUserID := r.Header.Get(utils.UserIdRequestHeader)

	if loggedInUserID == "" {
		http.Error(w, "Error: admin only endpoint", http.StatusUnauthorized)
		return
	}

	loggedInUserUID, err := utils.IsValidUUID(loggedInUserID)
	slog.Info(fmt.Sprintf("loggedInUserUID: %s/n", loggedInUserUID))

	if err != nil {
		http.Error(w, "Error: admin only endpoint", http.StatusUnauthorized)
		return
	}

	isUserAllowed := canDo(loggedInUserUID.String(), deleteUserRequest.UID.String())

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

// ********************************************************************************************************************* //

func canDo(userIDOfLoggedInUser string, userIDOfRequestedUserToBeDelete string) bool {
	subject := &pb.SubjectReference{Object: &pb.ObjectReference{

		ObjectType: spicedb_utils.ObjectTypeUser,
		ObjectId:   userIDOfLoggedInUser,
	}}

	resource := &pb.ObjectReference{
		ObjectType: spicedb_utils.ObjectTypeUser,
		ObjectId:   userIDOfRequestedUserToBeDelete,
	}

	permission := spicedb_utils.PermissionDelete

	ctx := context.Background()

	client := spicedb_utils.GetSpiceDBClient()

	return spicedb_utils.CheckPermission(ctx, client, resource, subject, permission)
}
