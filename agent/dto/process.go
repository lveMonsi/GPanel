package dto

// ProcessSearchReq 进程搜索请求
type ProcessSearchReq struct {
	PID      int32  `json:"pid"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// NetSearchReq 网络连接搜索请求
type NetSearchReq struct {
	ProcessID   int32  `json:"processID"`
	ProcessName string `json:"processName"`
	Port        uint32 `json:"port"`
}

// ProcessStopReq 终止进程请求
type ProcessStopReq struct {
	PID int32 `json:"PID" binding:"required"`
}

// ProcessInfo 进程信息（列表项）
type ProcessInfo struct {
	PID            int32   `json:"PID"`
	Name           string  `json:"name"`
	PPID           int32   `json:"PPID"`
	Username       string  `json:"username"`
	Status         string  `json:"status"`
	StartTime      string  `json:"startTime"`
	NumThreads     int32   `json:"numThreads"`
	NumConnections int     `json:"numConnections"`
	CpuPercent     string  `json:"cpuPercent"`
	CpuValue       float64 `json:"cpuValue"`
	Rss            string  `json:"rss"`
	RssValue       uint64  `json:"rssValue"`
}

// ProcessDetail 进程详情
type ProcessDetail struct {
	PID            int32   `json:"PID"`
	Name           string  `json:"name"`
	PPID           int32   `json:"PPID"`
	Username       string  `json:"username"`
	Status         string  `json:"status"`
	StartTime      string  `json:"startTime"`
	NumThreads     int32   `json:"numThreads"`
	NumConnections int     `json:"numConnections"`
	CpuPercent     string  `json:"cpuPercent"`
	CpuValue       float64 `json:"cpuValue"`

	DiskRead  string `json:"diskRead"`
	DiskWrite string `json:"diskWrite"`
	CmdLine   string `json:"cmdLine"`

	Rss      string `json:"rss"`
	RssValue uint64 `json:"rssValue"`
	VMS      string `json:"vms"`
	HWM      string `json:"hwm"`
	Data     string `json:"data"`
	Stack    string `json:"stack"`
	Locked   string `json:"locked"`
	Swap     string `json:"swap"`
	Dirty    string `json:"dirty"`
	PSS      string `json:"pss"`
	USS      string `json:"uss"`
	Shared   string `json:"shared"`
	Text     string `json:"text"`

	Envs      []string        `json:"envs"`
	OpenFiles []OpenFileStat  `json:"openFiles"`
	Connects  []NetConnection `json:"connects"`
}

// OpenFileStat 打开文件信息
type OpenFileStat struct {
	Path string `json:"path"`
	Fd   uint64 `json:"fd"`
}

// NetConnection 网络连接信息
type NetConnection struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	LocalAddr  string `json:"localaddr"`
	RemoteAddr string `json:"remoteaddr"`
	PID        int32  `json:"PID"`
	Name       string `json:"name"`
}
