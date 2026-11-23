//go:build windows

package main

import "golang.org/x/sys/windows/registry"

func policyOffline() bool {
	for _, root := range []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER} {
		k, err := registry.OpenKey(root, `SOFTWARE\Policies\YangiAlifbo`, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		v, _, err := k.GetIntegerValue("Offline")
		k.Close()
		if err == nil && v == 1 {
			return true
		}
	}
	return false
}
