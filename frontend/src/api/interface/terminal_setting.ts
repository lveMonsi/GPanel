export interface TerminalInfo {
  lineHeight: string;
  letterSpacing: string;
  fontSize: string;
  cursorBlink: string;
  cursorStyle: string;
  scrollback: string;
  scrollSensitivity: string;
}

export interface TerminalUpdate {
  lineHeight?: string;
  letterSpacing?: string;
  fontSize?: string;
  cursorBlink?: string;
  cursorStyle?: string;
  scrollback?: string;
  scrollSensitivity?: string;
}