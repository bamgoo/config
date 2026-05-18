package config

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/infrago/base"
	"github.com/infrago/infra"
)

func TestParseSingleDriverFlag(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	t.Setenv("INFRAGO_DRIVER", "")
	t.Setenv("INFRAGO_FILE", "")
	t.Setenv("INFRAGO_PATH", "")
	t.Setenv("INFRAGO_CONFIG", "")

	os.Args = []string{"app", "--driver=redis"}

	driver, params, err := (&Module{}).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if driver != "redis" {
		t.Fatalf("driver=%q, want redis", driver)
	}
	if params["file"] == "--driver=redis" {
		t.Fatalf("file=%v, want driver flag to be parsed as an option", params["file"])
	}
}

func TestParseNormalizesDriverName(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	t.Setenv("INFRAGO_DRIVER", "")

	os.Args = []string{"app", "--driver=Redis"}

	driver, _, err := (&Module{}).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if driver != "redis" {
		t.Fatalf("driver=%q, want redis", driver)
	}
}

func TestRegisterDriverNormalizesName(t *testing.T) {
	mod := &Module{drivers: map[string]Driver{}}
	mod.RegisterDriver(" Redis ", &defaultConfigDriver{})

	if _, ok := mod.drivers["redis"]; !ok {
		t.Fatalf("normalized driver not registered: %#v", mod.drivers)
	}
}

func TestDriversReturnsSortedNames(t *testing.T) {
	mod := &Module{drivers: map[string]Driver{}}
	mod.RegisterDriver("redis", &defaultConfigDriver{})
	mod.RegisterDriver("file", &defaultConfigDriver{})

	got := mod.Drivers()
	want := []string{"file", "redis"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("drivers=%v, want %v", got, want)
	}
}

func TestParseConfigFileEnvAlias(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	t.Setenv("INFRAGO_DRIVER", "")
	t.Setenv("INFRAGO_FILE", "")
	t.Setenv("INFRAGO_CONFIG_FILE", "app.yaml")

	os.Args = []string{"app"}

	_, params, err := (&Module{}).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if params["file"] != "app.yaml" {
		t.Fatalf("file=%v, want app.yaml", params["file"])
	}
}

func TestParseFindsDefaultConfigFile(t *testing.T) {
	oldArgs := os.Args
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Args = oldArgs
		_ = os.Chdir(oldWd)
	})
	t.Setenv("INFRAGO_DRIVER", "")
	t.Setenv("INFRAGO_FILE", "")
	t.Setenv("INFRAGO_PATH", "")
	t.Setenv("INFRAGO_CONFIG", "")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"name":"demo"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"app"}

	driver, params, err := (&Module{}).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if driver != infra.DEFAULT {
		t.Fatalf("driver=%q, want %q", driver, infra.DEFAULT)
	}
	if params["file"] != "config.json" {
		t.Fatalf("file=%v, want config.json", params["file"])
	}
}

func TestDecodeEmptyConfig(t *testing.T) {
	cfg, err := Decode([]byte(" \n\t"), "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || len(cfg) != 0 {
		t.Fatalf("cfg=%v, want empty map", cfg)
	}
	if got := DetectFormat([]byte(" \n\t")); got != "" {
		t.Fatalf("format=%q, want empty", got)
	}
}

func TestDefaultDriverNoConfigFile(t *testing.T) {
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg, err := (&defaultConfigDriver{}).Load(Map{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg != nil {
		t.Fatalf("cfg=%v, want nil", cfg)
	}
}
