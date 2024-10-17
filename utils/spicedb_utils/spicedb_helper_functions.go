package spicedb_utils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"
)

var (
	spicedbHost  = getEnv("SPICEDB_HOST", "localhost")
	spicedbPort  = getEnv("SPICEDB_PORT", "50051")
	spicedbToken = getEnv("SPICEDB_TOKEN", "spiceDBKey")
)

// getEnv retrieves the value of the environment variable named by the key.
// It returns the default value if the variable is not set.
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

const (
	ObjectTypeUser       = "user"
	PermissionDelete     = "delete"
	PermissionMakeAdmin  = "make_admin"
	PermissionMakeMember = "make_member"
	ObjectTypeOrg        = "organization"
	relationMember       = "member"
	relationAdmin        = "admin"
	relationBelongsTo    = "belongs_to"
	orgID                = "my_org"
)

func GetSpiceDBClient() *authzed.Client {
	// Load self-signed certificate
	caCert, err := os.ReadFile("/Users/ayushsinghal/server.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %s", err)
	}

	// Create a CA certificate pool and add the self-signed certificate
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatalf("failed to append CA certificate")
	}

	// Create a new TLS configuration
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(tlsConfig)

	spicedbEndpoint := fmt.Sprintf("%s:%s", spicedbHost, spicedbPort)

	// Initialize the client with TLS credentials
	client, err := authzed.NewClient(
		spicedbEndpoint,
		grpc.WithTransportCredentials(creds),
		grpcutil.WithBearerToken(spicedbToken),
	)

	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}

	return client
}

func InitializeSpiceDBSchema(client *authzed.Client, schemaFilePath string) {
	// Read schema from file
	schema, err := os.ReadFile(schemaFilePath)
	if err != nil {
		log.Fatalf("Error reading schema file: %v", err)
	}

	// Write the schema
	request := &pb.WriteSchemaRequest{Schema: string(schema)}
	_, err = client.WriteSchema(context.Background(), request)
	if err != nil {
		log.Fatalf("failed to write schema: %s", err)
	} else {
		log.Println("schema written")
	}
}

func HasPermissionForUserAction(loggedInUserID string, targetUserID string, permission string) bool {
	subject := &pb.SubjectReference{Object: &pb.ObjectReference{

		ObjectType: ObjectTypeUser,
		ObjectId:   loggedInUserID,
	}}

	resource := &pb.ObjectReference{
		ObjectType: ObjectTypeUser,
		ObjectId:   targetUserID,
	}

	ctx := context.Background()

	client := GetSpiceDBClient()

	return CheckPermission(ctx, client, resource, subject, permission)
}

func CheckPermission(ctx context.Context, client *authzed.Client, resourceObj *pb.ObjectReference, subjectSub *pb.SubjectReference, permission string) bool {
	resp, err := client.CheckPermission(ctx, &pb.CheckPermissionRequest{
		Resource:    resourceObj,
		Permission:  permission,
		Subject:     subjectSub,
		WithTracing: true,
	})
	handleError(err, "failed to check permission")
	log.Printf("response for %s attempting to perform '%s' on %s: %s\n", subjectSub.Object.ObjectId, permission, resourceObj.ObjectId, resp)

	if doesHavePermission(resp) {
		log.Printf("%s has permission to '%s' %s", subjectSub.Object.ObjectId, permission, resourceObj.ObjectId)
	} else {
		log.Printf("%s does not have permission to '%s' %s", subjectSub.Object.ObjectId, permission, resourceObj.ObjectId)
	}

	return doesHavePermission(resp)
}

func doesHavePermission(response *pb.CheckPermissionResponse) bool {
	if response == nil {
		return false
	}
	return response.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
}

func handleError(err error, msg string) {
	if err != nil {
		slog.Error(fmt.Sprintf("%s: %s", msg, err))
	}
}

func MakeUserAdmin(userID string) {
	client := GetSpiceDBClient()

	AssignRolesToUser(client, userID, []string{relationMember, relationAdmin}, orgID)
}

func MakeUserMember(userID string) {
	client := GetSpiceDBClient()

	AssignRolesToUser(client, userID, []string{relationMember}, orgID)
}

func AssignRolesToUser(client *authzed.Client, userID string, roles []string, orgID string) {
	ctx := context.Background()

	currentRoles := GetRolesForUser(userID)

	//Step 2: Identify roles to add and roles to remove
	rolesToAdd := difference(roles, currentRoles)    // Roles in newRoles but not in currentRoles
	rolesToRemove := difference(currentRoles, roles) // Roles in currentRoles but not in newRoles

	slog.Info(fmt.Sprintf("Roles to add: %s", rolesToAdd))
	slog.Info(fmt.Sprintf("Roles to remove: %s", rolesToRemove))

	// Step 3: Prepare relationship additionUpdates
	additionUpdates := make([]*pb.RelationshipUpdate, len(rolesToAdd))
	deletionUpdates := make([]*pb.DeleteRelationshipsRequest, len(rolesToRemove))

	// Add roles that are missing
	for i, role := range rolesToAdd {
		additionUpdates[i] = &pb.RelationshipUpdate{
			Operation: pb.RelationshipUpdate_OPERATION_TOUCH, // Use TOUCH to create or update the relationship
			Relationship: &pb.Relationship{
				Resource: &pb.ObjectReference{
					ObjectType: ObjectTypeOrg, // Assuming the context of an organization
					ObjectId:   orgID,         // Organization ID
				},
				Relation: role, // Role being assigned (e.g., "admin", "member")
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: ObjectTypeUser,
						ObjectId:   userID, // User ID to whom the role is being assigned
					},
				},
			},
		}
	}

	// Remove roles that are no longer needed
	for i, role := range rolesToRemove {
		deletionUpdates[i] = &pb.DeleteRelationshipsRequest{
			RelationshipFilter: &pb.RelationshipFilter{
				ResourceType:     ObjectTypeOrg, // Assuming roles are defined under an organization
				OptionalRelation: role,          // The role (relation) to be removed
				OptionalSubjectFilter: &pb.SubjectFilter{
					SubjectType:       ObjectTypeUser, // User object
					OptionalSubjectId: userID,         // User ID
				},
			},
			// OptionalPreconditions and OptionalLimit can be nil or 0 if not needed
			OptionalAllowPartialDeletions: false, // Set to false to avoid partial deletions
		}
	}

	//Step 4: Apply batch additionUpdates to SpiceDB
	if len(additionUpdates) > 0 {
		_, err := client.WriteRelationships(ctx, &pb.WriteRelationshipsRequest{
			Updates: additionUpdates,
		})
		if err != nil {
			slog.Error(fmt.Sprintf("failed to update roles for user %s: %v", userID, err))
		}
	}

	// Apply the deletions
	for _, deleteReq := range deletionUpdates {
		_, err := client.DeleteRelationships(ctx, deleteReq)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to delete roles for user %s: %v", userID, err))
		}
	}
}

func difference(slice1, slice2 []string) []string {
	diff := []string{}
	lookup := make(map[string]struct{}, len(slice2))
	for _, item := range slice2 {
		lookup[item] = struct{}{}
	}

	for _, item := range slice1 {
		if _, found := lookup[item]; !found {
			diff = append(diff, item)
		}
	}

	return diff
}

func GetRolesForUser(userID string) []string {
	client := GetSpiceDBClient()

	ctx := context.Background()

	// Construct the subject filter to filter by user
	subjectFilter := &pb.SubjectFilter{
		SubjectType:       ObjectTypeUser, // The type of the user object
		OptionalSubjectId: userID,         // The ID of the user whose roles we are fetching
	}

	// Construct the request to read relationships
	request := &pb.ReadRelationshipsRequest{
		RelationshipFilter: &pb.RelationshipFilter{
			ResourceType:          ObjectTypeOrg, // Assuming roles are defined under an organization
			OptionalSubjectFilter: subjectFilter, // Filter for the specific user
		},
	}

	// Attempt to read the relationships
	stream, err := client.ReadRelationships(ctx, request)
	if err != nil {
		log.Printf("failed to read roles for user %s: %s", userID, err)
		return nil
	}

	roles := []string{}

	// Process the stream of results
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break // End of the stream
		}
		if err != nil {
			log.Printf("error receiving stream response for user %s: %s", userID, err)
			return nil
		}

		// Extract the role (relation) and add it to the roles slice
		roles = append(roles, resp.Relationship.Relation)
	}

	return roles
}

func SaveUserToSpiceDB(userID string, roles []string) {
	ctx := context.Background()
	client := GetSpiceDBClient()

	updates := []*pb.RelationshipUpdate{}

	// Iterate over the roles and create relationship updates
	for _, role := range roles {
		// Create the relationship update for this role
		roleUpdate := &pb.RelationshipUpdate{
			Operation: pb.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &pb.Relationship{
				Resource: &pb.ObjectReference{
					ObjectType: ObjectTypeOrg, // Assuming the context of an organization
					ObjectId:   orgID,
				},
				Relation: role, // Role being assigned (e.g., "admin", "member")
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: ObjectTypeUser,
						ObjectId:   userID,
					},
				},
			},
		}
		updates = append(updates, roleUpdate)
	}

	// Adding user to organization
	orgUpdate := &pb.RelationshipUpdate{
		Operation: pb.RelationshipUpdate_OPERATION_CREATE,
		Relationship: &pb.Relationship{
			Resource: &pb.ObjectReference{
				ObjectType: ObjectTypeUser,
				ObjectId:   userID,
			},
			Relation: relationBelongsTo,
			Subject: &pb.SubjectReference{
				Object: &pb.ObjectReference{
					ObjectType: ObjectTypeOrg,
					ObjectId:   orgID,
				},
			},
		},
	}
	updates = append(updates, orgUpdate)

	// Construct the request with all updates
	request := &pb.WriteRelationshipsRequest{Updates: updates}

	// Attempt to write the relationships
	resp, err := client.WriteRelationships(ctx, request)
	if err != nil {
		log.Printf("failed to save roles for user %s in organization %s: %s", userID, orgID, err)
	} else {
		log.Printf("Roles %v and 'belongs_to' relationship assigned to user %s in organization %s with token: %s\n", roles, userID, orgID, resp.WrittenAt.Token)
	}
}

func SaveUserToSpiceDBUsingGoRoutines(userID string, roles []string) {
	ctx := context.Background()
	client := GetSpiceDBClient()

	var wg sync.WaitGroup
	roleChannel := make(chan *pb.RelationshipUpdate)

	// Start a goroutine to handle writing relationships to SpiceDB
	go func() {
		for update := range roleChannel {
			// Write relationship to SpiceDB
			request := &pb.WriteRelationshipsRequest{
				Updates: []*pb.RelationshipUpdate{update},
			}
			_, err := client.WriteRelationships(ctx, request)
			if err != nil {
				log.Printf("failed to save role '%s' for user '%s' in organization '%s': %s", update.Relationship.Relation, userID, orgID, err)
			} else {
				log.Printf("saved role '%s' for user '%s' in organization '%s'", update.Relationship.Relation, userID, orgID)
			}
		}
	}()

	// Iterate over the roles and send relationship updates to the channel
	for _, role := range roles {
		wg.Add(1)
		go func(role string) {
			defer wg.Done()

			// Create the relationship update for this role
			roleUpdate := &pb.RelationshipUpdate{
				Operation: pb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &pb.Relationship{
					Resource: &pb.ObjectReference{
						ObjectType: ObjectTypeOrg, // Adjust based on your resource type
						ObjectId:   orgID,
					},
					Relation: role,
					Subject: &pb.SubjectReference{
						Object: &pb.ObjectReference{
							ObjectType: ObjectTypeUser, // Adjust based on your subject type
							ObjectId:   userID,
						},
					},
				},
			}
			roleChannel <- roleUpdate // Send the update to the channel
		}(role)
	}

	// Adding user to organization
	wg.Add(1)
	go func() {
		defer wg.Done()
		orgUpdate := &pb.RelationshipUpdate{
			Operation: pb.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &pb.Relationship{
				Resource: &pb.ObjectReference{
					ObjectType: ObjectTypeUser,
					ObjectId:   userID,
				},
				Relation: relationBelongsTo,
				Subject: &pb.SubjectReference{
					Object: &pb.ObjectReference{
						ObjectType: ObjectTypeOrg,
						ObjectId:   orgID,
					},
				},
			},
		}

		roleChannel <- orgUpdate
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(roleChannel) // Close the channel to signal the writing goroutine to stop
}

func SaveUserInSpiceDBWithDefaults(userID string) {
	roles := []string{relationMember}
	SaveUserToSpiceDBUsingGoRoutines(userID, roles)
}

func DeleteUserFromSpiceDB(userID string) {
	ctx := context.Background()
	client := GetSpiceDBClient()

	// Step 1: Delete relationships where the user is the subject
	deleteRelationshipsRequest := &pb.DeleteRelationshipsRequest{
		RelationshipFilter: &pb.RelationshipFilter{
			OptionalSubjectFilter: &pb.SubjectFilter{
				SubjectType:       "user",
				OptionalSubjectId: userID,
			},
		},
	}

	// Execute the delete relationships request
	response, err := client.DeleteRelationships(ctx, deleteRelationshipsRequest)
	if err != nil {
		log.Printf("Failed to delete relationships for user %s: %s with response %s", userID, err, response)
		return
	}
	log.Printf("Successfully deleted relationships for user %s with response %s", userID, response)

	// Step 2: Delete the user object itself
	deleteUserObjectRequest := &pb.DeleteRelationshipsRequest{
		RelationshipFilter: &pb.RelationshipFilter{
			ResourceType:       "user",
			OptionalResourceId: userID,
		},
	}

	// Execute the delete user object request
	response, err = client.DeleteRelationships(ctx, deleteUserObjectRequest)
	if err != nil {
		log.Printf("Failed to delete user object %s: %s with response %s", userID, err, response)
	} else {
		log.Printf("Successfully deleted user object %s with response %s", userID, response)
	}
}
