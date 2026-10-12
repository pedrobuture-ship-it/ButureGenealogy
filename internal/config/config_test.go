package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "data")
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{c.DataDir, c.LogDir} {
		if st, err := os.Stat(p); err != nil || !st.IsDir() {
			t.Fatalf("expected dir %s", p)
		}
	}
	if c.DBPath != filepath.Join(dir, "farm.db") {
		t.Fatalf("unexpected db path %s", c.DBPath)
	}
}

func TestLoadReadsFileAndEnv(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"logLevel":"debug"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil || c.LogLevel != "debug" {
		t.Fatalf("got %+v err=%v", c, err)
	}
	t.Setenv(EnvDataDir, "/x/y")
	if d, _ := DefaultDataDir(); d != "/x/y" {
		t.Fatalf("env override ignored: %s", d)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{`), 0o600)
	if _, err := Load(dir); err == nil {
		t.Fatal("expected error")
	}
}
