package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func sameToken(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// guard lets only this app's own window talk to the local server:
// the Host must be our exact 127.0.0.1:port (stops DNS rebinding), the page
// itself needs ?t=<token>, and every API call needs the X-Token header (other
// programs and web pages do not know it). Fonts are public data.
func guard(next http.Handler, token, addr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.Host != addr {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/fonts/"):
		case r.URL.Path == "/":
			if !sameToken(r.URL.Query().Get("t"), token) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' data:; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
			w.Header().Set("X-Frame-Options", "DENY")
		default:
			if !sameToken(r.Header.Get("X-Token"), token) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

var allowedExt = map[string]bool{".docx": true, ".xlsx": true, ".pptx": true, ".txt": true}

// safeName keeps only a plain file name with one of our extensions (no folders, no "..").
func safeName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, name)
	ext := strings.ToLower(filepath.Ext(name))
	if !allowedExt[ext] {
		name += ".txt"
	}
	if strings.Trim(name, ". ") == "" || strings.HasPrefix(name, ".") {
		name = "hujjat" + name
	}
	return name
}

var saved sync.Map

func rememberSaved(p string) { saved.Store(filepath.Clean(p), true) }
func wasSaved(p string) bool { _, ok := saved.Load(filepath.Clean(p)); return ok }
