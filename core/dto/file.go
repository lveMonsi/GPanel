package dto

import "time"

// FileListReq 文件列表请求
type FileListReq struct {
	Path       string `json:"path" binding:"required"`
	Search     string `json:"search"`
	ShowHidden bool   `json:"showHidden"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortBy     string `json:"sortBy"`     // name, size, modTime
	SortOrder  string `json:"sortOrder"`  // ascending, descending
}

// FileListRes 文件列表响应
type FileListRes struct {
	Path      string      `json:"path"`
	Name      string      `json:"name"`
	Items     []FileInfo  `json:"items"`
	ItemTotal int         `json:"itemTotal"`
}

// FileInfo 文件信息
type FileInfo struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	IsDir      bool      `json:"isDir"`
	IsSymlink  bool      `json:"isSymlink"`
	IsHidden   bool      `json:"isHidden"`
	LinkPath   string    `json:"linkPath"`
	Extension  string    `json:"extension"`
	Mode       string    `json:"mode"`
	MimeType   string    `json:"mimeType"`
	ModTime    time.Time `json:"modTime"`
	User       string    `json:"user"`
	Group      string    `json:"group"`
	Uid        string    `json:"uid"`
	Gid        string    `json:"gid"`
}

// FileCreateReq 创建文件/目录请求
type FileCreateReq struct {
	Path   string `json:"path" binding:"required"`
	IsDir  bool   `json:"isDir"`
	Mode   int64  `json:"mode"`
}

// FileDeleteReq 删除文件/目录请求
type FileDeleteReq struct {
	Path      string `json:"path" binding:"required"`
	Force     bool   `json:"force"`
}

// FileRenameReq 重命名文件/目录请求
type FileRenameReq struct {
	OldPath string `json:"oldPath" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
}

// FileMoveReq 移动文件/目录请求
type FileMoveReq struct {
	OldPaths []string `json:"oldPaths" binding:"required"`
	NewPath  string   `json:"newPath" binding:"required"`
	Type     string   `json:"type"` // cut, copy
	Cover    bool     `json:"cover"`
}

// FileContentReq 获取文件内容请求
type FileContentReq struct {
	Path string `json:"path" binding:"required"`
}

// FileContentRes 文件内容响应
type FileContentRes struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Mode      string `json:"mode"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mimeType"`
	Truncated bool   `json:"truncated"`  // 文件是否被截断（超过大小限制）
}

// FileEditReq 编辑文件内容请求
type FileEditReq struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// DirSizeReq 获取目录大小请求
type DirSizeReq struct {
	Path string `json:"path" binding:"required"`
}

// DirSizeRes 目录大小响应
type DirSizeRes struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// FileCompressReq 压缩文件请求
type FileCompressReq struct {
	Files []string `json:"files" binding:"required"`
	Dst   string   `json:"dst" binding:"required"`   // 目标目录
	Name  string   `json:"name" binding:"required"`  // 压缩文件名
	Type  string   `json:"type"`                    // zip, tar.gz
}

// FileDecompressReq 解压文件请求
type FileDecompressReq struct {
	Path string `json:"path" binding:"required"` // 压缩文件路径
	Dst  string `json:"dst" binding:"required"`  // 解压目标目录
	Type string `json:"type"`                   // zip, tar.gz
}

// FileChmodReq 修改文件权限请求
type FileChmodReq struct {
	Path string `json:"path" binding:"required"`
	Mode int64  `json:"mode" binding:"required"` // 权限模式，如 0755
	Sub  bool   `json:"sub"`                    // 是否递归修改子文件
}

// FileChownReq 修改文件所有者请求
type FileChownReq struct {
	Path  string `json:"path" binding:"required"`
	User  string `json:"user"`  // 用户名
	Group string `json:"group"` // 组名
	Sub   bool   `json:"sub"`   // 是否递归修改子文件
}

// ProgressInfo 进度信息
type ProgressInfo struct {
	Key      string  `json:"key"`       // 进度唯一标识
	Name     string  `json:"name"`      // 操作名称
	Total    int64   `json:"total"`     // 总大小/总数
	Current  int64   `json:"current"`   // 当前进度
	Percent float64 `json:"percent"`   // 百分比
	Status   string  `json:"status"`    // 状态: pending, running, completed, failed
	Message  string  `json:"message"`   // 消息
}

// ProgressReq 进度查询请求
type ProgressReq struct {
	Key string `json:"key" binding:"required"`
}

// FilePreviewReq 文件预览请求
type FilePreviewReq struct {
	Path string `json:"path" binding:"required"`
}

// FilePreviewRes 文件预览响应
type FilePreviewRes struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Type     string `json:"type"`     // text, image, video, audio, binary
	Content  string `json:"content"`  // 文本内容或 base64 编码内容
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}