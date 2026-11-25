package main

// Self-update: reads latest.json from the GitHub release, checks the Ed25519
// signature with the public key built into the exe, downloads the installer,
// verifies its SHA-256 and runs it silently. No server of our own is needed.

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	appReleaseURL = "https://raw.githubusercontent.com/otashnst/yangi-alifbo-konvertori/main/release/"
	// Ed25519 public key of the release signing key (the private half stays with the author).
	releasePubKey = "KMIIzV7Yg4m7Haig88pZoUQXjOQhv0Mkst7ENdjMVHk="
)

type releaseInfo struct {
	Version string `json:"version"`
	Setup   string `json:"setup"`
	Sha256  string `json:"sha256"`
	Size    int64  `json:"size"`
	Notes   string `json:"notes"`
	Sig     string `json:"sig"`
}

// signedMessage is what the author signs; any change to it breaks the signature.
func (r *releaseInfo) signedMessage() []byte {
	return []byte("yangi-alifbo-konvertori|" + r.Version + "|" + r.Setup + "|" + r.Sha256 + "|" + strconv.FormatInt(r.Size, 10) + "|" + r.Notes)
}

func (r *releaseInfo) verify() error {
	pub, err := base64.StdEncoding.DecodeString(releasePubKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return errors.New("bad public key")
	}
	sig, err := base64.StdEncoding.DecodeString(r.Sig)
	if err != nil || !ed25519.Verify(ed25519.PublicKey(pub), r.signedMessage(), sig) {
		return errors.New("imzo noto‘g‘ri")
	}
	return nil
}

type appUpdateState struct {
	Current   string `json:"current"`
	Status    string `json:"status"` // idle | checking | latest | available | downloading | installing | offline | error
	Version   string `json:"version"`
	Notes     string `json:"notes"`
	Message   string `json:"message"`
	Installed bool   `json:"installed"` // running from the installed copy (setup can replace it)
}

type appUpdater struct {
	mu         sync.Mutex
	baseURL    string
	st         appUpdateState
	rel        *releaseInfo
	quit       func()
	netAllowed func() bool
}

func newAppUpdater(baseURL string, quit func(), netAllowed func() bool) *appUpdater {
	u := &appUpdater{baseURL: baseURL, quit: quit, netAllowed: netAllowed}
	u.st = appUpdateState{Current: appVersion, Status: "idle", Installed: runningInstalled()}
	return u
}

func runningInstalled() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	inst := filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Yangi alifbo konvertori")
	rel, err := filepath.Rel(inst, exe)
	return err == nil && !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.'
}

func (u *appUpdater) snapshot() appUpdateState    { u.mu.Lock(); defer u.mu.Unlock(); return u.st }
func (u *appUpdater) set(f func(*appUpdateState)) { u.mu.Lock(); f(&u.st); u.mu.Unlock() }

func (u *appUpdater) check() appUpdateState {
	if u.netAllowed != nil && !u.netAllowed() {
		u.set(func(s *appUpdateState) { s.Status, s.Message = "disabled", "" })
		return u.snapshot()
	}
	u.set(func(s *appUpdateState) { s.Status, s.Message = "checking", "" })
	b, err := fetch(u.baseURL+"latest.json", 1<<16)
	if err != nil {
		logln("app check: offline " + err.Error())
		u.set(func(s *appUpdateState) { s.Status = "offline" })
		return u.snapshot()
	}
	var r releaseInfo
	if err := json.Unmarshal(b, &r); err != nil || r.Version == "" {
		u.set(func(s *appUpdateState) { s.Status, s.Message = "error", "latest.json noto‘g‘ri" })
		return u.snapshot()
	}
	if err := r.verify(); err != nil {
		logln("app check: signature rejected for " + r.Version)
		u.set(func(s *appUpdateState) { s.Status, s.Message = "error", "Yangilanish imzosi tasdiqlanmadi" })
		return u.snapshot()
	}
	if cmpVersion(r.Version, appVersion) <= 0 {
		u.set(func(s *appUpdateState) { s.Status = "latest" })
		return u.snapshot()
	}
	u.mu.Lock()
	u.rel = &r
	u.st.Status, u.st.Version, u.st.Notes = "available", r.Version, r.Notes
	u.mu.Unlock()
	logln("app check: available " + r.Version)
	return u.snapshot()
}

// install downloads, verifies and starts the installer, then quits the app.
func (u *appUpdater) install() appUpdateState {
	if u.netAllowed != nil && !u.netAllowed() {
		return u.snapshot()
	}
	u.mu.Lock()
	r := u.rel
	if r == nil || u.st.Status == "downloading" || u.st.Status == "installing" {
		u.mu.Unlock()
		return u.snapshot()
	}
	u.st.Status, u.st.Message = "downloading", ""
	u.mu.Unlock()
	fail := func(msg string) appUpdateState {
		logln("app install: " + msg)
		u.set(func(s *appUpdateState) { s.Status, s.Message = "available", msg })
		return u.snapshot()
	}
	cl := &http.Client{Timeout: 10 * time.Minute}
	req, _ := http.NewRequest("GET", u.baseURL+r.Setup, nil)
	req.Header.Set("User-Agent", "YangiAlifboKonvertori/"+appVersion)
	resp, err := cl.Do(req)
	if err != nil {
		return fail("Yuklab bo‘lmadi: internetni tekshiring")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fail("Yuklab bo‘lmadi (HTTP " + strconv.Itoa(resp.StatusCode) + ")")
	}
	dir := filepath.Join(os.TempDir(), "YangiAlifboUpdate")
	os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "setup-"+r.Version+".exe")
	f, err := os.Create(path)
	if err != nil {
		return fail(err.Error())
	}
	h := sha256.New()
	n, err := copyBoth(f, h, resp.Body, r.Size+1)
	f.Close()
	if err != nil || n != r.Size || hex.EncodeToString(h.Sum(nil)) != r.Sha256 {
		os.Remove(path)
		return fail("Yuklangan fayl buzilgan, qayta urinib ko‘ring")
	}
	u.set(func(s *appUpdateState) { s.Status = "installing" })
	if os.Getenv("YA_UPDATE_DRYRUN") != "" {
		logln("app install: dry run ok")
		return u.snapshot()
	}
	if err := exec.Command(path, "/S").Start(); err != nil {
		return fail("O‘rnatuvchini ishga tushirib bo‘lmadi")
	}
	logln("app install: setup started")
	go func() { time.Sleep(700 * time.Millisecond); u.quit() }()
	return u.snapshot()
}

func copyBoth(dst io.Writer, h io.Writer, src io.Reader, limit int64) (int64, error) {
	n, err := io.Copy(io.MultiWriter(dst, h), io.LimitReader(src, limit))
	if err == nil && n >= limit {
		err = errors.New("too large")
	}
	return n, err
}

func (u *appUpdater) register(mux *http.ServeMux) {
	mux.HandleFunc("/app/update-info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(u.snapshot())
	})
	mux.HandleFunc("/app/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("check") == "1" {
			json.NewEncoder(w).Encode(u.check())
			return
		}
		json.NewEncoder(w).Encode(u.install())
	})
}
