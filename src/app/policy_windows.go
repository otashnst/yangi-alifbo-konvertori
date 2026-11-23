//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func policyValue(name string) bool {
	for _, root := range []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER} {
		k, err := registry.OpenKey(root, `SOFTWARE\Policies\YangiAlifbo`, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, err := k.GetIntegerValue(name)
		k.Close()
		if err == nil && v == 1 {
			return true
		}
	}
	return false
}

func policyOffline() bool   { return policyValue("Offline") }
func policyDisableAI() bool { return policyValue("DisableAI") }

// DPAPI: the API key is encrypted for the current Windows user only.
var (
	crypt32           = syscall.NewLazyDLL("crypt32.dll")
	procProtectData   = crypt32.NewProc("CryptProtectData")
	procUnprotectData = crypt32.NewProc("CryptUnprotectData")
	procLocalFree     = syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")
	dpapiEntropy      = []byte("YangiAlifbo/gemini-key/v1")
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func newBlob(d []byte) *dataBlob {
	if len(d) == 0 {
		return &dataBlob{}
	}
	return &dataBlob{cbData: uint32(len(d)), pbData: &d[0]}
}

func (b *dataBlob) bytes() []byte {
	out := make([]byte, b.cbData)
	copy(out, unsafe.Slice(b.pbData, b.cbData))
	return out
}

func protect(plain []byte) ([]byte, error) {
	var out dataBlob
	r, _, err := procProtectData.Call(uintptr(unsafe.Pointer(newBlob(plain))), 0, uintptr(unsafe.Pointer(newBlob(dpapiEntropy))), 0, 0, 0x1 /*UI_FORBIDDEN*/, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	return out.bytes(), nil
}

func unprotect(enc []byte) ([]byte, error) {
	var out dataBlob
	r, _, err := procUnprotectData.Call(uintptr(unsafe.Pointer(newBlob(enc))), 0, uintptr(unsafe.Pointer(newBlob(dpapiEntropy))), 0, 0, 0x1, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(out.pbData)))
	return out.bytes(), nil
}

var _ = windows.Handle(0)
