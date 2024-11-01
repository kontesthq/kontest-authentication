package user_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	error2 "github.com/kontesthq/go-security/Auth/ott/error"
	error3 "kontest-authentication/error"
	"kontest-authentication/model"
	"kontest-authentication/service/ott"
	"log/slog"
	"net/http"
)

func OTTLoginGenerateHandler(w http.ResponseWriter, r *http.Request) {
	jwtRequest := model.OTTGenerateRequest{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&jwtRequest); err != nil {
		http.Error(w, "Please provide Generate Login OTT Request body in correct form.", http.StatusBadRequest)
		return
	}

	if jwtRequest.Username == "" {
		http.Error(w, "username cannot be empty", http.StatusBadRequest)
		return
	}

	ottService := ott.NewOTTService()

	err := ottService.HandleLoginOTT(jwtRequest.Username)
	if err != nil {
		if errors.Is(err, &error3.UserNotFoundError{}) {
			slog.Error(fmt.Sprintf("User not found: %s\n", jwtRequest.Username))
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			slog.Error(fmt.Sprintf("Error: %v\n", err))
			http.Error(w, "Internal Server Error occurred", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("One Time Token generated successfully")) // Send JSON response
}

func ValidateLoginOTT(w http.ResponseWriter, r *http.Request) {
	jwtRequest := model.LoginOTTValidateRequest{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&jwtRequest); err != nil {
		http.Error(w, "Please provide Validate Login OTT Request body in correct form.", http.StatusBadRequest)
		return
	}

	if jwtRequest.OTT == "" {
		http.Error(w, "one time token cannot be empty", http.StatusBadRequest)
		return
	}

	ottService := ott.NewOTTService()

	jwtToken, refreshToken, err := ottService.ValidateLoginOTT(jwtRequest.OTT)
	if err != nil {
		if errors.Is(err, &error2.InvalidOneTimeTokenError{}) {
			slog.Error(fmt.Sprintf("Invalid OTT: %s\n", jwtRequest.OTT))
			http.Error(w, "Invalid OTT", http.StatusUnauthorized)
		} else {
			slog.Error(fmt.Sprintf("Error in validating login ott for ott %s, Error: %v\n", jwtRequest.OTT, err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Successful authentication response
	response := model.JWTResponse{
		JWTToken:     jwtToken,
		RefreshToken: refreshToken,
		Username:     "",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response) // Send JSON response
}
