package dto

// TerminalInfo 终端设置
type TerminalInfo struct {
	LineHeight        string `json:"lineHeight"`
	LetterSpacing     string `json:"letterSpacing"`
	FontSize          string `json:"fontSize"`
	CursorBlink       string `json:"cursorBlink"`
	CursorStyle       string `json:"cursorStyle"`
	Scrollback        string `json:"scrollback"`
	ScrollSensitivity string `json:"scrollSensitivity"`
}

// TerminalUpdate 更新终端设置
type TerminalUpdate struct {
	LineHeight        string `json:"lineHeight"`
	LetterSpacing     string `json:"letterSpacing"`
	FontSize          string `json:"fontSize"`
	CursorBlink       string `json:"cursorBlink"`
	CursorStyle       string `json:"cursorStyle"`
	Scrollback        string `json:"scrollback"`
	ScrollSensitivity string `json:"scrollSensitivity"`
}