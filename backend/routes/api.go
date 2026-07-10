package routes

import (
	"net/http"

	"voice-chat-api-go/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		MaxAge:           300,
		AllowCredentials: false,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	chatH := handler.NewChatHandler()
	motionH := handler.NewMotionHandler()
	eventStream := handler.NewEventStreamHandler()

	r.Post("/api/chat", chatH.ChatHandler)
	r.Post("/api/avatar/motion", motionH.PostInfer)
	r.Route("/api/event", func(r chi.Router) {
		r.Get("/stream/session/start", eventStream.SynthesisSessionStartHandler)
		r.Post("/stream/session/stop", eventStream.SynthesisSessionStopHandler)
		r.Post("/tts/aivis/synthesize", eventStream.SynthesizeAivisAudioHandler)
	})

	return r
}
