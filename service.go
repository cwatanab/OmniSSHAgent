package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/masahide/OmniSSHAgent/pkg/cygwinsocket"
	"github.com/masahide/OmniSSHAgent/pkg/namedpipe"
	"github.com/masahide/OmniSSHAgent/pkg/pageant"
	"github.com/masahide/OmniSSHAgent/pkg/sshutil"
	"github.com/masahide/OmniSSHAgent/pkg/store"
	"github.com/masahide/OmniSSHAgent/pkg/store/local"
	"github.com/masahide/OmniSSHAgent/pkg/unix"
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

	s.app.keyRing = sshutil.NewKeyRing(s.app.settings)
	if err := s.app.keyRing.AddKeys(); err != nil {
		log.Printf("KeyRing.AddKeys err: %s", err)
	}

	// Create a context for agent goroutines so they can be cancelled on shutdown
	agentCtx, cancel := context.WithCancel(context.Background())
	s.app.agentCtx = agentCtx
	s.app.cancelAgents = cancel

	debug := false

	pa := &pageant.Pageant{
		ExtendedAgent: s.app.keyRing,
		AppName:       AppName,
		Debug:         debug,
		CheckFunc:     func() {}, // No-op for service mode
	}
	if s.app.settings.PageantAgent {
		s.app.wg.Add(1)
		go func() {
			defer s.app.wg.Done()
			pa.RunAgent(s.app.agentCtx)
		}()
	}
	log.Println("Starting pageant...")

	if s.app.settings.NamedPipeAgent {
		pipeName := ""
		na := &namedpipe.NamedPipe{ExtendedAgent: s.app.keyRing, Debug: debug, Name: pipeName}
		log.Println("Starting NamedPipe agent..")
		s.app.wg.Add(1)
		go func() {
			defer s.app.wg.Done()
			if err := na.RunAgent(s.app.agentCtx); err != nil {
				log.Printf("NamedPipe agent error: %v", err)
			}
		}()
	}

	if s.app.settings.UnixSocketAgent {
		ua := &unix.DomainSock{ExtendedAgent: s.app.keyRing, Debug: debug, Path: s.app.settings.UnixSocketPath}
		log.Println("Start Unix domain socket agent..")
		s.app.wg.Add(1)
		go func() {
			defer s.app.wg.Done()
			if err := ua.RunAgent(s.app.agentCtx); err != nil {
				log.Printf("Unix socket agent error: %v", err)
			}
		}()
	}

	if s.app.settings.CygWinAgent {
		ca := &cygwinsocket.CygwinSock{ExtendedAgent: s.app.keyRing, Debug: debug, Path: s.app.settings.CygWinSocketPath}
		log.Println("Starting Cygwin unix domain socket agent..")
		s.app.wg.Add(1)
		go func() {
			defer s.app.wg.Done()
			if err := ca.RunAgent(s.app.agentCtx); err != nil {
				log.Printf("Cygwin socket agent error: %v", err)
			}
		}()
	}

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
