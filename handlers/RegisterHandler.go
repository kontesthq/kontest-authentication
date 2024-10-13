package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	error2 "kontest-authentication/error"
	"kontest-authentication/model"
	"kontest-authentication/service"
	"log"
	"net/http"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	userService := service.NewUserService()

	// Extract user body from request
	user := model.User{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Please provide user in correct JSON format", http.StatusBadRequest)
		return
	}

	// Check if Email and Password are provided
	if user.Email == "" || user.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	uid, err := userService.Register(user)
	if err != nil {
		// Declare a variable for error type checking
		if errors.Is(err, &error2.UserAlreadyPresentError{}) {
			http.Error(w, "User already exists", http.StatusConflict) // 409 Conflict
			return
		}

		// Handle other errors
		log.Printf("Error in registering user: %s", err)
		http.Error(w, fmt.Sprintf("Error in registering user: %s", err), http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	// Log the message
	log.Printf("User registered with email: %s, and ID %s", user.Email, uid)

	// Set the response status to 201 Created
	w.WriteHeader(http.StatusCreated)

	// Send a success message to the client
	fmt.Fprintf(w, "User registered successfully!")
}
