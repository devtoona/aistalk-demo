package quota

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"voice-chat-api-go/logger"
)

type Kind string

const (
	KindChat Kind = "chat"
	KindTTS  Kind = "tts"
)

type doc struct {
	DailyDate string    `firestore:"dailyDate"`
	ChatCount int       `firestore:"chatCount"`
	TtsCount  int       `firestore:"ttsCount"`
	UpdatedAt time.Time `firestore:"updatedAt"`
}

var (
	initOnce sync.Once
	initErr  error
	fsClient *firestore.Client
)

func Disabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("QUOTA_DISABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func dailyLimit(kind Kind) int {
	switch kind {
	case KindChat:
		return envInt("QUOTA_DAILY_CHAT_LIMIT", 30)
	case KindTTS:
		return envInt("QUOTA_DAILY_TTS_LIMIT", 30)
	default:
		return 0
	}
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func ensureClient(ctx context.Context) (*firestore.Client, error) {
	initOnce.Do(func() {
		projectID := strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID"))
		if projectID == "" {
			initErr = fmt.Errorf("FIREBASE_PROJECT_ID is required for quota")
			return
		}
		conf := &firebase.Config{ProjectID: projectID}
		var opts []option.ClientOption
		if cred := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); cred != "" {
			opts = append(opts, option.WithCredentialsFile(cred))
		}
		app, err := firebase.NewApp(ctx, conf, opts...)
		if err != nil {
			initErr = fmt.Errorf("firebase.NewApp (quota): %w", err)
			return
		}
		client, err := app.Firestore(ctx)
		if err != nil {
			initErr = fmt.Errorf("firestore: %w", err)
			return
		}
		fsClient = client
		logger.Info("Firestore client initialized for quota (project=%s)", projectID)
	})
	if initErr != nil {
		return nil, initErr
	}
	return fsClient, nil
}

// Consume increments the daily counter for uid/kind. Returns ErrExceeded when over limit.
func Consume(ctx context.Context, uid string, kind Kind) error {
	if Disabled() || uid == "" || uid == "local-dev" {
		return nil
	}
	limit := dailyLimit(kind)
	if limit == 0 {
		return nil
	}

	client, err := ensureClient(ctx)
	if err != nil {
		return err
	}

	today := time.Now().UTC().Format("2006-01-02")
	ref := client.Collection("quotas").Doc(uid)

	return client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var d doc
		snap, err := tx.Get(ref)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return err
			}
			d = doc{DailyDate: today}
		} else if err := snap.DataTo(&d); err != nil {
			return err
		}

		if d.DailyDate != today {
			d = doc{DailyDate: today}
		}

		switch kind {
		case KindChat:
			if d.ChatCount >= limit {
				return ErrExceeded{Kind: kind, Limit: limit, Count: d.ChatCount}
			}
			d.ChatCount++
		case KindTTS:
			if d.TtsCount >= limit {
				return ErrExceeded{Kind: kind, Limit: limit, Count: d.TtsCount}
			}
			d.TtsCount++
		default:
			return fmt.Errorf("unknown quota kind: %s", kind)
		}
		d.UpdatedAt = time.Now().UTC()
		return tx.Set(ref, d)
	})
}

type ErrExceeded struct {
	Kind  Kind
	Limit int
	Count int
}

func (e ErrExceeded) Error() string {
	return fmt.Sprintf("quota exceeded: kind=%s count=%d limit=%d", e.Kind, e.Count, e.Limit)
}

func IsExceeded(err error) bool {
	_, ok := err.(ErrExceeded)
	return ok
}
