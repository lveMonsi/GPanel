package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

func NormalizePath(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		if len(path) >= 2 && path[1] == ':' {
			path = strings.ToUpper(string(path[0])) + path[1:]
		}
		path = filepath.ToSlash(path)
	}
	return path
}

func IsAbsPath(path string) bool {
	return filepath.IsAbs(path)
}

func JoinPath(parts ...string) string {
	return filepath.Join(parts...)
}

func BasePath(path string) string {
	return filepath.Base(path)
}

func DirPath(path string) string {
	return filepath.Dir(path)
}

var protectedPaths = []string{"/", "/bin", "/sbin", "/etc", "/boot", "/usr", "/lib", "/lib64", "/dev", "/proc", "/sys", "/root"}

func IsProtected(path string) bool {
	real, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = real
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	for _, p := range protectedPaths {
		if path == p {
			return true
		}
	}
	return false
}

type FileOp struct {
	Fs afero.Fs
}

func NewFileOp() FileOp {
	return FileOp{Fs: afero.NewOsFs()}
}

func (f FileOp) Stat(path string) bool {
	info, _ := f.Fs.Stat(path)
	return info != nil
}

func (f FileOp) CreateDir(path string, mode os.FileMode) error {
	return f.Fs.MkdirAll(path, mode)
}

func (f FileOp) CreateFile(path string, mode os.FileMode) error {
	file, err := f.Fs.OpenFile(path, os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	return file.Close()
}

func (f FileOp) DeleteFile(path string) error {
	if IsProtected(path) {
		return fmt.Errorf("path %s is protected and cannot be deleted", path)
	}
	return f.Fs.Remove(path)
}

func (f FileOp) DeleteDir(path string) error {
	if IsProtected(path) {
		return fmt.Errorf("path %s is protected and cannot be deleted", path)
	}
	return f.Fs.RemoveAll(path)
}

func (f FileOp) Rename(oldPath, newPath string) error {
	return f.Fs.Rename(oldPath, newPath)
}

func (f FileOp) WriteFile(path string, content string, mode os.FileMode) error {
	if !f.Stat(filepath.Dir(path)) {
		_ = f.CreateDir(filepath.Dir(path), mode.Perm())
	}
	file, err := f.Fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

func (f FileOp) ReadFile(path string) ([]byte, error) {
	afs := &afero.Afero{Fs: f.Fs}
	return afs.ReadFile(path)
}

func (f FileOp) CopyFile(src, dst string) error {
	srcFile, err := f.Fs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := f.Fs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (f FileOp) CopyDir(src, dst string) error {
	srcInfo, err := f.Fs.Stat(src)
	if err != nil {
		return err
	}
	if err := f.Fs.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := afero.ReadDir(f.Fs, src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := f.CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else if err := f.CopyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func (f FileOp) GetDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func (f FileOp) Chmod(path string, mode os.FileMode) error {
	return f.Fs.Chmod(path, mode)
}

func GetFileMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0644
	}
	return info.Mode()
}

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func IsHidden(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}

func GetSymlink(path string) string {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return linkPath
}

func GetUsername(uid uint32) string {
	if runtime.GOOS == "windows" {
		if currentUser, err := user.Current(); err == nil {
			return currentUser.Username
		}
		return "unknown"
	}
	u, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		return strconv.Itoa(int(uid))
	}
	return u.Username
}

func GetGroup(gid uint32) string {
	if runtime.GOOS == "windows" {
		if currentUser, err := user.Current(); err == nil {
			return currentUser.Gid
		}
		return "unknown"
	}
	g, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err != nil {
		return strconv.Itoa(int(gid))
	}
	return g.Name
}

func GetFileOwner(path string) (username, groupname string, uid, gid uint32) {
	info, err := os.Stat(path)
	if err != nil {
		return "-", "-", 0, 0
	}
	if runtime.GOOS == "windows" {
		if currentUser, err := user.Current(); err == nil {
			uidStr, _ := strconv.Atoi(currentUser.Uid)
			gidStr, _ := strconv.Atoi(currentUser.Gid)
			return currentUser.Username, currentUser.Gid, uint32(uidStr), uint32(gidStr)
		}
		return "-", "-", 0, 0
	}
	sys := info.Sys()
	if sys == nil {
		return "-", "-", 0, 0
	}
	v := reflect.ValueOf(sys)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "-", "-", 0, 0
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "-", "-", 0, 0
	}
	uidField := v.FieldByName("Uid")
	gidField := v.FieldByName("Gid")
	if !uidField.IsValid() || !gidField.IsValid() {
		return "-", "-", 0, 0
	}
	if !uidField.CanUint() || !gidField.CanUint() {
		return "-", "-", 0, 0
	}
	uid = uint32(uidField.Uint())
	gid = uint32(gidField.Uint())
	return GetUsername(uid), GetGroup(gid), uid, gid
}

func CleanPath(path string) string {
	return filepath.Clean(path)
}

func IsPathValid(path string) bool {
	cleanPath := CleanPath(path)
	return !strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath)
}

func GetMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".txt": "text/plain", ".html": "text/html", ".css": "text/css", ".js": "application/javascript", ".json": "application/json",
		".xml": "application/xml", ".pdf": "application/pdf", ".zip": "application/zip", ".tar": "application/x-tar", ".gz": "application/gzip",
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".gif": "image/gif", ".svg": "image/svg+xml",
		".mp3": "audio/mpeg", ".mp4": "video/mp4", ".avi": "video/x-msvideo", ".mov": "video/quicktime", ".wmv": "video/x-ms-wmv",
		".flv": "video/x-flv", ".mkv": "video/x-matroska", ".doc": "application/msword", ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls": "application/vnd.ms-excel", ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".ppt": "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation", ".odt": "application/vnd.oasis.opendocument.text",
		".ods": "application/vnd.oasis.opendocument.spreadsheet", ".odp": "application/vnd.oasis.opendocument.presentation",
	}
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

var textExtensions = map[string]bool{
	".txt": true, ".md": true, ".markdown": true, ".json": true, ".xml": true, ".html": true, ".htm": true, ".css": true,
	".scss": true, ".sass": true, ".less": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true, ".vue": true,
	".svelte": true, ".py": true, ".rb": true, ".php": true, ".java": true, ".kt": true, ".kts": true, ".go": true,
	".rs": true, ".c": true, ".h": true, ".cpp": true, ".hpp": true, ".cc": true, ".cxx": true, ".cs": true,
	".swift": true, ".m": true, ".mm": true, ".sh": true, ".bash": true, ".zsh": true, ".fish": true, ".ps1": true,
	".psm1": true, ".bat": true, ".cmd": true, ".sql": true, ".yaml": true, ".yml": true, ".toml": true, ".ini": true,
	".conf": true, ".cfg": true, ".config": true, ".env": true, ".gitignore": true, ".dockerignore": true, ".editorconfig": true,
	".eslintrc": true, ".prettierrc": true, ".babelrc": true, ".log": true, ".csv": true, ".tsv": true, ".lua": true, ".r": true,
	".pl": true, ".pm": true, ".scala": true, ".groovy": true, ".gradle": true, ".mvn": true, ".properties": true, ".tf": true,
	".hcl": true, ".nomad": true, ".rego": true, ".proto": true, ".thrift": true, ".avdl": true, ".plantuml": true, ".puml": true,
	".mermaid": true, ".dockerfile": true, ".makefile": true, ".rakefile": true, ".gemfile": true, ".pipfile": true, ".po": true,
	".pot": true, ".srt": true, ".vtt": true, ".ass": true, ".ssa": true,
}

func IsTextExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if textExtensions[ext] {
		return true
	}
	name := strings.ToLower(filepath.Base(filename))
	switch name {
	case "dockerfile", "makefile", "rakefile", "gemfile", "pipfile", "jenkinsfile", "vagrantfile", "brewfile", "podfile", "cartfile", "fastfile", "matchfile":
		return true
	}
	if strings.HasPrefix(name, ".") && !strings.Contains(ext, ".") {
		return true
	}
	return false
}

func DetectBinary(buf []byte) bool {
	if len(buf) == 0 {
		return false
	}
	n := 8192
	if len(buf) < n {
		n = len(buf)
	}
	if n >= 3 && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		return false
	}
	if n >= 2 && ((buf[0] == 0xFF && buf[1] == 0xFE) || (buf[0] == 0xFE && buf[1] == 0xFF)) {
		return false
	}
	nullCount, controlCount, textCount := 0, 0, 0
	for i := 0; i < n; i++ {
		b := buf[i]
		switch {
		case b == 0x00:
			nullCount++
		case b == 0x09 || b == 0x0A || b == 0x0D:
			textCount++
		case b >= 0x20 && b <= 0x7E:
			textCount++
		case b >= 0x80:
			textCount++
		case b <= 0x08 || (b >= 0x0B && b <= 0x0C) || (b >= 0x0E && b <= 0x1F):
			controlCount++
		}
	}
	if nullCount > 0 {
		return true
	}
	if float64(controlCount)/float64(n) > 0.3 {
		return true
	}
	return float64(textCount)/float64(n) < 0.85
}

func (f FileOp) ChmodR(path string, mode os.FileMode, recursive bool) error {
	if !f.Stat(path) {
		return fmt.Errorf("path not found: %s", path)
	}
	if err := f.Fs.Chmod(path, mode); err != nil {
		return err
	}
	if recursive {
		info, err := f.Fs.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if filePath != path {
					return f.Fs.Chmod(filePath, mode)
				}
				return nil
			})
		}
	}
	return nil
}

func (f FileOp) ChownR(path, userName, groupName string, recursive bool) error {
	if !f.Stat(path) {
		return fmt.Errorf("path not found: %s", path)
	}
	uid, gid, err := getUserAndGroupIDs(userName, groupName)
	if err != nil {
		return fmt.Errorf("failed to get user/group IDs: %v", err)
	}
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s: %v", path, err)
	}
	if recursive {
		info, err := f.Fs.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if filePath != path {
					return os.Chown(filePath, uid, gid)
				}
				return nil
			})
		}
	}
	return nil
}

func getUserAndGroupIDs(username, groupname string) (int, int, error) {
	uid, gid := -1, -1
	if username != "" {
		u, err := user.Lookup(username)
		if err != nil {
			return 0, 0, fmt.Errorf("user not found: %s", username)
		}
		uid, _ = strconv.Atoi(u.Uid)
		if groupname == "" {
			gid, _ = strconv.Atoi(u.Gid)
		}
	}
	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			return 0, 0, fmt.Errorf("group not found: %s", groupname)
		}
		gid, _ = strconv.Atoi(g.Gid)
	}
	return uid, gid, nil
}

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
