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
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"{{ .LibDir }}/pkg/service"

	"{{ .RelDir }}/node/admin"
	"{{ .RelDir }}/node/conf"
	"{{ .RelDir }}/node/version"
	{{ .SnakeCaseName }} "{{ .RelDir }}/node/{{.Name}}"
)

var (
	configFile string
	isPProf    bool
	isVersion  bool

	gitCommit string // commit hash
	buildDate string // build datetime
)

func init() {
	flag.StringVar(&configFile, "c", "conf/website.json", "configure file")
	flag.BoolVar(&isPProf, "p", false, "setting of pprof")
	flag.BoolVar(&isVersion, "v", false, "version information")
}

func startPProf(address string) {
	log.Infof("Starting pprof server, %s", fmt.Sprintf("http://%s/debug/pprof", address))
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Error("Failure in running pprof server", "err", err)
		}
	}()
}

func main() {
	flag.Parse()

	// print version information
	if isVersion {
		version.Info(os.Args[0], gitCommit, buildDate)
		os.Exit(0)
	}

	// pprof enable
	if isPProf {
		startPProf("localhost:6060")
	}

	c := make(chan os.Signal)
	signal.Ignore()
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// setting logger
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006/01/02-15:04:05.000",
	})

	//// setting config
	//if err := config.Setup(configFile); err != nil {
	//	log.Fatalf("config error:%v", err)
	//}

	config := conf.DefaultConfig

	stack, err := admin.NewNode(config)
	if err != nil {
		log.Fatalf("new node error:%v", err)
	}

	log.Info("step1: new node is okay")

	constructor := func(ctx *service.ServiceContext) (service.Service, error) {
		return {{ .SnakeCaseName }}.NewManager(stack), nil
	}
	if err := stack.Register(constructor); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	if err := stack.Start(); err != nil {
		log.Errorf("start node error:%v", err)
		stack.Stop()
	}

	log.Info("step2: node is running now")

	<-c

	// shutdown http server
	log.Error("shutting down server begin")
	if err := stack.Stop(); err != nil {
		log.Errorf("stop node error:%v", err)
		stack.Stop()
	}

	log.Error("shutting down end")
}
