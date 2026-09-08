export type UpdatePhase = 'idle' | 'checking' | 'downloading' | 'validating' | 'staged' | 'restarting' | 'rolling_back' | 'completed' | 'failed'

export interface BuildInfo {
  version: string
  buildTime: string
  commit: string
  goVersion: string
  os: string
  arch: string
}

export interface UpdateStatus {
  state: 'idle' | 'running' | 'success' | 'failed'
  phase: UpdatePhase
  version?: string
  targetVersion?: string
  message?: string
  startedAt?: string
  finishedAt?: string
  error?: string
  output?: string
}

export interface UpdateRequest {
  version?: string
  prerelease?: boolean
  accelerated?: boolean
  force?: boolean
}

export interface ApiEnvelope<T> {
  code?: number
  message?: string
  status?: T
  build?: T
}
