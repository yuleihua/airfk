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
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"airman.com/airfk/pkg/types"
	"airman.com/airfk/pkg/zoo"
)

const (
	libNameDefault = "airman.com/airfk" // pkg base
	version        = "1.0.1"
)

// nodeFileList node files.
var nodeFileList = map[string][]string{
	"admin":     {"api.go", "backend.go", "node.go"},
	"conf":      {"conf.go"},
	"version":   {"version.go"},
	"common":    {"errors.go", "result.go"},
	"subscribe": {"interface.go", "subscribe.go", "subscribe_test.go"},
	"@@@@@":     {"@@@@@.go", "manager.go"},
}

// distFileList dist files.
var distFileList = map[string][]string{
	"etc": {"@@@@@.json"},
	"bin": {""},
}

// gitList git files. if it is not existed, git clone it.
var gitList = []string{
	"airman.com/airfk",
}

// NewProject create project as project namespace default.
func NewProject(website, name string) *zoo.Project {
	rPath := zoo.ProjectPath(website, name)

	f := zoo.ZooFolder{Name: name, AbsPath: rPath}
	f.AddFile("main.go", "main.go.tpl")
	f.AddFile("Makefile", "Makefile.tpl")
	f.AddFile("Dockerfile", "Dockerfile.tpl")
	f.AddFile(".gitignore", "gitignore.tpl")

	node := f.AddFolder("node")
	for k, v := range nodeFileList {
		ks := strings.Replace(k, "@@@@@", name, 1)
		f := node.AddFolder(ks)
		for _, ff := range v {
			fs := strings.Replace(ff, "@@@@@", name, 1)
			if fs != "" {
				f.AddFile(fs, ff+".tpl")
			}
		}
	}

	dist := f.AddFolder("dist")
	for k, v := range distFileList {
		f := dist.AddFolder(k)
		for _, ff := range v {
			fs := strings.Replace(ff, "@@@@@", name, 1)
			if fs != "" {
				f.AddFile(fs, ff+".tpl")
			}
		}
	}

	return &zoo.Project{
		Name:       name,
		Website:    website,
		ProjectDir: rPath,
		LibDir:     libNameDefault,
		Folder:     f,
		IsRemove:   true,
		RelDir:     fmt.Sprintf("%s/%s", website, name),
	}
}

// templatePath
var templatePath = filepath.Join(zoo.ProjectPath("airman.com", "airfk"), "template")

var (
	webName     string
	projectName string
	template    string
	isVersion   bool
)

func init() {
	flag.StringVar(&webName, "w", "airman.com", "website name")
	flag.StringVar(&projectName, "p", "website", "project name")
	flag.StringVar(&template, "t", templatePath, "template path")
	flag.BoolVar(&isVersion, "v", false, "version information")
}

func main() {
	flag.Parse()

	if isVersion {
		fmt.Println(types.NewVersion(version).String())
		os.Exit(0)
	}

	fmt.Printf("project information: %s %s\n", webName, projectName)

	project := NewProject(webName, projectName)
	if err := project.Write(template); err != nil {
		fmt.Println(err)
	}
	fmt.Printf(">>create %s step1: file create ok\n", projectName)

	if err := project.Clone(gitList); err != nil {
		fmt.Println(err)
	}
	fmt.Printf(">>create %s step2: pkg %v is ok\n", projectName, gitList)

	dir := fmt.Sprintf("cd %s", project.ProjectDir)
	modCmd := dir + "; go mod init ; go mod edit -require=airman.com/airfk@v0.0.0 ; go mod edit -replace=airman.com/airfk@v0.0.0=$GOPATH/src/airman.com/airfk ; go mod tidy && make"
	cmd := exec.Command("/bin/sh", "-c", modCmd)
	result, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("exec mod cmd error, %s:%v", modCmd, err)
		os.Exit(1)
	}
	fmt.Println(string(result))
	fmt.Println("Now get your project! ")
}
