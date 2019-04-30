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
package version

import (
	"fmt"
	"runtime"
)

var (
	Version = "1.0.0"
)

func Info(name, commit, buildDate string) {
	fmt.Println("Version:", Version)
	if commit != "" {
		fmt.Println("Git Commit:", commit)
	}
	if buildDate != "" {
		fmt.Println("Build Date:", buildDate)
	}
	if name != "" {
		fmt.Println("Binary Name:", name)
	}
	fmt.Println("Architecture:", runtime.GOARCH)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Operating System:", runtime.GOOS)
}
