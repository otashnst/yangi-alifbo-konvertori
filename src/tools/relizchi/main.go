// relizchi — Yangi alifbo konvertori relizlarini imzolash vositasi.
//
//	relizchi.exe                         (ikki marta bosish) = init
//	relizchi.exe init                    yangi kalit + imzolangan lug'at manifesti
//	relizchi.exe app  <kalit> <setup.exe> <versiya> <izoh> <latest.json>
//	relizchi.exe dict <kalit> <lug'at papkasi> <versiya> <izoh>
package main

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type dictFile struct {
	Name   string
	Sha256 string
	Size   int64
}

const minApp = "1.6.0"

func appMessage(version, setup, sha string, size int64, notes string) string {
	return "yangi-alifbo-konvertori|" + version + "|" + setup + "|" + sha + "|" + strconv.FormatInt(size, 10) + "|" + notes
}

func dictMessage(version, minApp, notes string, files []dictFile) string {
	s := "yangi-alifbo-lugat|" + version + "|" + minApp + "|" + notes
	for _, f := range files {
		s += "|" + f.Name + ":" + f.Sha256 + ":" + strconv.FormatInt(f.Size, 10)
	}
	return s
}

func dictManifest(priv ed25519.PrivateKey, version, notes string, files []dictFile) []byte {
	fm := map[string]any{}
	for _, f := range files {
		fm[f.Name] = map[string]any{"sha256": f.Sha256, "size": f.Size}
	}
	sig := ed25519.Sign(priv, []byte(dictMessage(version, minApp, notes, files)))
	m := map[string]any{"version": version, "min_app": minApp, "notes": notes, "files": fm, "sig": base64.StdEncoding.EncodeToString(sig)}
	b, _ := json.MarshalIndent(m, "", "  ")
	return append(b, '\n')
}

func loadKey(p string) ed25519.PrivateKey {
	kb, err := os.ReadFile(p)
	die(err, "Kalit faylini o‘qib bo‘lmadi")
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(kb)))
	if err != nil || len(seed) != ed25519.SeedSize {
		die(fmt.Errorf("format"), "Kalit fayli noto‘g‘ri")
	}
	return ed25519.NewKeyFromSeed(seed)
}

func hashFile(p string) (string, int64) {
	b, err := os.ReadFile(p)
	die(err, "Faylni o‘qib bo‘lmadi: "+p)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:]), int64(len(b))
}

func die(err error, msg string) {
	if err != nil {
		fmt.Println("XATO:", msg, "—", err)
		pause()
		os.Exit(1)
	}
}

var interactive bool

func pause() {
	if interactive {
		fmt.Print("\nYopish uchun Enter bosing…")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}

func cmdInit() {
	exe, _ := os.Executable()
	dir := filepath.Join(filepath.Dir(exe), "reliz-kaliti")
	privPath := filepath.Join(dir, "release_private_key.txt")
	if _, err := os.Stat(privPath); err == nil {
		fmt.Println("Kalit allaqachon yaratilgan:", privPath)
		fmt.Println("Yangisini yaratmayman — eskisi ustiga yozilib ketmasligi uchun.")
		pause()
		return
	}
	die(os.MkdirAll(dir, 0o700), "Papka yaratib bo‘lmadi")
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	die(err, "Kalit yaratib bo‘lmadi")
	die(os.WriteFile(privPath, []byte(base64.StdEncoding.EncodeToString(priv.Seed())+"\n"), 0o600), "Kalitni yozib bo‘lmadi")
	pubB64 := base64.StdEncoding.EncodeToString(pub)
	die(os.WriteFile(filepath.Join(dir, "public_key.txt"), []byte(pubB64+"\n"), 0o644), "Ochiq kalitni yozib bo‘lmadi")
	die(os.WriteFile(filepath.Join(dir, "manifest.json"), dictManifest(priv, initDictVersion, "Birinchi versiya", initDict), 0o644), "Manifestni yozib bo‘lmadi")
	fmt.Println("Tayyor! Papka:", dir)
	fmt.Println()
	fmt.Println("  release_private_key.txt  — MAXFIY. Hech kimga yubormang, xavfsiz joyda saqlang.")
	fmt.Println("  public_key.txt           — Claude'ga yuboring (maxfiy emas).")
	fmt.Println("  manifest.json            — Claude'ga yuboring (maxfiy emas).")
	fmt.Println()
	fmt.Println("Ochiq kalit:", pubB64)
	if runtime.GOOS == "windows" {
		exec.Command("explorer.exe", dir).Start()
	}
	pause()
}

func cmdApp(a []string) {
	if len(a) != 5 {
		fmt.Println("relizchi app <kalit> <setup.exe> <versiya> <izoh> <latest.json>")
		os.Exit(2)
	}
	priv := loadKey(a[0])
	sha, size := hashFile(a[1])
	setup := filepath.Base(a[1])
	sig := ed25519.Sign(priv, []byte(appMessage(a[2], setup, sha, size, a[3])))
	m := map[string]any{"version": a[2], "setup": setup, "sha256": sha, "size": size, "notes": a[3], "sig": base64.StdEncoding.EncodeToString(sig)}
	b, _ := json.MarshalIndent(m, "", "  ")
	die(os.WriteFile(a[4], append(b, '\n'), 0o644), "Yozib bo‘lmadi")
	fmt.Println(string(b))
}

func cmdDict(a []string) {
	if len(a) != 4 {
		fmt.Println("relizchi dict <kalit> <lug'at papkasi> <versiya> <izoh>")
		os.Exit(2)
	}
	priv := loadKey(a[0])
	var files []dictFile
	for _, n := range []string{"uz_UZ-2026.aff", "uz_UZ-2026.dic", "freq.txt"} {
		sha, size := hashFile(filepath.Join(a[1], n))
		files = append(files, dictFile{n, sha, size})
	}
	out := filepath.Join(a[1], "manifest.json")
	die(os.WriteFile(out, dictManifest(priv, a[2], a[3], files), 0o644), "Yozib bo‘lmadi")
	fmt.Println("Yozildi:", out)
}

func main() {
	if len(os.Args) < 2 {
		interactive = true
		fmt.Println("Yangi alifbo konvertori — reliz kalitini yaratish")
		fmt.Println("=================================================")
		cmdInit()
		return
	}
	switch os.Args[1] {
	case "init":
		cmdInit()
	case "app":
		cmdApp(os.Args[2:])
	case "dict":
		cmdDict(os.Args[2:])
	default:
		fmt.Println("Buyruqlar: init | app | dict")
		os.Exit(2)
	}
}
