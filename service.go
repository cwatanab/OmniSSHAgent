package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/masahide/OmniSSHAgent/pkg/store"
	"github.com/masahide/OmniSSHAgent/pkg/store/local"
)

type Service struct {
	app    *App
	port   int
	server *http.Server
}

func NewService(port int) *Service {
	return &Service{
		app:  NewApp(),
		port: port,
	}
}

func (s *Service) Start() error {
	// Initialize settings
	s.app.settings = store.NewSettings(AppName, local.NewLocalCred(AppName))
	if err := s.app.settings.Load(); err != nil {
		return err
	}
	Logger.SetEnable(s.app.settings.SaveData.DebugLog)
	s.app.pageantCheckFunc = func() {}

	s.app.initializeKeyRing()
	s.app.startConfiguredAgents()

	// Set up HTTP API router
	mux := http.NewServeMux()
	mux.HandleFunc("/keys", s.handleKeys)
	mux.HandleFunc("/settings", s.handleSettings)
	mux.HandleFunc("/shutdown", s.handleShutdown)
	mux.HandleFunc("/checkkey", s.handleCheckKey)

	s.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	log.Printf("Starting HTTP API service on http://127.0.0.1:%d\n", s.port)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	return nil
}

func (s *Service) handleKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		list, err := s.app.KeyList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(list)
		return
	} else if r.Method == http.MethodPost {
		// Add Key
		var req struct {
			FilePath   string `json:"filePath"`
			Passphrase string `json:"passphrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pk, err := s.app.CheckKeyType(req.FilePath, req.Passphrase)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.app.AddLocalFile(*pk); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	} else if r.Method == http.MethodDelete {
		// Delete Key
		sha256 := r.URL.Query().Get("sha256")
		if sha256 == "" {
			http.Error(w, "missing sha256", http.StatusBadRequest)
			return
		}
		if err := s.app.keyRing.RemoveKey(sha256); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := s.app.keyRing.DeleteKeySettings(sha256); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Service) handleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		settings := s.app.GetSettings()
		json.NewEncoder(w).Encode(settings)
		return
	} else if r.Method == http.MethodPost {
		var settings store.SaveData
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := s.app.Save(settings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Service) handleCheckKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		var req struct {
			FilePath   string `json:"filePath"`
			Passphrase string `json:"passphrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pk, err := s.app.CheckKeyType(req.FilePath, req.Passphrase)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(pk)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Service) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"shutting_down"}`))

	go func() {
		s.app.shutdown(context.Background())
		_ = s.server.Shutdown(context.Background())
		os.Exit(0)
	}()
}
