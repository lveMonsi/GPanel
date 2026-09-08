export type UpdatePhase = 'idle' | 'checking' | 'downloading' | 'validating' | 'staged' | 'restarting' | 'rolling_back' | 'completed' | 'failed'
export type ReleaseChannel = 'stable' | 'prerelease'
export type CurrentReleaseChannel = ReleaseChannel | 'unknown'

export interface BuildInfo {
  version: string
  buildTime: string
  commit: string
  goVersion: string
  os: string
  arch: string
}

export interface LatestRelease {
  version: string
  channel: ReleaseChannel
  prerelease: boolean
  publishedAt: string
  htmlUrl?: string
}

export interface LatestVersionResponse {
  current: BuildInfo
  currentChannel: CurrentReleaseChannel
  channel: ReleaseChannel
  latest: LatestRelease | null
}

export interface UpdateStatus {
  state: 'idle' | 'running' | 'success' | 'failed'
  phase: UpdatePhase
  channel?: ReleaseChannel
  version?: string
  targetVersion?: string
  message?: string
  percent?: number
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
