package handlers

import (
	"encoding/json"
	"fmt"
	"kontest-authentication/Auth"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	jwtRequest := model.JWTRequest{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&jwtRequest); err != nil {
		http.Error(w, "Please provide login request in correct JSON format", http.StatusBadRequest)
		return
	}

	if jwtRequest.Email == "" || jwtRequest.Password == "" {
		http.Error(w, "Email or Password cannot be empty", http.StatusBadRequest)
		return
	}

	us := service.NewUserService()

	if !us.IsValidDeviceID(jwtRequest.DeviceID) {
		http.Error(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	uid, err := us.DoAuthenticate(jwtRequest.Email, jwtRequest.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusBadRequest)
		return
	}

	fmt.Println(uid)

	jwtToken, err := Auth.GenerateJWTOnly(uid.String(), []byte(Auth.JWTSecret), Auth.JWTTokenExpiryDuration)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
	}

	refreshTokenService := service.NewRefreshTokenService()

	refreshToken, err := refreshTokenService.CreateRefreshToken(uid, jwtRequest.DeviceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}

	// Successful authentication response
	response := model.JWTResponse{
		JWTToken:     jwtToken,
		RefreshToken: refreshToken.RefreshToken,
		Username:     jwtRequest.Email,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response) // Send JSON response
}
