package global

import "runtime"

// These values are populated by release builds with -ldflags -X.
var (
	Version   = "dev"
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
	return BuildInfo{
		Version: Version, BuildTime: BuildTime, Commit: Commit,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH,
	}
}
