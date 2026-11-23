package main

import (
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
)

//go:embed index.html
var page []byte

//go:embed fonts
var fontFS embed.FS

const appTitle = "Yangi alifbo konvertori"

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	comdlg32        = syscall.NewLazyDLL("comdlg32.dll")
	procMessageBoxW = user32.NewProc("MessageBoxW")
	procGetSaveFile = comdlg32.NewProc("GetSaveFileNameW")
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func messageBox(text string, flags uintptr) int {
	r, _, _ := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(appTitle))), flags)
	return int(r)
}

type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	Flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	FlagsEx           uint32
}

func filterFor(ext string) []uint16 {
	names := map[string]string{"docx": "Word hujjati", "xlsx": "Excel jadvali", "pptx": "PowerPoint taqdimoti", "txt": "Matn fayli"}
	n := names[ext]
	if n == "" {
		n = "Fayl"
	}
	s := n + " (*." + ext + ")\x00*." + ext + "\x00Barcha fayllar (*.*)\x00*.*\x00\x00"
	return syscall.StringToUTF16(s)
}

// saveDialog shows the Windows "Save as" dialog; returns "" when cancelled.
func saveDialog(owner uintptr, name string) string {
	buf := make([]uint16, 1024)
	copy(buf, syscall.StringToUTF16(name))
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	filter := filterFor(ext)
	home, _ := os.UserHomeDir()
	initDir := filepath.Join(home, "Documents")
	ofn := openFileName{
		hwndOwner:       owner,
		lpstrFilter:     &filter[0],
		nFilterIndex:    1,
		lpstrFile:       &buf[0],
		nMaxFile:        uint32(len(buf)),
		lpstrInitialDir: u16(initDir),
		lpstrTitle:      u16("Yangi alifbodagi faylni saqlash"),
		Flags:           0x2 | 0x8 | 0x800 | 0x80000, // OVERWRITEPROMPT|NOCHANGEDIR|PATHMUSTEXIST|EXPLORER
		lpstrDefExt:     u16(ext),
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	r, _, _ := procGetSaveFile.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func jsq(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", "", " ", ` `, " ", ` `)
	return `"` + r.Replace(s) + `"`
}

func startServer(exitWhenIdle bool) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		messageBox("Dasturni ishga tushirib bo‘lmadi: "+err.Error(), 0x10)
		os.Exit(1)
	}
	var last atomic.Int64
	last.Store(time.Now().Unix() + 60)
	token := randomToken()
	mux := http.NewServeMux()
	settings := loadSettings(baseDir)
	settings.register(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(page)
	})
	dictURL := dictBaseURL
	if v := os.Getenv("YA_DICT_URL"); v != "" {
		dictURL = v
	}
	store := newDictStore(baseDir, dictURL, settings.netAllowed)
	store.register(mux)
	go func() {
		time.Sleep(3 * time.Second)
		if settings.netAllowed() {
			store.check()
		}
	}()
	appURL := appReleaseURL
	if v := os.Getenv("YA_APP_URL"); v != "" {
		appURL = v
	}
	upd := newAppUpdater(appURL, func() { os.Exit(0) }, settings.netAllowed)
	upd.register(mux)
	go func() {
		time.Sleep(5 * time.Second)
		if settings.netAllowed() {
			upd.check()
		}
	}()
	fsub, _ := fs.Sub(fontFS, "fonts")
	mux.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.FS(fsub))))
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		last.Store(time.Now().Unix())
		w.WriteHeader(204)
	})
	if exitWhenIdle {
		go func() {
			for {
				time.Sleep(3 * time.Second)
				if time.Now().Unix()-last.Load() > 12 {
					os.Exit(0)
				}
			}
		}()
	}
	addr := ln.Addr().String()
	srv := &http.Server{Handler: guard(mux, token, addr), ReadHeaderTimeout: 10 * time.Second}
	go srv.Serve(ln)
	return "http://" + addr + "/?t=" + token
}

func init() { runtime.LockOSThread() }

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procSetStdHandle = kernel32.NewProc("SetStdHandle")
	procCommDlgErr   = comdlg32.NewProc("CommDlgExtendedError")
	baseDir          = filepath.Join(os.Getenv("LOCALAPPDATA"), "YangiAlifbo")
	logf             *os.File
)

func logln(s string) {
	if logf != nil {
		logf.WriteString(time.Now().Format("2006-01-02 15:04:05 ") + s + "\r\n")
	}
}

// autoSave writes into Documents\Yangi alifbo without any dialog.
func autoSave(name string, data []byte) (string, error) {
	name = safeName(name)
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Documents", "Yangi alifbo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	p := filepath.Join(dir, name)
	for i := 2; ; i++ {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			break
		}
		p = filepath.Join(dir, stem+" ("+itoa(i)+")"+ext)
	}
	return p, os.WriteFile(p, data, 0o644)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	b := []byte{}
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

// runDialogChild: separate process that only shows the Save dialog and prints the chosen path.
func runDialogChild(owner uintptr, name string) {
	ole32 := syscall.NewLazyDLL("ole32.dll")
	ole32.NewProc("CoInitializeEx").Call(0, 2)
	path := saveDialog(owner, name)
	if path == "" {
		code, _, _ := procCommDlgErr.Call()
		if code != 0 {
			os.Exit(2)
		}
		os.Exit(1)
	}
	os.Stdout.WriteString(path)
	os.Exit(0)
}

func main() {
	if len(os.Args) == 4 && os.Args[1] == "--save-dialog" {
		h, _ := strconv.ParseUint(os.Args[2], 10, 64)
		runDialogChild(uintptr(h), safeName(os.Args[3]))
		return
	}
	os.MkdirAll(baseDir, 0o755)
	dataDir := filepath.Join(baseDir, "WebView2")
	os.MkdirAll(dataDir, 0o755)
	if f, err := os.OpenFile(filepath.Join(baseDir, "log.txt"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
		logf = f
		os.Stderr = f
		procSetStdHandle.Call(uintptr(0xFFFFFFF4), f.Fd()) // STD_ERROR_HANDLE: Go crash reports go to log.txt
	}
	safeMode := false

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath:  dataDir,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  appTitle,
			Width:  1000,
			Height: 860,
			IconId: 1,
			Center: true,
		},
	})
	if w == nil {
		// WebView2 runtime missing: offer browser mode
		if messageBox("Kompyuterda Microsoft Edge WebView2 komponenti topilmadi, shuning uchun dastur o‘z oynasida ochila olmaydi.\n\nDasturni brauzerda ochaymi?\n\n(WebView2 ni Microsoft saytidan bepul o‘rnatsangiz, keyingi safar o‘z oynasida ochiladi.)", 0x4|0x30) == 6 {
			url := startServer(true)
			exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			select {}
		}
		return
	}
	defer w.Destroy()
	w.SetSize(760, 600, webview2.HintMin)

	w.Bind("saveFile", func(id, name, b64 string) {
		owner := uintptr(w.Window())
		finish := func(status, info string) {
			w.Dispatch(func() {
				w.Eval("window.__saveDone(" + jsq(id) + "," + jsq(status) + "," + jsq(info) + ")")
			})
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logln("panic in save: " + fmt.Sprint(r) + "\r\n" + string(debug.Stack()))
					finish("error", "kutilmagan xato")
				}
			}()
			name = safeName(name)
			data, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				finish("error", "ma’lumot buzilgan")
				return
			}
			auto := func() {
				p, err := autoSave(name, data)
				if err != nil {
					finish("error", err.Error())
				} else {
					logln("autosaved")
					rememberSaved(p)
					finish("saved", p)
				}
			}
			if safeMode {
				auto()
				return
			}
			self, _ := os.Executable()
			cmd := exec.Command(self, "--save-dialog", strconv.FormatUint(uint64(owner), 10), name)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, err := cmd.Output()
			code := 0
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					code = ee.ExitCode()
				} else {
					code = -1
				}
			}
			logln("dialog exit " + strconv.Itoa(code))
			switch {
			case code == 1:
				finish("cancel", "")
			case code != 0 || len(out) == 0:
				auto()
			default:
				path := strings.TrimSpace(string(out))
				if err := os.WriteFile(path, data, 0o644); err != nil {
					logln("write error")
					finish("error", "faylga yozib bo‘lmadi (u Word yoki Excel’da ochiq bo‘lishi mumkin)")
					return
				}
				logln("saved")
				rememberSaved(path)
				finish("saved", path)
			}
		}()
	})
	w.Bind("openFolder", func(path string) {
		if wasSaved(path) { // only folders of files this app itself saved
			exec.Command("explorer.exe", "/select,", path).Start()
		}
	})

	w.Navigate(startServer(false))
	w.Run()
}
