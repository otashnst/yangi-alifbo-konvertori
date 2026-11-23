//go:build !windows

package main

import "os"

// test builds: env vars simulate the Windows policy; "DPAPI" is a no-op.
func policyOffline() bool   { return os.Getenv("YA_POLICY_OFFLINE") == "1" }
func policyDisableAI() bool { return os.Getenv("YA_POLICY_DISABLEAI") == "1" }

func protect(plain []byte) ([]byte, error) { return append([]byte("TEST:"), plain...), nil }
func unprotect(enc []byte) ([]byte, error) {
	if len(enc) < 5 {
		return nil, os.ErrInvalid
	}
	return enc[5:], nil
}
