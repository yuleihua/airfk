package conf

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"airman.com/airfk/pkg/types"
)

const (
	DefaultHTTPPort = 5050
	DefaultWSPort   = 5051
	Version         = "0.1"
)

type Config struct {
	Name        string         `toml:"-" json:"name"`
	Version     *types.Version `toml:"-" json:"version"`
	DataDir     string         `toml:""  json:"dataDir"`
	HTTPHost    string         `toml:",omitempty" json:"http_host"`
	HTTPPort    int            `toml:",omitempty" json:"http_port"`
	HTTPOrigins []string       `toml:",omitempty" json:"http_origins"`
	HTTPModules []string       `toml:",omitempty" json:"http_modules"`
	WSHost      string         `toml:",omitempty" json:"ws_host"`
	WSPort      int            `toml:",omitempty" json:"ws_port"`
	WSOrigins   []string       `toml:",omitempty" json:"ws_origins"`
	WSModules   []string       `toml:",omitempty" json:"ws_modules"`
}

// DefaultConfig contains reasonable default settings.
var DefaultConfig = &Config{
	Name:     "task",
	Version:  types.NewVersion(Version),
	DataDir:  DefaultDataDir(),
	HTTPHost: "localhost",
	WSHost:   "localhost",
	HTTPPort: DefaultHTTPPort,
	WSPort:   DefaultWSPort,
}

// DefaultDataDir is the default data directory to use for modules.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "task", "modules")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "task", "modules")
		} else {
			return filepath.Join(home, ".task", "modules")
		}
	}
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// GetWSHost returns websocket host.
func (c *Config) GetWSHost() string {
	return c.WSHost
}

// GetWSPort returns websocket port.
func (c *Config) GetWSPort() int {
	return c.WSPort
}
