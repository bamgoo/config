package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/infrago/infra"
	. "github.com/infrago/base"
)

type defaultConfigDriver struct{}

func init() {
	infra.Register(infra.DEFAULT, &defaultConfigDriver{})
}

func (d *defaultConfigDriver) Load(params Map) (Map, error) {
	file := "config.toml"
	if vv, ok := params["file"].(string); ok {
		file = vv
	}
	if vv, ok := params["path"].(string); ok {
		file = vv
	}
	if vv, ok := params["config"].(string); ok {
		file = vv
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	format, _ := params["format"].(string)
	if format == "" {
		ext := strings.ToLower(filepath.Ext(file))
		switch ext {
		case ".json":
			format = "json"
		case ".toml", ".tml":
			format = "toml"
		case ".yaml", ".yml":
			format = "yaml"
		}
	}
	if format == "" {
		format = detectFormat(data)
	}

	return decodeConfig(data, format)
}

func decodeConfig(data []byte, format string) (Map, error) {
	var out Map
	switch strings.ToLower(format) {
	case "json":
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "toml":
		if err := toml.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		return nil, errors.New("Unknown config format: " + format)
	}
}

func detectFormat(data []byte) string {
	s := strings.TrimSpace(string(data))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return "json"
	}
	if looksLikeToml(s) {
		return "toml"
	}
	if looksLikeYaml(s) {
		return "yaml"
	}
	return "toml"
}

var (
	tomlKeyValPattern = regexp.MustCompile(`(?m)^\s*[\w\.\-]+\s*=`)
	tomlSection       = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]\s*$`)
	yamlKeyValPattern = regexp.MustCompile(`(?m)^\s*[\w\.\-]+\s*:\s*`)
	yamlListPattern   = regexp.MustCompile(`(?m)^\s*-\s+`)
)

func looksLikeToml(s string) bool {
	return tomlKeyValPattern.MatchString(s) || tomlSection.MatchString(s)
}

func looksLikeYaml(s string) bool {
	return yamlKeyValPattern.MatchString(s) || yamlListPattern.MatchString(s)
}
