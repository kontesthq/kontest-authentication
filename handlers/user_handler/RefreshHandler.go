package user_handler

import (
	"encoding/json"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"net/http"
)

func RefreshHandler(writer http.ResponseWriter, request *http.Request) {
	refreshTokenRequest := model.RefreshTokenRequest{}
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()

	if err := decoder.Decode(&refreshTokenRequest); err != nil {
		http.Error(writer, "Please provide refresh token request in correct JSON format", http.StatusBadRequest)
		return
	}

	if refreshTokenRequest.RefreshToken == "" {
		http.Error(writer, "Refresh Token can not be empty", http.StatusBadRequest)
		return
	}

	refreshTokenService := service.NewRefreshTokenService()

	// Call the new Refresh method
	response, err := refreshTokenService.Refresh(refreshTokenRequest.RefreshToken)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(response) // Send JSON response
}
