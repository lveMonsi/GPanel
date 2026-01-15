package models

type SystemInfo struct {
	Hostname         string       `json:"hostname"`
	OS               string       `json:"os"`
	Platform         string       `json:"platform"`
	PlatformFamily   string       `json:"platformFamily"`
	PlatformVersion  string       `json:"platformVersion"`
	KernelArch       string       `json:"kernelArch"`
	KernelVersion    string       `json:"kernelVersion"`
	BootTime         uint64       `json:"bootTime"`
	Uptime           uint64       `json:"uptime"`
	Procs            uint64       `json:"procs"`
	HostAddress      string       `json:"hostAddress"`
	CurrentInfo      CurrentInfo  `json:"currentInfo"`
}

type CurrentInfo struct {
	CPUInfo          CPUInfo      `json:"cpuInfo"`
	MemoryInfo       MemoryInfo   `json:"memoryInfo"`
	DiskInfo         []DiskInfo   `json:"diskInfo"`
	LoadInfo         LoadInfo     `json:"loadInfo"`
	NetworkInfo      NetworkInfo  `json:"networkInfo"`
}

type CPUInfo struct {
	Cores           int       `json:"cores"`
	LogicalCores    int       `json:"logicalCores"`
	ModelName       string    `json:"modelName"`
	Mhz             float64   `json:"mhz"`
	UsedPercent     float64   `json:"usedPercent"`
	PerCorePercent  []float64 `json:"perCorePercent"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"usedPercent"`
	Cached      uint64  `json:"cached"`
	Buffers     uint64  `json:"buffers"`
}

type DiskInfo struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Fstype     string  `json:"fstype"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type LoadInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type NetworkInfo struct {
	BytesSent uint64 `json:"bytesSent"`
	BytesRecv uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
}