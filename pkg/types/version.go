// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"fmt"
	"runtime"
	"time"
)

// application version
type Version struct {
	ver       string
	runtime   string
	timestamp time.Time
}

func NewVersion(ver string) *Version {
	return &Version{
		ver:       ver,
		runtime:   runtime.Version(),
		timestamp: time.Now(),
	}
}

func (v *Version) String() string {
	str := fmt.Sprintf("%s %s %s", v.ver, v.runtime, TimeToISO8601(v.timestamp))
	return str
}

func (v *Version) WithCommit(gitCommit string) string {
	str := v.String()
	if len(gitCommit) >= 8 {
		str += " " + gitCommit[:8]
	}
	return str
}

func TimeToISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}
