package service

import (
	"encoding/json"
	"gpanel/agent/dto"
	"gpanel/agent/global"
	"gpanel/agent/models"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type MonitorService struct {
	lastDiskIO map[string]disk.IOCountersStat
	lastNetIO  map[string]net.IOCountersStat
	lastTime   time.Time
}

func NewMonitorService() *MonitorService {
	return &MonitorService{
		lastDiskIO: make(map[string]disk.IOCountersStat),
		lastNetIO:  make(map[string]net.IOCountersStat),
		lastTime:   time.Now(),
	}
}

func (s *MonitorService) QueryData(req dto.MonitorQueryReq) ([]interface{}, error) {
	var result []interface{}

	if req.Param == "all" || req.Param == "load" || req.Param == "cpu" || req.Param == "memory" {
		var bases []models.MonitorBase
		err := global.DB.Where("created_at BETWEEN ? AND ?", req.StartTime, req.EndTime).
			Order("created_at ASC").Find(&bases).Error
		if err != nil {
			return nil, err
		}

		data := make([]dto.MonitorBaseData, len(bases))
		for i, base := range bases {
			data[i] = dto.MonitorBaseData{
				Date:      base.CreatedAt,
				CPU:       base.CPU,
				Memory:    base.Memory,
				Load1:     base.Load1,
				Load5:     base.Load5,
				Load15:    base.Load15,
				LoadUsage: base.LoadUsage,
			}
			if req.Param == "all" || req.Param == "cpu" || req.Param == "load" {
				json.Unmarshal([]byte(base.TopCPU), &data[i].TopCPU)
			}
			if req.Param == "all" || req.Param == "memory" {
				json.Unmarshal([]byte(base.TopMem), &data[i].TopMem)
			}
		}
		result = append(result, data)
	}

	if req.Param == "all" || req.Param == "io" {
		var ios []models.MonitorIO
		query := global.DB.Where("created_at BETWEEN ? AND ?", req.StartTime, req.EndTime)
		if req.IO != "" && req.IO != "all" {
			query = query.Where("name = ?", req.IO)
		}
		err := query.Order("created_at ASC").Find(&ios).Error
		if err != nil {
			return nil, err
		}

		data := make([]dto.MonitorIOData, len(ios))
		for i, io := range ios {
			data[i] = dto.MonitorIOData{
				Date:  io.CreatedAt,
				Read:  io.Read,
				Write: io.Write,
				Count: io.Count,
				Time:  io.Time,
			}
		}
		result = append(result, data)
	}

	if req.Param == "all" || req.Param == "network" {
		var nets []models.MonitorNetwork
		query := global.DB.Where("created_at BETWEEN ? AND ?", req.StartTime, req.EndTime)
		if req.Network != "" && req.Network != "all" {
			query = query.Where("name = ?", req.Network)
		}
		err := query.Order("created_at ASC").Find(&nets).Error
		if err != nil {
			return nil, err
		}

		data := make([]dto.MonitorNetworkData, len(nets))
		for i, n := range nets {
			data[i] = dto.MonitorNetworkData{
				Date: n.CreatedAt,
				Up:   n.Up,
				Down: n.Down,
			}
		}
		result = append(result, data)
	}

	return result, nil
}

func (s *MonitorService) GetIOOptions() ([]string, error) {
	var names []string
	err := global.DB.Model(&models.MonitorIO{}).Distinct("name").Pluck("name", &names).Error
	return names, err
}

func (s *MonitorService) GetNetworkOptions() ([]string, error) {
	var names []string
	err := global.DB.Model(&models.MonitorNetwork{}).Distinct("name").Pluck("name", &names).Error
	return names, err
}

func (s *MonitorService) getTopProcesses(limit int) ([]dto.Process, []dto.Process, error) {
	procs, _ := process.Processes()
	var cpuProcs, memProcs []dto.Process

	for _, p := range procs {
		if name, _ := p.Name(); name != "" {
			if cpuPct, _ := p.CPUPercent(); cpuPct > 0 {
				if user, _ := p.Username(); user != "" {
					if memInfo, _ := p.MemoryInfo(); memInfo != nil {
						cpuProcs = append(cpuProcs, dto.Process{
							Name:    name,
							Pid:     p.Pid,
							Percent: cpuPct,
							Memory:  memInfo.RSS,
							User:    user,
						})
					}
				}
			}
			if memInfo, _ := p.MemoryInfo(); memInfo != nil && memInfo.RSS > 0 {
				if user, _ := p.Username(); user != "" {
					if memPct, _ := p.MemoryPercent(); memPct > 0 {
						memProcs = append(memProcs, dto.Process{
							Name:    name,
							Pid:     p.Pid,
							Percent: float64(memPct),
							Memory:  memInfo.RSS,
							User:    user,
						})
					}
				}
			}
		}
	}

	sort.Slice(cpuProcs, func(i, j int) bool { return cpuProcs[i].Percent > cpuProcs[j].Percent })
	sort.Slice(memProcs, func(i, j int) bool { return memProcs[i].Memory > memProcs[j].Memory })

	if len(cpuProcs) > limit {
		cpuProcs = cpuProcs[:limit]
	}
	if len(memProcs) > limit {
		memProcs = memProcs[:limit]
	}

	return cpuProcs, memProcs, nil
}

func (s *MonitorService) CollectData() error {
	now := time.Now()
	interval := now.Sub(s.lastTime).Seconds()
	s.lastTime = now

	cpuPct, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	loadInfo, _ := load.Avg()
	cpuCount, _ := cpu.Counts(true)

	topCPU, topMem, _ := s.getTopProcesses(5)
	topCPUJson, _ := json.Marshal(topCPU)
	topMemJson, _ := json.Marshal(topMem)

	base := models.MonitorBase{
		CPU:       cpuPct[0],
		Memory:    memInfo.UsedPercent,
		Load1:     loadInfo.Load1,
		Load5:     loadInfo.Load5,
		Load15:    loadInfo.Load15,
		LoadUsage: (loadInfo.Load1 / float64(cpuCount)) * 100,
		TopCPU:    string(topCPUJson),
		TopMem:    string(topMemJson),
	}
	if err := global.DB.Create(&base).Error; err != nil {
		return err
	}

	diskStats, _ := disk.IOCounters()
	for name, stat := range diskStats {
		if last, ok := s.lastDiskIO[name]; ok && interval > 0 {
			io := models.MonitorIO{
				Name:  name,
				Read:  uint64(float64(stat.ReadBytes-last.ReadBytes) / interval),
				Write: uint64(float64(stat.WriteBytes-last.WriteBytes) / interval),
				Count: uint64(float64(stat.ReadCount+stat.WriteCount-last.ReadCount-last.WriteCount) / interval),
				Time:  uint64(float64(stat.ReadTime+stat.WriteTime-last.ReadTime-last.WriteTime) / interval),
			}
			global.DB.Create(&io)
		}
		s.lastDiskIO[name] = stat
	}

	netStats, _ := net.IOCounters(true)
	for _, stat := range netStats {
		if last, ok := s.lastNetIO[stat.Name]; ok && interval > 0 {
			netData := models.MonitorNetwork{
				Name: stat.Name,
				Up:   float64(stat.BytesSent-last.BytesSent) / interval / 1024,
				Down: float64(stat.BytesRecv-last.BytesRecv) / interval / 1024,
			}
			global.DB.Create(&netData)
		}
		s.lastNetIO[stat.Name] = stat
	}

	return nil
}
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
		DefaultNetwork:  setting.DefaultNetwork,
		DefaultIO:       setting.DefaultIO,
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
	if req.DefaultNetwork != nil {
		setting.DefaultNetwork = *req.DefaultNetwork
	}
	if req.DefaultIO != nil {
		setting.DefaultIO = *req.DefaultIO
	}

	return global.DB.Save(&setting).Error
}

func (s *MonitorService) CleanOldData(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	global.DB.Where("created_at < ?", cutoff).Delete(&models.MonitorBase{})
	global.DB.Where("created_at < ?", cutoff).Delete(&models.MonitorIO{})
	global.DB.Where("created_at < ?", cutoff).Delete(&models.MonitorNetwork{})
	return nil
}
