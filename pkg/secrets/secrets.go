// Package secrets implements Google Secret Manager interactions
package secrets

import (
	"context"
	"fmt"
	"log"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Create new Google Secrets Manager Secret
func createSecret(ctx context.Context, client secretmanager.Client, secretName string) (*secretmanagerpb.Secret, error) {
	// Sprint secretName `projects/*/secrets/*` into projectId and secretId
	s := strings.Split(secretName, "/")
	projectId, secretId := s[1], s[3]

	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", projectId),
		SecretId: secretId,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	secret, err := client.CreateSecret(ctx, createSecretReq)
	if err != nil {
		// Unhandled Error
		log.Fatalf("failed to create secret: %v", err)
	}

	return secret, nil
}

func createSecretVersion(ctx context.Context, client secretmanager.Client, secretName string, payload []byte) {
	// Build the request.
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: payload,
		},
	}

	// Call the API
	version, err := client.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		log.Fatalf("failed to add secret version: %v", err)
	}

	log.Printf("Added secret version: %v", version.Name)
}

// Create a new secret unless it already exists.
func CreateSecret(secretId string, projectId string, payload []byte) {
	// The resource name of the Secret, in the format `projects/*/secrets/*`.
	secretName := fmt.Sprintf("projects/%s/secrets/%s", projectId, secretId)

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}
	defer client.Close()

	req := &secretmanagerpb.GetSecretRequest{
		Name: secretName,
	}

	// Describe secret to see if it already exists
	_, err = client.GetSecret(ctx, req)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Secret not found, attempt to create
			log.Printf("Secret does not exist, creating: %v", secretName)

			_, err = createSecret(ctx, *client, secretName)

			if err != nil {
				// Unable to create secret. Panic.
				log.Fatalf("Unable to create secret %v. %v", secretName, err)
			}
		} else {
			// Unable to describe secret. Panic.
			log.Fatalf("Unable to describe secret: %v. %v", secretName, err)
		}
	}

	createSecretVersion(ctx, *client, secretName, payload)
}
