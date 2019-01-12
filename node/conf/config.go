package conf

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"airman.com/airfk/pkg/common"
	"airman.com/airfk/pkg/types"
)

const (
	DefaultHTTPPort = 5050
	DefaultWSPort   = 5051
	Version         = "0.1"
	NodeId          = "NODEID"
)

type Config struct {
	Name        string         `toml:"" json:"name"`
	Id          string         `toml:"" json:"node_id"`
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
	DataDir:  "/tmp/task",
	HTTPHost: "127.0.0.1",
	WSHost:   "127.0.0.1",
	HTTPPort: DefaultHTTPPort,
	WSPort:   DefaultWSPort,
}

func NewConfig(name, version, dataDir, host string) *Config {
	return NewConfigEx(name, version, "", dataDir, host, "127.0.0.1", DefaultHTTPPort, DefaultWSPort)
}

func NewConfigEx(name, version, nodeId, dataDir, host, wsHost string, port, wsPort int) *Config {
	httpHost := host
	if host == "" {
		httpHost = "127.0.0.1"
	}

	httpPort := port
	if port == 0 {
		port = DefaultHTTPPort
	}

	// node id
	nid := nodeId
	if nid == "" {
		nid := os.Getenv(NodeId)
		if nid == "" {
			if httpHost != "0.0.0.0" && httpHost != "localhost" {
				ips := strings.Split(httpHost, ".")
				nid = ips[len(ips)-1]
			}
		}
	}

	return &Config{
		Name:     name,
		Id:       nid,
		Version:  types.NewVersion(version),
		DataDir:  common.AbsolutePath(DefaultDataDir(), dataDir),
		HTTPHost: httpHost,
		WSHost:   wsHost,
		HTTPPort: httpPort,
		WSPort:   wsPort,
	}
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
func (c *Config) GetNodeId() string {
	if c.Id != "" {
		return c.Id
	}

	nid := os.Getenv(NodeId)
	if nid == "" {
		if c.HTTPHost != "0.0.0.0" && c.HTTPHost != "localhost" {
			ips := strings.Split(c.HTTPHost, ".")
			nid = ips[len(ips)-1]
		}
	}
	c.Id = nid
	return nid
}

// GetWSHost returns websocket host.
func (c *Config) GetWSHost() string {
	return c.WSHost
}

// GetWSPort returns websocket port.
func (c *Config) GetWSPort() int {
	return c.WSPort
}
