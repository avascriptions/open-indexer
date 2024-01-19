package main

import (
	"flag"
	"open-indexer/handlers"
	"os"
	"os/signal"
	"time"
)

var snapfile = ""

func init() {
	flag.StringVar(&snapfile, "snapshot", "", "the filename of snapshot")

	flag.Parse()
}

func main() {

	oss := make(chan os.Signal)
	done := make(chan bool)
	var interrupt = false

	var logger = handlers.GetLogger()

	logger.Info("start index")

	if snapfile != "" {
		handlers.InitFromSnapshot(snapfile)
	}

	go func() {
		logger.Info("app is started")
		for !interrupt {
			finished, err := handlers.SyncBlock()
			if err != nil {
				if err.Error() == "no more new block" {
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

		handlers.Snapshot()

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
