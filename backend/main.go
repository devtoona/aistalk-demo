package main

import (
	"net/http"
	"os"

	"voice-chat-api-go/logger"
	"voice-chat-api-go/routes"

	"github.com/joho/godotenv"
)

func main() {
	if err := logger.InitLogger("logs"); err != nil {
		logger.Fatal("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting AISTalk demo API")

	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("Continuing without .env file: %v", err)
	}

	router := routes.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "50037"
	}

	logger.Info("Server starting on port %s", port)
	logger.Info("GET  http://localhost:%s/healthz", port)
	logger.Info("POST http://localhost:%s/api/chat", port)
	logger.Info("POST http://localhost:%s/api/avatar/motion", port)
	logger.Info("POST http://localhost:%s/api/event/tts/aivis/synthesize", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatal("Server failed to start: %v", err)
	}
}
