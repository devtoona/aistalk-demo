package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"voice-chat-api-go/logger"
)

var (
	initOnce   sync.Once
	initErr    error
	authClient *auth.Client
)

// Disabled returns true when AUTH_DISABLED is set (local development).
func Disabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_DISABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func ProjectID() string {
	return strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID"))
}

func ensureClient(ctx context.Context) (*auth.Client, error) {
	initOnce.Do(func() {
		projectID := ProjectID()
		if projectID == "" {
			initErr = fmt.Errorf("FIREBASE_PROJECT_ID is required when auth is enabled")
			return
		}
		conf := &firebase.Config{ProjectID: projectID}
		var opts []option.ClientOption
		if cred := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); cred != "" {
			opts = append(opts, option.WithCredentialsFile(cred))
		}
		app, err := firebase.NewApp(ctx, conf, opts...)
		if err != nil {
			initErr = fmt.Errorf("firebase.NewApp: %w", err)
			return
		}
		client, err := app.Auth(ctx)
		if err != nil {
			initErr = fmt.Errorf("firebase Auth: %w", err)
			return
		}
		authClient = client
		logger.Info("Firebase Auth client initialized (project=%s)", projectID)
	})
	if initErr != nil {
		return nil, initErr
	}
	return authClient, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the UID.
func VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	client, err := ensureClient(ctx)
	if err != nil {
		return "", err
	}
	tok, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	if tok.UID == "" {
		return "", fmt.Errorf("empty uid")
	}
	return tok.UID, nil
}
