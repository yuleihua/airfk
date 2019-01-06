package registry

import (
	"fmt"
	"strings"
	"time"
)

type Service struct {
	Name    string        `json:"name"`
	Version string        `json:"version"`
	TTL     time.Duration `json:"ttl"`
	Tags    []string      `json:"tags"`
	Nodes   []*Node       `json:"nodes"`
}

type Node struct {
	Id   string   `json:"id"`
	Host string   `json:"host"`
	Port int      `json:"port"`
	Tags []string `json:"tags"`
}

func NewService(name, version, address string, port int) *Service {
	hostIP := address
	if address == "" {
		hostIP = "127.0.0.1"
	}
	hostPort := port
	if port == 0 {
		hostPort = 8500
	}
	node := &Node{
		Id:   fmt.Sprintf("%s:%s:%d", name, hostIP, hostPort),
		Host: hostIP,
		Port: hostPort,
	}
	s := &Service{
		Name:    name,
		Version: version,
		Nodes:   []*Node{node},
	}
	return s
}

func (s *Service) GetTags() []string {
	if s == nil {
		return nil
	}
	tags := []string{"a-" + DefaultTagService}
	if s.Version != "" {
		tags = append(tags, "v-"+s.Version)
	}
	if s.Nodes != nil && s.Nodes[0].Host != "" {
		tags = append(tags, "h-"+s.Nodes[0].Host)
	}
	return tags
}

func (s *Service) GetId() string {
	if s.Nodes != nil && s.Nodes[0].Id != "" {
		return s.Nodes[0].Id
	}
	return ""
}

func tagsVersion(tags []string) string {
	for i := 0; i < len(tags); i++ {
		if strings.HasPrefix(tags[i], "v-") {
			return string([]byte(tags[i])[2:])
		}
	}
	return ""
}

func tagsHost(tags []string) string {
	for i := 0; i < len(tags); i++ {
		if strings.HasPrefix(tags[i], "h-") {
			return string([]byte(tags[i])[2:])
		}
	}
	return ""
}
