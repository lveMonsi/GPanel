import type { LatestRelease, UpdatePhase } from '@/api/interface/update'

const stageablePhases: UpdatePhase[] = ['idle', 'failed', 'completed']

export const canStageUpdate = (
  phase: UpdatePhase,
  currentVersion: string,
  latestRelease: Pick<LatestRelease, 'version'> | null
) => {
  if (!stageablePhases.includes(phase) || !currentVersion || !latestRelease?.version) return false
  return currentVersion !== latestRelease.version
}
