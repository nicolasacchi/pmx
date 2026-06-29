package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempHome points HOME at a temp dir so config round-trips don't touch the
// real ~/.config/pmx/config.toml.
func withTempHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	_ = filepath.Join(dir, ".config", "pmx")
}

func TestAddListResolveRoundTrip(t *testing.T) {
	withTempHome(t)
	// Ensure no env leakage from the host into resolution.
	t.Setenv("POSTMARK_SERVER_TOKEN", "")
	t.Setenv("POSTMARK_ACCOUNT_TOKEN", "")

	if err := AddProject("production", "srv-tok", "acct-tok"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	cfg, err := ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if cfg.DefaultProject != "production" {
		t.Errorf("default project = %q, want production", cfg.DefaultProject)
	}
	p, ok := cfg.Projects["production"]
	if !ok {
		t.Fatal("production project missing")
	}
	if p.ServerToken != "srv-tok" || p.AccountToken != "acct-tok" {
		t.Errorf("tokens = %q/%q, want srv-tok/acct-tok", p.ServerToken, p.AccountToken)
	}

	creds, err := LoadCredentials("", "", "")
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds.ServerToken != "srv-tok" || creds.AccountToken != "acct-tok" {
		t.Errorf("resolved %q/%q, want srv-tok/acct-tok", creds.ServerToken, creds.AccountToken)
	}
}

func TestFlagBeatsEnvBeatsFile(t *testing.T) {
	withTempHome(t)
	if err := AddProject("p", "file-srv", "file-acct"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	t.Setenv("POSTMARK_SERVER_TOKEN", "env-srv")
	t.Setenv("POSTMARK_ACCOUNT_TOKEN", "")

	creds, err := LoadCredentials("flag-srv", "", "")
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds.ServerToken != "flag-srv" {
		t.Errorf("server token = %q, want flag-srv (flag wins)", creds.ServerToken)
	}
	if creds.AccountToken != "file-acct" {
		t.Errorf("account token = %q, want file-acct (falls back to file)", creds.AccountToken)
	}
}

func TestRemoveProject(t *testing.T) {
	withTempHome(t)
	_ = AddProject("a", "s", "")
	_ = AddProject("b", "s2", "")
	if err := RemoveProject("a"); err != nil {
		t.Fatalf("RemoveProject: %v", err)
	}
	cfg, _ := ListProjects()
	if _, ok := cfg.Projects["a"]; ok {
		t.Error("project a should be removed")
	}
	if _, ok := cfg.Projects["b"]; !ok {
		t.Error("project b should remain")
	}
}

func TestMaskKey(t *testing.T) {
	if MaskKey("") != "" {
		t.Error("empty key should mask to empty")
	}
	if MaskKey("short") != "***" {
		t.Error("short key should mask to ***")
	}
	masked := MaskKey("abcdefghijklmnop")
	if masked == "abcdefghijklmnop" || len(masked) == 0 {
		t.Errorf("long key not masked: %q", masked)
	}
}

func TestConfigFileIsScoped(t *testing.T) {
	withTempHome(t)
	_ = AddProject("x", "tok", "")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "pmx", "config.toml")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("config not written to %s: %v", want, err)
	}
}
