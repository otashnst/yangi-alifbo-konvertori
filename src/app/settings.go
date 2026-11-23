package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// appSettings: per-user choices, plus an administrator policy that overrides them.
// Policy (set by IT, e.g. through Group Policy):
//
//	HKLM\SOFTWARE\Policies\YangiAlifbo  Offline (DWORD) = 1
//	HKCU\SOFTWARE\Policies\YangiAlifbo  Offline (DWORD) = 1
type appSettings struct {
	mu      sync.Mutex
	path    string
	Offline bool `json:"offline"`
}

func loadSettings(baseDir string) *appSettings {
	s := &appSettings{path: filepath.Join(baseDir, "settings.json")}
	if b, err := os.ReadFile(s.path); err == nil {
		json.Unmarshal(b, s)
	}
	return s
}

func (s *appSettings) netAllowed() bool {
	s.mu.Lock()
	off := s.Offline
	s.mu.Unlock()
	return !off && !policyOffline()
}

func (s *appSettings) save() error {
	b, _ := json.MarshalIndent(map[string]bool{"offline": s.Offline}, "", "  ")
	return os.WriteFile(s.path, b, 0o644)
}

func (s *appSettings) register(mux *http.ServeMux) {
	mux.HandleFunc("/app/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodPost && !policyOffline() {
			var in struct {
				Offline bool `json:"offline"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&in); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			s.mu.Lock()
			s.Offline = in.Offline
			err := s.save()
			s.mu.Unlock()
			if err != nil {
				logln("settings save failed")
			}
		}
		s.mu.Lock()
		out := map[string]bool{"offline": s.Offline || policyOffline(), "policy": policyOffline()}
		s.mu.Unlock()
		json.NewEncoder(w).Encode(out)
	})
}
