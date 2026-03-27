package service

import (
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"gpanel/agent/models"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type MonitorService struct{}

func NewMonitorService() *MonitorService {
	return &MonitorService{}
}

func (s *MonitorService) CollectData() error {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	loadInfo, _ := load.Avg()
	diskIO, _ := disk.IOCounters()
	netIO, _ := net.IOCounters(false)

	var diskRead, diskWrite uint64
	for _, d := range diskIO {
		diskRead += d.ReadBytes
		diskWrite += d.WriteBytes
	}

	var netRecv, netSent uint64
	if len(netIO) > 0 {
		netRecv = netIO[0].BytesRecv
		netSent = netIO[0].BytesSent
	}

	data := models.MonitorData{
		Timestamp:      time.Now(),
		CPUPercent:     cpuPercent[0],
		MemTotal:       memInfo.Total,
		MemUsed:        memInfo.Used,
		MemPercent:     memInfo.UsedPercent,
		Load1:          loadInfo.Load1,
		Load5:          loadInfo.Load5,
		Load15:         loadInfo.Load15,
		DiskReadBytes:  diskRead,
		DiskWriteBytes: diskWrite,
		NetRecvBytes:   netRecv,
		NetSentBytes:   netSent,
	}

	return global.DB.Create(&data).Error
}

func (s *MonitorService) GetData(startTime, endTime time.Time) ([]dto.MonitorDataResp, error) {
	var data []models.MonitorData
	err := global.DB.Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Order("timestamp ASC").Find(&data).Error
	if err != nil {
		return nil, err
	}

	result := make([]dto.MonitorDataResp, len(data))
	for i, d := range data {
		result[i] = dto.MonitorDataResp{
			Timestamp:      d.Timestamp,
			CPUPercent:     d.CPUPercent,
			MemTotal:       d.MemTotal,
			MemUsed:        d.MemUsed,
			MemPercent:     d.MemPercent,
			Load1:          d.Load1,
			Load5:          d.Load5,
			Load15:         d.Load15,
			DiskReadBytes:  d.DiskReadBytes,
			DiskWriteBytes: d.DiskWriteBytes,
			NetRecvBytes:   d.NetRecvBytes,
			NetSentBytes:   d.NetSentBytes,
		}
	}
	return result, nil
}

func (s *MonitorService) GetSetting() (*dto.MonitorSettingResp, error) {
	var setting models.MonitorSetting
	err := global.DB.First(&setting).Error
	if err != nil {
		return &dto.MonitorSettingResp{
			Enabled:         false,
			RetentionDays:   7,
			CollectInterval: 60,
		}, nil
	}
	return &dto.MonitorSettingResp{
		Enabled:         setting.Enabled,
		RetentionDays:   setting.RetentionDays,
		CollectInterval: setting.CollectInterval,
	}, nil
}

func (s *MonitorService) UpdateSetting(req dto.MonitorSettingReq) error {
	var setting models.MonitorSetting
	global.DB.FirstOrCreate(&setting, models.MonitorSetting{ID: 1})

	if req.Enabled != nil {
		setting.Enabled = *req.Enabled
	}
	if req.RetentionDays != nil {
		setting.RetentionDays = *req.RetentionDays
	}
	if req.CollectInterval != nil {
		setting.CollectInterval = *req.CollectInterval
	}

	return global.DB.Save(&setting).Error
}

func (s *MonitorService) ClearData() error {
	return global.DB.Exec("DELETE FROM monitor_data").Error
}

func (s *MonitorService) CleanOldData(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return global.DB.Where("timestamp < ?", cutoff).Delete(&models.MonitorData{}).Error
}
