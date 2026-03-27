package service

import (
	"log"
	"time"
)

var monitorTicker *time.Ticker
var monitorStop chan bool

func StartMonitorScheduler() {
	monitorStop = make(chan bool)
	go func() {
		service := GetMonitorService()

		for {
			setting, err := service.ensureSetting()
			if err != nil {
				log.Printf("Monitor load setting error: %v", err)
				time.Sleep(time.Minute)
				continue
			}

			if setting.Enabled {
				if err := service.CollectData(); err != nil {
					log.Printf("Monitor collect error: %v", err)
				}

				if setting.RetentionDays > 0 {
					service.CleanOldData(setting.RetentionDays)
				}
			}

			interval := time.Duration(setting.CollectInterval) * time.Second
			if interval < minMonitorCollectInterval*time.Second {
				interval = defaultMonitorCollectInterval * time.Second
			}

			select {
			case <-time.After(interval):
			case <-monitorStop:
				return
			}
		}
	}()
}

func StopMonitorScheduler() {
	if monitorStop != nil {
		close(monitorStop)
	}
}
