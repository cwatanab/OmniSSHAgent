//go:build windows
// +build windows

package pageant

import (
	"strings"
	"testing"
)

func TestObfuscatedPipeName(t *testing.T) {
	pipeName, err := ObfuscatedPipeName()
	if err != nil {
		t.Fatalf("ObfuscatedPipeName failed: %v", err)
	}
	if pipeName == "" {
		t.Fatal("ObfuscatedPipeName returned empty string")
	}
	t.Logf("Generated pipe name: %s", pipeName)

	// Verify prefix format: pageant.{username}.{hash}
	if !strings.HasPrefix(pipeName, "pageant.") {
		t.Errorf("pipeName does not start with 'pageant.': %s", pipeName)
	}

	parts := strings.Split(pipeName, ".")
	if len(parts) < 3 {
		t.Errorf("pipeName has fewer than 3 parts: %s", pipeName)
	}

	// The last part should be a 64-character hex string (SHA256)
	hashPart := parts[len(parts)-1]
	if len(hashPart) != 64 {
		t.Errorf("hash part length is not 64: %s", hashPart)
	}
}
