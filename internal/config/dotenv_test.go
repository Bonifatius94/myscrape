package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotEnv(t *testing.T) {
	m := parseDotEnv("# a comment\n\n  # indented\nA=1\nexport B=two\nC=\"q v\"\nD='s'\n  E = sp \nEMPTY=\nNOEQ\n")
	want := map[string]string{"A": "1", "B": "two", "C": "q v", "D": "s", "E": "sp"}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("%s = %q, want %q", k, m[k], v)
		}
	}
	if _, ok := m["EMPTY"]; ok {
		t.Errorf("empty value should be skipped")
	}
	if _, ok := m["NOEQ"]; ok {
		t.Errorf("line without = should be skipped")
	}
}

func TestLoadDotEnvRealEnvWins(t *testing.T) {
	p := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(p, []byte("MYSCRAPE_X=fromfile\nMYSCRAPE_Y=yval\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MYSCRAPE_X", "fromenv") // already set -> file must not override
	t.Setenv("MYSCRAPE_Y", "")        // unset effectively

	LoadDotEnv(p)

	if os.Getenv("MYSCRAPE_X") != "fromenv" {
		t.Errorf("real env should win, got %q", os.Getenv("MYSCRAPE_X"))
	}
	if os.Getenv("MYSCRAPE_Y") != "yval" {
		t.Errorf("file value should fill an empty var, got %q", os.Getenv("MYSCRAPE_Y"))
	}
}

func TestLoadDotEnvMissingFileIsNoop(t *testing.T) {
	if n := LoadDotEnv(filepath.Join(t.TempDir(), "nope.env")); n != 0 {
		t.Fatalf("missing file should load nothing, got %d", n)
	}
}
