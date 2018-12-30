package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"airman.com/airfk/node"
	"airman.com/airfk/node/conf"
)

var confname string

func init() {
	flag.StringVar(&confname, "c", "conf/node.ini", "configure file")
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

	<-c

	// shutdown http server
	log.Error(">> shutting down server begin")
	if err := stack.Stop(); err != nil {
		log.Errorf("stop node error:%v", err)
		stack.Stop()
	}

	log.Error(">> shutting down end")
}
