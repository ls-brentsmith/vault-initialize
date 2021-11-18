package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/hashicorp/vault/api"
	fromenv "github.com/ls-brentsmith/vault-initialize/pkg/fromenv"
	k8s "github.com/ls-brentsmith/vault-initialize/pkg/k8s"
	secrets "github.com/ls-brentsmith/vault-initialize/pkg/secrets"
)

var (
	VaultClient  *api.Client
	InitResponse *api.InitResponse
	UserAgent    = fmt.Sprintf("vault-initialize/0.1.0 (%s)", runtime.Version())
	SecretID     = fromenv.Required("SECRET_ID")
	ProjectID    = fromenv.Required("PROJECT_ID")
)

func main() {
	// initalize vault
	Init()

	json, err := (json.Marshal(InitResponse))
	if err != nil {
		log.Fatalf("Unable to unmarshall response: %v", err)
	}

	// Upload root token and recovery keys
	secrets.CreateSecret(SecretID, ProjectID, json)

	k8s.CreateK8sSecret()
}

func initClient() {
	// Instantiate a new client
	var err error
	VaultClient, err = api.NewClient(&api.Config{
		Address: fromenv.String("VAULT_ADDR", "http://127.0.0.1:8200"),
	})

	VaultClient.AddHeader("User-Agent", UserAgent)

	if err != nil {
		log.Fatalf("Could not instantiate vault client: %v", err)
	}
}

func Init() {
	// Instantiate a new client
	initClient()

	// Poll for vault status
	for {
		resp, err := status()
		if err != nil {
			log.Println("Vault is unreachable, retrying.")
		} else if resp.Initialized {
			if InitResponse == nil {
				log.Println("Vault already initialized. Nothing to do.")
				// Possible we should exit with some other code to indicate no-op?
				os.Exit(0)
			} else {
				log.Println("Initialized and Unsealed.")
				break
			}
		} else if resp.Sealed {
			fmt.Println("Sealed and Unitialized. Initializing!")
			InitResponse = initialize()
			unseal()
			continue
		} else {
			// We shouldn't be able to hit this based on the current
			// implementation of api.StatusReponse. Print debug and retry.
			log.Printf("Unknown init response. Retrying. (%v)", resp)
		}
		wait()
	}

}

func wait() {
	// Capture SIGINT and SIGTERM
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Shutdown gracefully if signal received
	stop := func() {
		log.Printf("Shutting down")
		// Do stuff
		os.Exit(0)
	}

	log.Printf("Next check in %v", 10*time.Second)

	checkInterval := fromenv.Duration("CHECK_INTERVAL", 10*time.Second)
	select {
	case <-signalCh:
		// Stop if signal received
		stop()
	case <-time.After(checkInterval):
		// Wait for checkInterval
	}
}

func unseal() {
	if len(InitResponse.RecoveryKeys) > 0 {
		// if initResponse has recovery keys, it means the server is set to auto-unseal
		log.Println("Auto-unseal enabled on the server.")
	} else if len(InitResponse.Keys) > 0 {
		// if initResponse has Keys, it means that server needs to be unsealed
		log.Println("Detected unseal keys, unsealing.")
		// TODO: iterate over the array of keys for unseal. Right now hardcoded to just 1, but that could be configurable
		_, err := VaultClient.Sys().Unseal(InitResponse.Keys[0])
		if err != nil {
			log.Fatalf("Unable to unseal vault %v", err)
		}
		log.Println("Successfully Unsealed.")
	}
}

func initialize() (initResponse *api.InitResponse) {
	// TODO: make these values configurable
	initRequest := api.InitRequest{
		SecretShares:      1,
		SecretThreshold:   1,
		RecoveryShares:    1,
		RecoveryThreshold: 1,
	}
	initResponse, err := VaultClient.Sys().Init(&initRequest)
	if err != nil {
		log.Fatalf("Failed to initalize: %v", err)
	}

	return
}

func status() (*api.HealthResponse, error) {
	resp, err := VaultClient.Sys().Health()
	if err != nil {
		// TODO: implement log levels
		log.Printf("Vault is unreachable (%v)", err)
		return nil, err
	}
	return resp, nil
}
