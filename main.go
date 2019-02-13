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
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"airman.com/airfk/node"
	"airman.com/airfk/node/conf"
)

var (
	confname string
	isPProf  bool
)

func init() {
	flag.StringVar(&confname, "c", "conf/node.ini", "configure file")
	flag.BoolVar(&isPProf, "p", false, "setting of pprof")
}

func StartPProf(address string) {
	log.Info("Starting pprof server", "addr", fmt.Sprintf("http://%s/debug/pprof", address))
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Error("Failure in running pprof server", "err", err)
		}
	}()
}

func main() {
	flag.Parse()

	c := make(chan os.Signal)
	signal.Ignore()
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// setting logger
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006/01/02-15:04:05.000",
	})

	config := conf.DefaultConfig

	// setting config

	stack, err := node.NewNode(config)
	if err != nil {
		log.Fatalf("new node error:%v", err)
	}

	log.Info(">> step1: new node is okay")

	if err := stack.Start(); err != nil {
		log.Errorf("start node error:%v", err)
		stack.Stop()
	}

	log.Info(">> step2: node is running now")

	if isPProf {
		StartPProf("localhost:6060")
	}

	<-c

	// shutdown http server
	log.Error(">> shutting down server begin")
	if err := stack.Stop(); err != nil {
		log.Errorf("stop node error:%v", err)
		stack.Stop()
	}

	log.Error(">> shutting down end")
}
