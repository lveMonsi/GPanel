export interface Cronjob {
  id: number
  name: string
  type: CronjobType
  spec: string
  specCustom: boolean
  script: string
  url: string
  sourceDir: string
  exclusionRules: string
  retainCopies: number
  retryCount: number
  timeout: number
  ignoreErr: boolean
  status: 'enabled' | 'disabled'
  lastRecordStatus: string
  lastRecordTime: string
  createdAt: string
  updatedAt: string
}

export type CronjobType = 'shell' | 'curl' | 'directory' | 'clean' | 'cleanLog'

export interface CronjobCreate {
  name: string
  type: CronjobType
  spec: string
  specCustom: boolean
  script?: string
  url?: string
  sourceDir?: string
  exclusionRules?: string
  retainCopies?: number
  retryCount?: number
  timeout?: number
  ignoreErr?: boolean
}

export interface CronjobUpdate {
  id: number
  name?: string
  spec?: string
  specCustom?: boolean
  script?: string
  url?: string
  sourceDir?: string
  exclusionRules?: string
  retainCopies?: number
  retryCount?: number
  timeout?: number
  ignoreErr?: boolean
}

export interface CronjobDelete {
  ids: number[]
}

export interface CronjobSearch {
  page: number
  pageSize: number
  keyword?: string
  type?: string
  status?: string
}

export interface CronjobPageResult {
  items: Cronjob[]
  total: number
}

export interface CronjobToggle {
  id: number
  status: 'enabled' | 'disabled'
}

export interface JobRecord {
  id: number
  cronjobId: number
  startTime: string
  duration: number
  status: string
  message: string
}

export interface RecordSearch {
  page: number
  pageSize: number
  cronjobId: number
  status?: string
}

export interface RecordPageResult {
  items: JobRecord[]
  total: number
}

export interface SpecObj {
  specType: string
  week: number
  day: number
  hour: number
  minute: number
}
