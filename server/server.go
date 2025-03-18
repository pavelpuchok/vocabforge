package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/pavelpuchok/vocabforge/job"
	"github.com/pavelpuchok/vocabforge/preply"
)

type Server struct {
	addr   string
	token  string
	logger *slog.Logger
	queue  job.Queue[job.TranslateWordsJob]
}

type PreplyRequest struct {
	UserID     int64             `json:"userId"`
	Vocabulary preply.Vocabulary `json:"vocabulary"`
}

func New(addr string, token string, q job.Queue[job.TranslateWordsJob], logger *slog.Logger) *Server {
	return &Server{addr: addr, token: token, queue: q, logger: logger}
}

func (s Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /preply-sync", func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("POST /preply-sync start processing")

		expectedToken := "Bearer " + s.token
		auth := r.Header.Get("authorization")
		if auth != expectedToken {
			s.logger.Warn("POST /preply-sync unauthorized access try")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var req PreplyRequest
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		if err := decoder.Decode(&req); err != nil {
			s.logger.Warn("POST /preply-sync failed to parse payload", slog.String("error", err.Error()))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := s.queue.Enqueue(ctx, job.TranslateWordsJob{
			Words:          req.Vocabulary.Words.Nodes,
			UserID:         req.UserID,
			TargetLanguage: "ru",
		})
		if err != nil {
			s.logger.Warn("POST /preply-sync unable to enquee words translation", slog.String("error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		s.logger.Info("POST /preply-sync sucessfully enqueued")
	})

	return http.ListenAndServe(s.addr, mux)
}
