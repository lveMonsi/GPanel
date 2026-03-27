package service

import (
	"gpanel/agent/global"
	"gpanel/agent/models"
	"log"
	"time"
)

var monitorTicker *time.Ticker
var monitorStop chan bool

func StartMonitorScheduler() {
	monitorStop = make(chan bool)
	go func() {
		service := NewMonitorService()

		for {
			var setting models.MonitorSetting
			global.DB.FirstOrCreate(&setting, models.MonitorSetting{ID: 1})

			if setting.Enabled {
				if err := service.CollectData(); err != nil {
					log.Printf("Monitor collect error: %v", err)
				}

				if setting.RetentionDays > 0 {
					service.CleanOldData(setting.RetentionDays)
				}
			}

			interval := time.Duration(setting.CollectInterval) * time.Second
			if interval < 10*time.Second {
				interval = 60 * time.Second
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
