package utils

import (
	"gpanel/models"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func GetSystemInfo() (*models.SystemInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	processes, _ := process.Processes()
	procs := uint64(len(processes))

	cpuInfo, memInfo, diskInfo, loadInfo, networkInfo, err := GetCurrentInfo()
	if err != nil {
		return nil, err
	}

	return &models.SystemInfo{
		Hostname:        hostInfo.Hostname,
		OS:              hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformFamily:  hostInfo.PlatformFamily,
		PlatformVersion: hostInfo.PlatformVersion,
		KernelArch:      hostInfo.KernelArch,
		KernelVersion:   hostInfo.KernelVersion,
		BootTime:        hostInfo.BootTime,
		Uptime:          hostInfo.Uptime,
		Procs:           procs,
		CurrentInfo: models.CurrentInfo{
			CPUInfo:     cpuInfo,
			MemoryInfo:  memInfo,
			DiskInfo:    diskInfo,
			LoadInfo:    loadInfo,
			NetworkInfo: networkInfo,
		},
	}, nil
}

func GetCurrentInfo() (models.CPUInfo, models.MemoryInfo, []models.DiskInfo, models.LoadInfo, models.NetworkInfo, error) {
	cpuInfo, err := getCPUInfo()
	if err != nil {
		return models.CPUInfo{}, models.MemoryInfo{}, nil, models.LoadInfo{}, models.NetworkInfo{}, err
	}

	memInfo, err := getMemoryInfo()
	if err != nil {
		return models.CPUInfo{}, models.MemoryInfo{}, nil, models.LoadInfo{}, models.NetworkInfo{}, err
	}

	diskInfo, err := getDiskInfo()
	if err != nil {
		return models.CPUInfo{}, models.MemoryInfo{}, nil, models.LoadInfo{}, models.NetworkInfo{}, err
	}

	loadInfo, err := getLoadInfo()
	if err != nil {
		return models.CPUInfo{}, models.MemoryInfo{}, nil, models.LoadInfo{}, models.NetworkInfo{}, err
	}

	networkInfo, err := getNetworkInfo()
	if err != nil {
		return models.CPUInfo{}, models.MemoryInfo{}, nil, models.LoadInfo{}, models.NetworkInfo{}, err
	}

	return cpuInfo, memInfo, diskInfo, loadInfo, networkInfo, nil
}

func getCPUInfo() (models.CPUInfo, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return models.CPUInfo{}, err
	}

	perCore, err := cpu.Percent(time.Second, true)
	if err != nil {
		return models.CPUInfo{}, err
	}

	cores, err := cpu.Counts(false)
	if err != nil {
		return models.CPUInfo{}, err
	}

	logicalCores, err := cpu.Counts(true)
	if err != nil {
		return models.CPUInfo{}, err
	}

	info, err := cpu.Info()
	if err != nil {
		return models.CPUInfo{}, err
	}

	modelName := ""
	mhz := 0.0
	if len(info) > 0 {
		modelName = info[0].ModelName
		mhz = info[0].Mhz
	}

	usedPercent := 0.0
	if len(percent) > 0 {
		usedPercent = percent[0]
	}

	return models.CPUInfo{
		Cores:          cores,
		LogicalCores:   logicalCores,
		ModelName:      modelName,
		Mhz:            mhz,
		UsedPercent:    usedPercent,
		PerCorePercent: perCore,
	}, nil
}

func getMemoryInfo() (models.MemoryInfo, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return models.MemoryInfo{}, err
	}

	return models.MemoryInfo{
		Total:       vmem.Total,
		Used:        vmem.Used,
		Free:        vmem.Free,
		Available:   vmem.Available,
		UsedPercent: vmem.UsedPercent,
		Cached:      vmem.Cached,
		Buffers:     vmem.Buffers,
	}, nil
}

func getDiskInfo() ([]models.DiskInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var diskInfos []models.DiskInfo
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskInfos = append(diskInfos, models.DiskInfo{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	return diskInfos, nil
}

func getLoadInfo() (models.LoadInfo, error) {
	avg, err := load.Avg()
	if err != nil {
		return models.LoadInfo{}, err
	}

	return models.LoadInfo{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}

func getNetworkInfo() (models.NetworkInfo, error) {
	counters, err := net.IOCounters(false)
	if err != nil {
		return models.NetworkInfo{}, err
	}

	if len(counters) == 0 {
		return models.NetworkInfo{}, nil
	}

	return models.NetworkInfo{
		BytesSent:   counters[0].BytesSent,
		BytesRecv:   counters[0].BytesRecv,
		PacketsSent: counters[0].PacketsSent,
		PacketsRecv: counters[0].PacketsRecv,
	}, nil
}

func GetOSInfo() string {
	return runtime.GOOS
}

func GetArchInfo() string {
	return runtime.GOARCH
}