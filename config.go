package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/infrago/base"
	"github.com/infrago/infra"
)

const (
	configEnvPrefix = "INFRAGO_"

	KEY  = "infrago-config"
	JSON = "json"
	TOML = "toml"
	YAML = "yaml"
)

var (
	errConfigDriverNotFound = errors.New("config driver not found")
	errConfigSourceNotFound = errors.New("config source not found")
)

var (
	module = &Module{drivers: map[string]Driver{}}
	host   = infra.Mount(module)
)

type (
	Module struct {
		drivers map[string]Driver
	}
	Driver interface {
		Load(Map) (Map, error)
	}
)

// Register dispatches config driver registrations.
func (c *Module) Register(name string, value Any) {
	if drv, ok := value.(Driver); ok {
		c.RegisterDriver(name, drv)
	}
}

func Drivers() []string {
	return module.Drivers()
}

func (c *Module) Drivers() []string {
	out := make([]string, 0, len(c.drivers))
	for name := range c.drivers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (c *Module) RegisterDriver(name string, driver Driver) {
	name = normalizeDriverName(name)
	if name == "" {
		name = infra.DEFAULT
	}
	if driver == nil {
		panic("Invalid config driver: " + name)
	}
	if _, ok := c.drivers[name]; ok {
		panic("Config driver already registered: " + name)
	}
	c.drivers[name] = driver
}

// Module methods (no-op for now)
func (c *Module) Config(Map) {}
func (c *Module) Setup()     {}
func (c *Module) Open()      {}
func (c *Module) Start() {
	infra.Log(infra.LogLevelInfo, "config", "module started", Map{"drivers": len(c.drivers)})
}
func (c *Module) Stop()  {}
func (c *Module) Close() {}

func (c *Module) LoadConfig() (Map, error) {
	drvName, params, err := c.Parse()
	if err != nil {
		return nil, err
	}

	if drvName == "" {
		return nil, errConfigSourceNotFound
	}

	driver, ok := c.drivers[drvName]
	if !ok {
		return nil, fmt.Errorf("unknown config driver %q (registered: %s)", drvName, strings.Join(c.Drivers(), ", "))
	}
	return driver.Load(params)
}

// Parse reads env (INFRAGO_*) then args (--key) and returns params + driver name.
func (c *Module) Parse() (string, Map, error) {
	params := Map{}

	// env first
	for k, v := range c.parseEnv() {
		params[k] = v
	}
	// args override env
	for k, v := range c.parseArgs() {
		params[k] = v
	}

	driver := infra.DEFAULT
	if v, ok := params["driver"].(string); ok && strings.TrimSpace(v) != "" {
		driver = normalizeDriverName(v)
	}

	if driver == infra.DEFAULT || driver == "file" {
		if !hasConfigSource(params) {
			if file := defaultConfigFile(); file != "" {
				params["file"] = file
			}
		}
	}

	return driver, params, nil
}

func (c *Module) parseEnv() Map {
	envs := os.Environ()
	params := Map{}

	for _, kv := range envs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		if !strings.HasPrefix(key, configEnvPrefix) {
			continue
		}
		if strings.TrimSpace(val) == "" {
			continue
		}
		k := normalizeParamKey(strings.TrimPrefix(key, configEnvPrefix))
		params[k] = val
	}
	return params
}

func (c *Module) parseArgs() Map {
	args := os.Args[1:]
	params := Map{}

	if len(args) == 1 {
		if isConfigFileArg(args[0]) {
			params["driver"] = infra.DEFAULT
			params["file"] = args[0]
			return params
		}
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			if i == 0 {
				params["driver"] = arg
			}
			continue
		}
		kv := strings.TrimPrefix(arg, "--")
		if kv == "" {
			continue
		}
		if strings.Contains(kv, "=") {
			parts := strings.SplitN(kv, "=", 2)
			params[normalizeParamKey(parts[0])] = parts[1]
			continue
		}
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			params[normalizeParamKey(kv)] = args[i+1]
			i++
		} else {
			params[normalizeParamKey(kv)] = "true"
		}
	}

	return params
}

func hasConfigSource(params Map) bool {
	for _, key := range []string{"file", "path", "config"} {
		if v, ok := params[key].(string); ok && strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

func normalizeDriverName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeParamKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, ".", "_")
	switch key {
	case "config_driver":
		return "driver"
	case "config_file", "configfile":
		return "file"
	case "config_path", "configpath":
		return "path"
	case "config_addr", "redis_addr", "redisaddr":
		return "addr"
	case "redis_host":
		return "host"
	case "redis_server":
		return "server"
	case "redis_port":
		return "port"
	}
	return key
}

func isConfigFileArg(arg string) bool {
	arg = strings.TrimSpace(arg)
	if arg == "" || strings.HasPrefix(arg, "-") {
		return false
	}
	if _, err := os.Stat(arg); err == nil {
		return true
	}
	switch strings.ToLower(filepath.Ext(arg)) {
	case ".json", ".toml", ".tml", ".yaml", ".yml":
		return true
	}
	return strings.ContainsAny(arg, `/\`)
}

func defaultConfigFile() string {
	candidates := []string{"config.toml", "config.json", "config.yaml", "config.yml"}

	if exe := filepath.Base(os.Args[0]); exe != "" {
		name := strings.TrimSuffix(exe, filepath.Ext(exe))
		candidates = append(candidates, name+".toml", name+".json", name+".yaml", name+".yml")
	}

	for _, file := range candidates {
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}

	return ""
}
