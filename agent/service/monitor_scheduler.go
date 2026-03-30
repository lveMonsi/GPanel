package service

import (
	"log"
	"time"
)

var monitorStop chan struct{}
var monitorReload chan struct{}

func StartMonitorScheduler() {
	monitorStop = make(chan struct{})
	monitorReload = make(chan struct{}, 1)
	go func() {
		service := GetMonitorService()
		timer := time.NewTimer(0)
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
			case <-monitorReload:
			case <-monitorStop:
				return
			}

			setting, err := service.ensureSetting()
			if err != nil {
				log.Printf("Monitor load setting error: %v", err)
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(time.Minute)
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

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(interval)
		}
	}()
}

func NotifyMonitorScheduler() {
	if monitorReload == nil {
		return
	}

	select {
	case monitorReload <- struct{}{}:
	default:
	}
}

func StopMonitorScheduler() {
	if monitorStop != nil {
		close(monitorStop)
	}
}
