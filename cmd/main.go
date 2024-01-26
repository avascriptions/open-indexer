package main

import (
	"open-indexer/handlers"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {

	oss := make(chan os.Signal)
	done := make(chan bool)
	var interrupt = false

	var logger = handlers.GetLogger()

	logger.Info("start indexer")

	handlers.InitFromSnapshot()

	go func() {
		handlers.StartRpc()
	}()

	go func() {
		logger.Info("app is started")
		for !interrupt {
			finished, err := handlers.SyncBlock()
			if err != nil {
				if strings.HasPrefix(err.Error(), "no more new block") {
					logger.Println(err.Error() + ", wait 1s")
					time.Sleep(time.Duration(1) * time.Second)
					continue
				}
				logger.Errorln("sync error:", err)
				break
			}
			if finished {
				break
			}
		}

		handlers.CloseDb()

		go func() {
			handlers.StopRpc()
		}()

		done <- true

	}()

	stop := func() {
		interrupt = true
		<-done
		logger.Info("gracefully stopped")
	}

	// 监听信号
	signal.Notify(oss, os.Interrupt, os.Kill)

	for {
		select {
		case <-oss: // kill -9 pid，no effect
			logger.Info("stop by system...")
			stop()
			return
		case <-done:
			logger.Info("app is stopped.")
			return
		}
	}
}
