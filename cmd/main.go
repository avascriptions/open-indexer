package main

import (
	"open-indexer/handlers"
	"os"
	"os/signal"
	"time"
)

func main() {

	oss := make(chan os.Signal)

	var logger = handlers.GetLogger()

	logger.Info("start indexer")

	handlers.InitFromSnapshot()

	go func() {
		handlers.StartRpc()
	}()

	if handlers.DataSourceType == "rpc" {
		go func() {
			handlers.StartFetch()
		}()
	}

	go func() {
		handlers.StartSync()
	}()

	stop := func() {
		logger.Info("app is stopping")
		go func() {
			handlers.StopRpc()
		}()
		go func() {
			handlers.StopSync()
		}()
		go func() {
			handlers.StopFetch()
		}()

		// wait all stopped
		for handlers.StopSuccessCount < 3 {
			time.Sleep(time.Duration(100) * time.Millisecond)
		}

		// close db
		handlers.CloseDb()

		logger.Info("app stopped.")

	}

	// 监听信号
	signal.Notify(oss, os.Interrupt, os.Kill)

	for {
		select {
		case <-oss: // kill -9 pid，no effect
			logger.Info("stopped by system...")
			stop()
			logger.Info("gracefully stopped.")
			return
		case <-handlers.QuitChan:
			logger.Info("stopped by app.")
			stop()
			logger.Info("app is auto stopped.")
			return
		}
	}
}
