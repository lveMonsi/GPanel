package global

import (
	"bufio"
	"os"
	"runtime"
	"strings"
)

// These values are populated by release builds with -ldflags -X.
var (
	Version   = ""
	BuildTime = "unknown"
	Commit    = "unknown"
)

// BuildInfo describes the running core binary.
type BuildInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	Commit    string `json:"commit"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func GetBuildInfo() BuildInfo {
	version := Version
	if version == "" {
		version = installedVersion()
	}
	if version == "" {
		version = "unknown"
	}
	return BuildInfo{
		Version: version, BuildTime: BuildTime, Commit: Commit,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
	}
}

func installedVersion() string {
	file, err := os.Open("/opt/gpanel/.install_info")
	if err != nil {
		return ""
	}
	defer file.Close()
	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if ok && key == "version" && value != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
