package envconfig

import (
	"testing"
)

func TestProcess(t *testing.T) {
	type Nested struct {
		Token string `envconfig:"TEST_TOKEN"`
	}
	type Config struct {
		Name    string   `envconfig:"TEST_NAME"`
		Enabled bool     `envconfig:"TEST_ENABLED"`
		Tags    []string `envconfig:"TEST_TAGS"`
		Ignored string
		Nested  Nested
	}

	t.Setenv("TEST_NAME", "hello")
	t.Setenv("TEST_ENABLED", "true")
	t.Setenv("TEST_TAGS", "a, b, c")
	t.Setenv("TEST_TOKEN", "secret")

	var cfg Config
	if err := Process(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "hello" {
		t.Errorf("Name = %q, want %q", cfg.Name, "hello")
	}
	if !cfg.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(cfg.Tags) != 3 || cfg.Tags[0] != "a" || cfg.Tags[1] != "b" || cfg.Tags[2] != "c" {
		t.Errorf("Tags = %v, want [a b c]", cfg.Tags)
	}
	if cfg.Ignored != "" {
		t.Errorf("Ignored = %q, want empty", cfg.Ignored)
	}
	if cfg.Nested.Token != "secret" {
		t.Errorf("Nested.Token = %q, want %q", cfg.Nested.Token, "secret")
	}
}

func TestProcessUnsetVars(t *testing.T) {
	type Config struct {
		Name string `envconfig:"TEST_UNSET_VAR"`
	}

	var cfg Config
	cfg.Name = "original"
	if err := Process(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "original" {
		t.Errorf("Name = %q, want %q (should be unchanged)", cfg.Name, "original")
	}
}

func TestProcessNotPointer(t *testing.T) {
	type Config struct{}
	if err := Process(Config{}); err == nil {
		t.Error("expected error for non-pointer, got nil")
	}
}
