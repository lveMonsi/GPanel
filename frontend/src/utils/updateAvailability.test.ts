import { describe, expect, it } from 'vitest'
import type { UpdatePhase } from '@/api/interface/update'
import { canStageUpdate } from './updateAvailability'

const latest = { version: 'v1.1.3' }

const phases: UpdatePhase[] = [
  'checking',
  'downloading',
  'validating',
  'restarting',
  'rolling_back'
]

describe('canStageUpdate', () => {
  it.each([
    ['idle', true],
    ['failed', true],
    ['completed', true]
  ] as const)('%s allows staging when a newer release is available', (phase, expected) => {
    expect(canStageUpdate(phase, 'v1.1.2', latest)).toBe(expected)
  })

  it('does not stage when the current version is latest', () => {
    expect(canStageUpdate('completed', 'v1.1.3', latest)).toBe(false)
  })

  it('does not stage a pending update again', () => {
    expect(canStageUpdate('staged', 'v1.1.2', latest)).toBe(false)
  })

  it.each(phases)('does not stage while phase is %s', (phase) => {
    expect(canStageUpdate(phase, 'v1.1.2', latest)).toBe(false)
  })

  it.each([
    ['completed', '', latest],
    ['completed', 'v1.1.2', null],
    ['completed', 'v1.1.2', { version: '' }]
  ] as const)('does not stage when version data is incomplete (%s)', (phase, current, release) => {
    expect(canStageUpdate(phase, current, release)).toBe(false)
  })
})
