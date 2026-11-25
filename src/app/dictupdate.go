package main

// Spelling dictionary: served from the copy embedded in the exe, or from a newer
// copy downloaded into %LOCALAPPDATA%\YangiAlifbo\dict. Updates come from a plain
// file host (GitHub), described by manifest.json; no server of our own is needed.

import (
	"crypto/ed25519"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	appVersion          = "1.9.3"
	embeddedDictVersion = "2026.11.18"
	dictBaseURL         = "https://raw.githubusercontent.com/otashnst/yangi-alifbo-lugat/main/"
)

var dictFiles = []string{"uz_UZ-2026.aff", "uz_UZ-2026.dic", "freq.txt"}

//go:embed dict
var dictFS embed.FS

type fileInfo struct {
	Sha256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type manifest struct {
	Version string              `json:"version"`
	MinApp  string              `json:"min_app"`
	Notes   string              `json:"notes"`
	Files   map[string]fileInfo `json:"files"`
	Sig     string              `json:"sig"`
}

// verify checks the author's Ed25519 signature over the version and every file hash.
func (m *manifest) verify() error {
	msg := "yangi-alifbo-lugat|" + m.Version + "|" + m.MinApp + "|" + m.Notes
	for _, name := range dictFiles {
		fi, ok := m.Files[name]
		if !ok {
			return errors.New("missing " + name)
		}
		msg += "|" + name + ":" + fi.Sha256 + ":" + strconv.FormatInt(fi.Size, 10)
	}
	pub, err := base64.StdEncoding.DecodeString(releasePubKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return errors.New("bad public key")
	}
	sig, err := base64.StdEncoding.DecodeString(m.Sig)
	if err != nil || !ed25519.Verify(ed25519.PublicKey(pub), []byte(msg), sig) {
		return errors.New("bad signature")
	}
	return nil
}

type dictInfo struct {
	Version string `json:"version"`
	Source  string `json:"source"` // "embedded" | "updated"
	Status  string `json:"status"` // "idle" | "checking" | "latest" | "updated" | "offline" | "error" | "app_old"
	Message string `json:"message"`
	Checked string `json:"checked"`
}

type dictStore struct {
	mu         sync.Mutex
	dir        string
	baseURL    string
	info       dictInfo
	busy       bool
	netAllowed func() bool
}

func newDictStore(baseDir, baseURL string, netAllowed func() bool) *dictStore {
	d := &dictStore{dir: filepath.Join(baseDir, "dict"), baseURL: baseURL, netAllowed: netAllowed}
	d.info = dictInfo{Version: embeddedDictVersion, Source: "embedded", Status: "idle"}
	if m, err := d.localManifest(); err == nil && cmpVersion(m.Version, embeddedDictVersion) > 0 && m.verify() == nil && d.verifyLocal(m) == nil {
		d.info.Version, d.info.Source = m.Version, "updated"
	}
	return d
}

func (d *dictStore) localManifest() (*manifest, error) {
	b, err := os.ReadFile(filepath.Join(d.dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func sha256File(p string) (string, int64, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, err
}

func (d *dictStore) verifyLocal(m *manifest) error {
	for _, name := range dictFiles {
		fi, ok := m.Files[name]
		if !ok {
			return errors.New("missing " + name)
		}
		sum, n, err := sha256File(filepath.Join(d.dir, name))
		if err != nil || sum != fi.Sha256 || n != fi.Size {
			return errors.New("bad " + name)
		}
	}
	return nil
}

func (d *dictStore) snapshot() dictInfo {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.info
}

func (d *dictStore) set(f func(*dictInfo)) {
	d.mu.Lock()
	f(&d.info)
	d.mu.Unlock()
}

// cmpVersion compares dotted numeric versions: 2026.11.18 < 2026.12.1
func cmpVersion(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

func fetch(url string, limit int64) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "YangiAlifboKonvertori/"+appVersion)
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

// check looks for a newer dictionary and installs it. Safe to call concurrently.
func (d *dictStore) check() dictInfo {
	if d.netAllowed != nil && !d.netAllowed() {
		d.set(func(i *dictInfo) { i.Status, i.Message = "disabled", "" })
		return d.snapshot()
	}
	d.mu.Lock()
	if d.busy {
		d.mu.Unlock()
		return d.snapshot()
	}
	d.busy = true
	d.info.Status, d.info.Message = "checking", ""
	cur := d.info.Version
	d.mu.Unlock()
	defer func() { d.mu.Lock(); d.busy = false; d.mu.Unlock() }()

	done := func(status, msg string) dictInfo {
		d.set(func(i *dictInfo) {
			i.Status, i.Message, i.Checked = status, msg, time.Now().Format("2006-01-02 15:04")
		})
		logln("dict check: " + status + " " + msg)
		return d.snapshot()
	}

	b, err := fetch(d.baseURL+"manifest.json", 1<<20)
	if err != nil {
		return done("offline", "Internetga ulanib bo‘lmadi")
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil || m.Version == "" {
		return done("error", "Yangilanish ma’lumoti noto‘g‘ri")
	}
	if err := m.verify(); err != nil {
		return done("error", "Lug‘at imzosi tasdiqlanmadi")
	}
	if cmpVersion(m.Version, cur) <= 0 {
		return done("latest", "")
	}
	if m.MinApp != "" && cmpVersion(appVersion, m.MinApp) < 0 {
		return done("app_old", "Yangi lug‘at uchun dasturning "+m.MinApp+" versiyasi kerak")
	}
	tmp := filepath.Join(d.dir, ".yangi")
	os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return done("error", err.Error())
	}
	defer os.RemoveAll(tmp)
	for _, name := range dictFiles {
		fi, ok := m.Files[name]
		if !ok {
			return done("error", "Yangilanishda "+name+" yo‘q")
		}
		data, err := fetch(d.baseURL+name, 64<<20)
		if err != nil {
			return done("offline", "Yuklab olish uzildi: "+name)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != fi.Sha256 || int64(len(data)) != fi.Size {
			return done("error", "Fayl buzilgan: "+name)
		}
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0o644); err != nil {
			return done("error", err.Error())
		}
	}
	// install: files first, manifest last, so a half-finished install is never trusted
	os.Remove(filepath.Join(d.dir, "manifest.json"))
	for _, name := range dictFiles {
		if err := os.Rename(filepath.Join(tmp, name), filepath.Join(d.dir, name)); err != nil {
			return done("error", err.Error())
		}
	}
	mb, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(d.dir, "manifest.json"), mb, 0o644); err != nil {
		return done("error", err.Error())
	}
	d.set(func(i *dictInfo) { i.Version, i.Source = m.Version, "updated" })
	return done("updated", m.Notes)
}

func (d *dictStore) register(mux *http.ServeMux) {
	sub, _ := fs.Sub(dictFS, "dict")
	embedded := http.FileServer(http.FS(sub))
	mux.HandleFunc("/dict/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		i := d.snapshot()
		if d.netAllowed != nil && !d.netAllowed() {
			i.Status, i.Message = "disabled", ""
		}
		json.NewEncoder(w).Encode(i)
	})
	mux.HandleFunc("/dict/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d.check())
	})
	mux.HandleFunc("/dict/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/dict/")
		w.Header().Set("Cache-Control", "no-store")
		if d.snapshot().Source == "updated" {
			for _, f := range dictFiles {
				if f == name {
					http.ServeFile(w, r, filepath.Join(d.dir, name))
					return
				}
			}
		}
		http.StripPrefix("/dict/", embedded).ServeHTTP(w, r)
	})
}
