package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gpanel/agent/dto"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

const processTimeout = 10 * time.Second

type ProcessService struct{}

func NewProcessService() *ProcessService {
	return &ProcessService{}
}

// ListProcesses 获取进程列表
func (s *ProcessService) ListProcesses(req dto.ProcessSearchReq) ([]dto.ProcessInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	conns, err := net.ConnectionsMaxWithContext(ctx, "all", 32768)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}

	pidConns := make(map[int32]int, len(procs))
	for _, c := range conns {
		if c.Pid != 0 {
			pidConns[c.Pid]++
		}
	}

	result := make([]dto.ProcessInfo, 0, len(procs))
	for _, proc := range procs {
		info := buildProcessInfo(proc, &req, pidConns)
		if info != nil {
			result = append(result, *info)
		}
	}
	return result, nil
}

func buildProcessInfo(proc *process.Process, req *dto.ProcessSearchReq, pidConns map[int32]int) *dto.ProcessInfo {
	if req.PID > 0 && req.PID != proc.Pid {
		return nil
	}

	name, _ := proc.Name()
	if name == "" {
		name = "<UNKNOWN>"
	}
	if req.Name != "" && !strings.Contains(name, req.Name) {
		return nil
	}

	username, _ := proc.Username()
	if req.Username != "" && !strings.Contains(username, req.Username) {
		return nil
	}

	ppid, _ := proc.Ppid()
	statusArr, _ := proc.Status()
	status := ""
	if len(statusArr) > 0 {
		status = strings.Join(statusArr, ",")
	}

	var startTime string
	if ct, err := proc.CreateTime(); err == nil {
		startTime = time.Unix(ct/1000, 0).Format("2006-01-02 15:04:05")
	}

	numThreads, _ := proc.NumThreads()
	cpuVal, _ := proc.CPUPercent()

	var rssVal uint64
	var rssStr string
	if memInfo, err := proc.MemoryInfo(); err == nil {
		rssVal = memInfo.RSS
		rssStr = formatBytes(memInfo.RSS)
	}

	return &dto.ProcessInfo{
		PID:            proc.Pid,
		Name:           name,
		PPID:           ppid,
		Username:       username,
		Status:         status,
		StartTime:      startTime,
		NumThreads:     numThreads,
		NumConnections: pidConns[proc.Pid],
		CpuPercent:     fmt.Sprintf("%.2f%%", cpuVal),
		CpuValue:       cpuVal,
		Rss:            rssStr,
		RssValue:       rssVal,
	}
}

// GetProcessDetail 获取进程详情
func (s *ProcessService) GetProcessDetail(pid int32) (*dto.ProcessDetail, error) {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("process %d not found: %w", pid, err)
	}

	running, err := proc.IsRunning()
	if err != nil || !running {
		return nil, fmt.Errorf("process %d is not running", pid)
	}

	d := &dto.ProcessDetail{PID: pid}

	d.Name, _ = proc.Name()
	d.PPID, _ = proc.Ppid()
	d.Username, _ = proc.Username()

	if statusArr, err := proc.Status(); err == nil && len(statusArr) > 0 {
		d.Status = statusArr[0]
	}
	if ct, err := proc.CreateTime(); err == nil {
		d.StartTime = time.Unix(ct/1000, 0).Format("2006-01-02 15:04:05")
	}

	d.NumThreads, _ = proc.NumThreads()

	if connections, err := proc.Connections(); err == nil {
		d.NumConnections = len(connections)
		connects := make([]dto.NetConnection, 0, len(connections))
		for _, conn := range connections {
			connects = append(connects, dto.NetConnection{
				Status:     conn.Status,
				LocalAddr:  fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port),
				RemoteAddr: fmt.Sprintf("%s:%d", conn.Raddr.IP, conn.Raddr.Port),
				PID:        pid,
				Name:       d.Name,
			})
		}
		d.Connects = connects
	}

	if cpuPct, err := proc.CPUPercent(); err == nil {
		d.CpuValue = cpuPct
		d.CpuPercent = fmt.Sprintf("%.2f%%", cpuPct)
	}

	if io, err := proc.IOCounters(); err == nil {
		d.DiskRead = formatBytes(io.ReadBytes)
		d.DiskWrite = formatBytes(io.WriteBytes)
	}

	d.CmdLine, _ = proc.Cmdline()

	if mem, err := readMemoryDetail(pid); err == nil {
		d.Rss = formatBytes(mem.rss)
		d.RssValue = mem.rss
		d.VMS = formatBytes(mem.vms)
		d.HWM = formatBytes(mem.hwm)
		d.Data = formatBytes(mem.data)
		d.Stack = formatBytes(mem.stack)
		d.Locked = formatBytes(mem.locked)
		d.Swap = formatBytes(mem.swap)
		d.Dirty = formatBytes(mem.dirty)
		d.PSS = formatBytes(mem.pss)
		d.USS = formatBytes(mem.uss)
		d.Shared = formatBytes(mem.shared)
		d.Text = formatBytes(mem.text)
	}

	if envs, err := proc.Environ(); err == nil {
		d.Envs = envs
	}

	if files, err := proc.OpenFiles(); err == nil {
		openFiles := make([]dto.OpenFileStat, 0, len(files))
		for _, f := range files {
			openFiles = append(openFiles, dto.OpenFileStat{
				Path: f.Path,
				Fd:   f.Fd,
			})
		}
		d.OpenFiles = openFiles
	}

	return d, nil
}

// StopProcess 终止进程
func (s *ProcessService) StopProcess(pid int32) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d not found: %w", pid, err)
	}
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("kill process %d: %w", pid, err)
	}
	return nil
}

// ListNetConnections 获取网络连接列表
func (s *ProcessService) ListNetConnections(req dto.NetSearchReq) ([]dto.NetConnection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	conns, err := net.ConnectionsMaxWithContext(ctx, "all", 32768)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}

	pidNameCache := make(map[int32]string, 256)
	pidConnsMap := make(map[int32][]net.ConnectionStat, 256)

	for _, conn := range conns {
		if conn.Family != 2 && conn.Family != 10 {
			continue
		}
		if conn.Pid == 0 {
			continue
		}
		if req.ProcessID > 0 && conn.Pid != req.ProcessID {
			continue
		}
		if req.Port > 0 && conn.Laddr.Port != req.Port && conn.Raddr.Port != req.Port {
			continue
		}

		if _, ok := pidNameCache[conn.Pid]; !ok {
			pName := readProcessName(ctx, conn.Pid)
			if pName == "" {
				pName = "<UNKNOWN>"
			}
			pidNameCache[conn.Pid] = pName
		}

		pidConnsMap[conn.Pid] = append(pidConnsMap[conn.Pid], conn)
	}

	result := make([]dto.NetConnection, 0, 1024)
	for pid, connections := range pidConnsMap {
		pName := pidNameCache[pid]
		if req.ProcessName != "" && !strings.Contains(pName, req.ProcessName) {
			continue
		}
		for _, conn := range connections {
			result = append(result, dto.NetConnection{
				Type:       connTypeName(conn.Type, conn.Family),
				Status:     conn.Status,
				LocalAddr:  fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port),
				RemoteAddr: fmt.Sprintf("%s:%d", conn.Raddr.IP, conn.Raddr.Port),
				PID:        conn.Pid,
				Name:       pName,
			})
		}
	}

	return result, nil
}

func readProcessName(ctx context.Context, pid int32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return ""
	}
	name, _ := p.Name()
	return name
}

func connTypeName(connType uint32, family uint32) string {
	switch {
	case connType == 1 && family == 2:
		return "tcp"
	case connType == 1 && family == 10:
		return "tcp6"
	case connType == 2 && family == 2:
		return "udp"
	case connType == 2 && family == 10:
		return "udp6"
	default:
		return "unknown"
	}
}

// formatBytes 格式化字节数
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), units[exp])
}

// 内存详情
type memDetail struct {
	rss, vms, hwm, data, stack, locked, swap uint64
	pss, uss, shared, text, dirty            uint64
}

func readMemoryDetail(pid int32) (*memDetail, error) {
	mem := &memDetail{}
	if err := parseStatus(pid, mem); err != nil {
		return nil, err
	}
	if err := parseSmapsRollup(pid, mem); err != nil {
		_ = parseSmaps(pid, mem)
	}
	return mem, nil
}

func parseStatus(pid int32, mem *memDetail) error {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024

		switch key {
		case "VmRSS":
			mem.rss = val
		case "VmSize":
			mem.vms = val
		case "VmData":
			mem.data = val
		case "VmSwap":
			mem.swap = val
		case "VmExe":
			mem.text = val
		case "RssShmem":
			mem.shared = val
		case "VmHWM":
			mem.hwm = val
		case "VmStk":
			mem.stack = val
		case "VmLck":
			mem.locked = val
		}
	}
	return sc.Err()
}

func parseSmapsRollup(pid int32, mem *memDetail) error {
	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps_rollup", pid))
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024

		switch key {
		case "Pss":
			mem.pss = val
		case "Private_Clean", "Private_Dirty":
			mem.uss += val
		case "Shared_Clean", "Shared_Dirty":
			if mem.shared == 0 {
				mem.shared = val
			}
		}
	}
	return sc.Err()
}

func parseSmaps(pid int32, mem *memDetail) error {
	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024

		switch key {
		case "Pss":
			mem.pss += val
		case "Private_Clean", "Private_Dirty":
			mem.uss += val
		case "Shared_Clean", "Shared_Dirty":
			mem.shared += val
		}
	}
	return sc.Err()
}
