//go:build !windows

package main

import "os"

// test builds: YA_POLICY_OFFLINE=1 simulates the Windows policy
func policyOffline() bool { return os.Getenv("YA_POLICY_OFFLINE") == "1" }
