package main

import (
	"os"
	"time"

	"github.com/niksteff/minlog"
)

func main() {
	// a little example for development
	
	logger := minlog.New(minlog.WithTarget(os.Stdout))
	logger.Log(minlog.InfoLevel, "hello, world!")
	logger.Info("foo, bar")
	logger.Infof("foo, bar%s", "!")
	time.Sleep(time.Second * 3)
}