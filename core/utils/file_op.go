package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

// NormalizePath 规范化路径（跨平台）
func NormalizePath(path string) string {
	// 清理路径
	path = filepath.Clean(path)
	
	// 在 Windows 上，确保驱动器字母大写
	if runtime.GOOS == "windows" {
		if len(path) >= 2 && path[1] == ':' {
			path = strings.ToUpper(string(path[0])) + path[1:]
		}
		// 将反斜杠转换为正斜杠（用于显示）
		path = filepath.ToSlash(path)
	}
	
	return path
}

// IsAbsPath 检查是否为绝对路径（跨平台）
func IsAbsPath(path string) bool {
	return filepath.IsAbs(path)
}

// JoinPath 连接路径（跨平台）
func JoinPath(parts ...string) string {
	return filepath.Join(parts...)
}

// BasePath 获取路径的最后一部分（跨平台）
func BasePath(path string) string {
	return filepath.Base(path)
}

// DirPath 获取路径的目录部分（跨平台）
func DirPath(path string) string {
	return filepath.Dir(path)
}

var protectedPaths = []string{
	"/",
	"/bin",
	"/sbin",
	"/etc",
	"/boot",
	"/usr",
	"/lib",
	"/lib64",
	"/dev",
	"/proc",
	"/sys",
	"/root",
}

// IsProtected 检查路径是否受保护
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

// FileOp 文件操作工具
type FileOp struct {
	Fs afero.Fs
}

// NewFileOp 创建文件操作工具
func NewFileOp() FileOp {
	return FileOp{
		Fs: afero.NewOsFs(),
	}
}

// Stat 检查文件是否存在
func (f FileOp) Stat(path string) bool {
	info, _ := f.Fs.Stat(path)
	return info != nil
}

// CreateDir 创建目录
func (f FileOp) CreateDir(path string, mode os.FileMode) error {
	return f.Fs.MkdirAll(path, mode)
}

// CreateFile 创建文件
func (f FileOp) CreateFile(path string, mode os.FileMode) error {
	file, err := f.Fs.OpenFile(path, os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	return file.Close()
}

// DeleteFile 删除文件
func (f FileOp) DeleteFile(path string) error {
	if IsProtected(path) {
		return fmt.Errorf("path %s is protected and cannot be deleted", path)
	}
	return f.Fs.Remove(path)
}

// DeleteDir 删除目录
func (f FileOp) DeleteDir(path string) error {
	if IsProtected(path) {
		return fmt.Errorf("path %s is protected and cannot be deleted", path)
	}
	return f.Fs.RemoveAll(path)
}

// Rename 重命名文件或目录
func (f FileOp) Rename(oldPath, newPath string) error {
	return f.Fs.Rename(oldPath, newPath)
}

// WriteFile 写入文件
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

// ReadFile 读取文件
func (f FileOp) ReadFile(path string) ([]byte, error) {
	afs := &afero.Afero{Fs: f.Fs}
	return afs.ReadFile(path)
}

// CopyFile 复制文件
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

// CopyDir 复制目录
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
		} else {
			if err := f.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetDirSize 获取目录大小
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

// Chmod 修改文件权限
func (f FileOp) Chmod(path string, mode os.FileMode) error {
	return f.Fs.Chmod(path, mode)
}

// GetFileMode 获取文件权限模式
func GetFileMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0644
	}
	return info.Mode()
}

// IsSymlink 检查是否为符号链接
func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

// IsHidden 检查是否为隐藏文件
func IsHidden(path string) bool {
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".")
}

// GetSymlink 获取符号链接目标
func GetSymlink(path string) string {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return linkPath
}

// GetUsername 获取用户名（跨平台）
func GetUsername(uid uint32) string {
	if runtime.GOOS == "windows" {
		// Windows 上返回当前用户名
		if currentUser, err := user.Current(); err == nil {
			return currentUser.Username
		}
		return "unknown"
	}
	
	// Linux/Unix 上从 /etc/passwd 获取
	u, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		return strconv.Itoa(int(uid))
	}
	return u.Username
}

// GetGroup 获取组名（跨平台）
func GetGroup(gid uint32) string {
	if runtime.GOOS == "windows" {
		// Windows 上返回当前组名
		if currentUser, err := user.Current(); err == nil {
			return currentUser.Gid
		}
		return "unknown"
	}
	
	// Linux/Unix 上从 /etc/group 获取
	g, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err != nil {
		return strconv.Itoa(int(gid))
	}
	return g.Name
}

// GetFileOwner 获取文件所有者信息
func GetFileOwner(path string) (username, groupname string, uid, gid uint32) {
	info, err := os.Stat(path)
	if err != nil {
		return "-", "-", 0, 0
	}

	if runtime.GOOS == "windows" {
		// Windows 上返回当前用户信息
		if currentUser, err := user.Current(); err == nil {
			uidStr, _ := strconv.Atoi(currentUser.Uid)
			gidStr, _ := strconv.Atoi(currentUser.Gid)
			return currentUser.Username, currentUser.Gid, uint32(uidStr), uint32(gidStr)
		}
		return "-", "-", 0, 0
	}

	// Linux/Unix 上使用 syscall.Stat_t
	type statT interface {
		Uid() uint32
		Gid() uint32
	}
	
	if stat, ok := info.Sys().(statT); ok {
		return GetUsername(stat.Uid()), GetGroup(stat.Gid()), stat.Uid(), stat.Gid()
	}

	return "-", "-", 0, 0
}

// CleanPath 清理路径
func CleanPath(path string) string {
	return filepath.Clean(path)
}

// IsPathValid 检查路径是否有效
func IsPathValid(path string) bool {
	cleanPath := CleanPath(path)
	return !strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath)
}

// GetMimeType 获取文件 MIME 类型
func GetMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".mkv":  "video/x-matroska",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".odt":  "application/vnd.oasis.opendocument.text",
		".ods":  "application/vnd.oasis.opendocument.spreadsheet",
		".odp":  "application/vnd.oasis.opendocument.presentation",
	}
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// DetectBinary 检测是否为二进制文件
func DetectBinary(buf []byte) bool {
	if len(buf) == 0 {
		return false
	}
	
	// 检查前 1024 字节
	n := 1024
	if len(buf) < n {
		n = len(buf)
	}
	
	whiteByte := 0
	for i := 0; i < n; i++ {
		if (buf[i] >= 0x20 && buf[i] <= 0x7E) || buf[i] == 9 || buf[i] == 10 || buf[i] == 13 {
			whiteByte++
		} else if buf[i] <= 6 || (buf[i] >= 14 && buf[i] <= 31) {
			return true
		}
	}
	
	// 如果可打印字符占比小于 80%，认为是二进制文件
	return float64(whiteByte)/float64(n) < 0.8
}

// ChmodR 递归修改文件权限
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

// ChownR 递归修改文件所有者
func (f FileOp) ChownR(path, user, group string, recursive bool) error {
	if !f.Stat(path) {
		return fmt.Errorf("path not found: %s", path)
	}

	// 获取用户和组的 UID/GID
	uid, gid, err := getUserAndGroupIDs(user, group)
	if err != nil {
		return fmt.Errorf("failed to get user/group IDs: %v", err)
	}

	// 修改当前文件/目录的所有者
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s: %v", path, err)
	}

	// 如果需要递归修改
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

// getUserAndGroupIDs 获取用户和组的 UID/GID
func getUserAndGroupIDs(username, groupname string) (int, int, error) {
	uid := -1
	gid := -1
	var err error

	// 获取用户 ID
	if username != "" {
		u, err := user.Lookup(username)
		if err != nil {
			return 0, 0, fmt.Errorf("user not found: %s", username)
		}
		uidInt, _ := strconv.Atoi(u.Uid)
		uid = uidInt
	}

	// 获取组 ID
	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			return 0, 0, fmt.Errorf("group not found: %s", groupname)
		}
		gidInt, _ := strconv.Atoi(g.Gid)
		gid = gidInt
	}

	// 如果只提供了用户名，使用用户的主组
	if username != "" && groupname == "" {
		u, err := user.Lookup(username)
		if err != nil {
			return 0, 0, fmt.Errorf("user not found: %s", username)
		}
		gidInt, _ := strconv.Atoi(u.Gid)
		gid = gidInt
	}

	return uid, gid, err
}

// Contains 检查字符串是否在数组中
func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// Base64Encode 将字节数组编码为 base64 字符串
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}