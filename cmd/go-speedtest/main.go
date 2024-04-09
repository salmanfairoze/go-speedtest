package main

import (
	"context"
	"github.com/salmanfairoze/go-speedtest/internal/speed"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"os/user"
	"sync"
)

func main() {
	logger := log.WithField("func", "main")
	logger.Info("starting go speedtest")

	usr, err := user.Current()
	if err != nil {
		logger.Errorf("unable to get homepath: %v", err)
	}

	homePath := usr.HomeDir

	ctx, cancel := context.WithCancel(context.Background())
	st := speed.New(ctx, cancel, homePath)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		st.ExecuteSpeedTestAsync()
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	logger.Info("Interrupt received, shutting down....")
	st.CloseSpeedTest()
}
