package main

import (
	"fmt"
	"github.com/google/uuid"
	consulServiceManager "github.com/kontesthq/go-consul-service-manager/consulservicemanager"
	"kontest-authentication/database"
	"kontest-authentication/model"
	"kontest-authentication/routes"
	"kontest-authentication/service"
	"kontest-authentication/utils/kafka_utils"
	"kontest-authentication/utils/spicedb_utils"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	applicationHost = "localhost"                      // Default value for local development
	applicationPort = 5155                             // Default value for local development
	serviceName     = "KONTEST-AUTHENTICATION-SERVICE" // Service name for Service Registry
	consulHost      = "localhost"                      // Default value for local development
	consulPort      = 5150

	// DB properties
	dbHost           = "localhost"
	dbPort           = "5432"
	dbName           = "kontest"
	dbUser           = "ayushsinghal"
	dbPassword       = ""
	isSSLModeEnabled = false
)

func initializeVariables() {
	// Get the hostname of the machine
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error fetching hostname: %v", err)
	}

	// Attempt to read the KONTEST_API_SERVER_HOST environment variable
	if host := os.Getenv("KONTEST_AUTHENTICATION_SERVICE_HOST"); host != "" {
		applicationHost = host // Override with the environment variable if set
	} else {
		applicationHost = hostname // Use the machine's hostname if the env var is not set
	}

	// Attempt to read the KONTEST_API_SERVER_PORT environment variable
	if port := os.Getenv("KONTEST_AUTHENTICATION_SERVICE_PORT"); port != "" {
		parsedPort, err := strconv.Atoi(port)
		if err != nil {
			log.Fatalf("Invalid port value: %v", err)
		}
		applicationPort = parsedPort // Override with the environment variable if set
	}

	// Attempt to read the CONSUL_ADDRESS environment variable
	if host := os.Getenv("CONSUL_HOST"); host != "" {
		consulHost = host // Override with the environment variable if set
	}

	// Attempt to read the CONSUL_PORT environment variable
	if port := os.Getenv("CONSUL_PORT"); port != "" {
		if portInt, err := strconv.Atoi(port); err == nil {
			consulPort = portInt // Override with the environment variable if set and valid
		}
	}

	// Attempt to read the DB_HOST environment variable
	if host := os.Getenv("DB_HOST"); host != "" {
		dbHost = host // Override with the environment variable if set
	}

	// Attempt to read the DB_PORT environment variable
	if port := os.Getenv("DB_PORT"); port != "" {
		dbPort = port // Override with the environment variable if set
	}

	// Attempt to read the DB_NAME environment variable
	if name := os.Getenv("DB_NAME"); name != "" {
		dbName = name // Override with the environment variable if set
	}

	// Attempt to read the DB_USER environment variable
	if user := os.Getenv("DB_USER"); user != "" {
		dbUser = user // Override with the environment variable if set
	}

	// Attempt to read the DB_PASSWORD environment variable
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		dbPassword = password // Override with the environment variable if set
	}

	// Attempt to read the DB_SSL_MODE environment variable
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		isSSLModeEnabled = sslMode == "enable"
	}
}

func main() {
	initializeVariables()

	kafkaConfig := kafka_utils.GetKafkaConfig()

	kafkaBroker := kafkaConfig.KafkaHost + ":" + kafkaConfig.KafkaPort
	service.InitKafka(kafkaBroker)

	consulService := consulServiceManager.NewConsulService(consulHost, consulPort)
	consulService.Start(applicationHost, applicationPort, serviceName, []string{})

	// Initialize the database connection
	database.InitializeDatabase(dbName, dbPort, dbHost, dbUser, dbPassword, map[bool]string{true: "enable", false: "disable"}[isSSLModeEnabled])
	database.SetupDatabase()
	defer database.CloseDB()

	DoStartupTasks()

	router := http.NewServeMux()

	routes.RegisterRoutes(router)

	server := http.Server{
		Addr:    ":" + strconv.Itoa(applicationPort),
		Handler: router,
	}

	fmt.Println("Server listening at applicationPort: " + strconv.Itoa(applicationPort))

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func DoStartupTasks() {
	// Make users admin
	MakeUsersAdmin()
}

func MakeUsersAdmin() {
	tx, err := database.GetDB().Beginx()

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to begin transaction: %v", err))
		os.Exit(1)
	}

	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				slog.Warn("Cannot rollback transaction")
				return
			} // Rollback if there was an error
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				slog.Error(fmt.Sprintf("Failed to commit transaction: %v", commitErr))
				os.Exit(1)
			}
		}
	}()

	emailsOfUsersToMakeAdmin := []string{"ayushsinghals02@gmail.com"}

	for _, email := range emailsOfUsersToMakeAdmin {
		user, err := database.FindUserByEmail(email)
		if err != nil {
			slog.Error(fmt.Sprintf("Error finding user with email %s: %v\n", email, err))
			continue
		}

		// Assign the admin role to the user in DB
		_, err = database.AssignRoleToUser(user.ID, model.GetRoleAdmin().ID, tx)
		if err != nil {
			slog.Error(fmt.Sprintf("Error assigning admin role to user with email %s: %v\n", email, err))
		}

		// Assign the admin role to the user in spiceDB
		spicedb_utils.MakeUserAdmin(user.ID.String())
	}

}

func doDatabaseTest() {
	tx, err := database.GetDB().Beginx()

	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				log.Println("Cannot rollback transaction")
				return
			} // Rollback if there was an error
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				log.Fatalf("Failed to commit transaction: %v", commitErr)
			}
		}
	}()

	// Inserting sample users into the database
	users := []model.User{
		{ID: uuid.New(), Email: "email@s.com", Password: "hashed_password"},
		{ID: uuid.New(), Email: "user2@example.com", Password: "hashed_password"},
	}

	// Use the db variable to create multiple users
	for _, user := range users {
		refreshToken := model.RefreshToken{
			TokenID:      uuid.New(),
			RefreshToken: "sample_refresh_token",         // generate or pass a token here
			Expiry:       time.Now().Add(24 * time.Hour), // set your expiration time
			UserID:       user.ID,
		}

		// Insert users
		_, err := database.InsertUserIntoDB(user, tx)
		if err != nil {
			log.Printf("Error adding user %s: %v\n", user.Email, err)
			continue
		}

		// Insert refresh token
		_, err = database.InsertRefreshTokenIntoDB(refreshToken, tx)

		// Add a device for the refresh token
		device := model.Device{
			RefreshTokenID: refreshToken.TokenID,
		}

		_, err = database.InsertDeviceIntoDB(device, tx)
	}

	log.Println("Users added successfully")

	// Inserting sample roles into the database
	roles := []model.Role{
		{ID: 1, Name: "user"},
		{ID: 2, Name: "admin"},
	}

	// Use the db variable to create multiple roles
	for _, role := range roles {
		if _, err := database.InsertRoleIntoDB(role, tx); err != nil {
			log.Printf("Error adding role %s: %v\n", role.Name, err)
		}
	}

	log.Println("Roles added successfully")

	// Inserting role assignments (assign roles to users) into the user_roles table
	userRoles := []struct {
		UserID uuid.UUID `db:"user_id"`
		RoleID int       `db:"role_id"`
	}{
		{UserID: users[0].ID, RoleID: 1}, // Assign "user" role to first user
		{UserID: users[0].ID, RoleID: 2}, // Assign "admin" role to first user
		{UserID: users[1].ID, RoleID: 1}, // Assign "user" role to second user
	}

	// Insert user roles
	for _, userRole := range userRoles {
		if _, err := database.AssignRoleToUser(userRole.UserID, userRole.RoleID, tx); err != nil {
			log.Printf("Error assigning role ID %d to user ID %s: %v\n", userRole.RoleID, userRole.UserID, err)
		}
	}
	log.Println("User roles added successfully")
}
