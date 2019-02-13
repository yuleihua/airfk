// Copyright 2018 The huayulei_2003@hotmail.com Authors
// This file is part of the airfk library.
//
// The airfk library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The airfk library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the airfk library. If not, see <http://www.gnu.org/licenses/>.
package registry

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultPrefixService = "airman"
)

type Service struct {
	Name    string        `json:"name"`
	Version string        `json:"version"`
	Check   string        `json:"check"`
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
		Id:   fmt.Sprintf("/%s/%s%s/%s/%d", DefaultPrefixService, name, version, hostIP, hostPort),
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

func tagsCheck(tags []string, input string) string {
	for i := 0; i < len(tags); i++ {
		if strings.Compare(tags[i], input) == 0 {
			return string([]byte(tags[i])[2:])
		}
	}
	return ""
}
