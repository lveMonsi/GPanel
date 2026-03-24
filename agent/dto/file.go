package dto

import "time"

type FileListReq struct {
	Path       string `json:"path" binding:"required"`
	Search     string `json:"search"`
	ShowHidden bool   `json:"showHidden"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortBy     string `json:"sortBy"`
	SortOrder  string `json:"sortOrder"`
}

type FileListRes struct {
	Path      string     `json:"path"`
	Name      string     `json:"name"`
	Items     []FileInfo `json:"items"`
	ItemTotal int        `json:"itemTotal"`
}

type FileInfo struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	IsDir     bool      `json:"isDir"`
	IsSymlink bool      `json:"isSymlink"`
	IsHidden  bool      `json:"isHidden"`
	LinkPath  string    `json:"linkPath"`
	Extension string    `json:"extension"`
	Mode      string    `json:"mode"`
	MimeType  string    `json:"mimeType"`
	ModTime   time.Time `json:"modTime"`
	User      string    `json:"user"`
	Group     string    `json:"group"`
	Uid       string    `json:"uid"`
	Gid       string    `json:"gid"`
}

type FileCreateReq struct {
	Path  string `json:"path" binding:"required"`
	IsDir bool   `json:"isDir"`
	Mode  int64  `json:"mode"`
}

type FileDeleteReq struct {
	Path  string `json:"path" binding:"required"`
	Force bool   `json:"force"`
}

type FileRenameReq struct {
	OldPath string `json:"oldPath" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
}

type FileMoveReq struct {
	OldPaths []string `json:"oldPaths" binding:"required"`
	NewPath  string   `json:"newPath" binding:"required"`
	Type     string   `json:"type"`
	Cover    bool     `json:"cover"`
}

type FileContentReq struct {
	Path string `json:"path" binding:"required"`
}

type FileContentRes struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Mode      string `json:"mode"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mimeType"`
	Truncated bool   `json:"truncated"`
}

type FileEditReq struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type DirSizeReq struct {
	Path string `json:"path" binding:"required"`
}

type DirSizeRes struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type FileCompressReq struct {
	Files []string `json:"files" binding:"required"`
	Dst   string   `json:"dst" binding:"required"`
	Name  string   `json:"name" binding:"required"`
	Type  string   `json:"type"`
}

type FileDecompressReq struct {
	Path string `json:"path" binding:"required"`
	Dst  string `json:"dst" binding:"required"`
	Type string `json:"type"`
}

type FileChmodReq struct {
	Path string `json:"path" binding:"required"`
	Mode int64  `json:"mode" binding:"required"`
	Sub  bool   `json:"sub"`
}

type FileChownReq struct {
	Path  string `json:"path" binding:"required"`
	User  string `json:"user"`
	Group string `json:"group"`
	Sub   bool   `json:"sub"`
}

type ProgressInfo struct {
	Key     string  `json:"key"`
	Name    string  `json:"name"`
	Total   int64   `json:"total"`
	Current int64   `json:"current"`
	Percent float64 `json:"percent"`
	Status  string  `json:"status"`
	Message string  `json:"message"`
}

type FilePreviewReq struct {
	Path string `json:"path" binding:"required"`
}

type FilePreviewRes struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

type RemoteDownloadReq struct {
	URL                string `json:"url" binding:"required"`
	Path               string `json:"path" binding:"required"`
	Name               string `json:"name" binding:"required"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

type RemoteDownloadRes struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}
