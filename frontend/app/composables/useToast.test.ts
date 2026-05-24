import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from './useToast'

describe('useToast', () => {
  beforeEach(() => {
    const { toasts } = useToast()
    toasts.value = []
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('starts with an empty toast list', () => {
    const { toasts } = useToast()
    expect(toasts.value).toHaveLength(0)
  })

  it('success() adds a success toast with title and message', () => {
    const t = useToast()
    t.success('Saved', 'Changes persisted', 0)
    expect(t.toasts.value).toHaveLength(1)
    expect(t.toasts.value[0]).toMatchObject({
      kind: 'success',
      title: 'Saved',
      message: 'Changes persisted',
    })
  })

  it('error() defaults to a 6s duration when not specified', () => {
    const t = useToast()
    t.error('Boom')
    expect(t.toasts.value[0]!.duration).toBe(6000)
  })

  it('info() and warning() preserve the kind on the toast', () => {
    const t = useToast()
    t.info('Heads up', undefined, 0)
    t.warning('Careful', undefined, 0)
    expect(t.toasts.value.map((x) => x.kind)).toEqual(['info', 'warning'])
  })

  it('auto-dismisses after the configured duration', () => {
    const t = useToast()
    t.success('Quick', 'gone soon', 2000)
    expect(t.toasts.value).toHaveLength(1)
    vi.advanceTimersByTime(1999)
    expect(t.toasts.value).toHaveLength(1)
    vi.advanceTimersByTime(1)
    expect(t.toasts.value).toHaveLength(0)
  })

  it('does not auto-dismiss when duration is 0', () => {
    const t = useToast()
    t.error('Persist', undefined, 0)
    vi.advanceTimersByTime(60_000)
    expect(t.toasts.value).toHaveLength(1)
  })

  it('dismiss() removes a toast by id', () => {
    const t = useToast()
    const id = t.success('A', undefined, 0)
    t.error('B', undefined, 0)
    expect(t.toasts.value).toHaveLength(2)
    t.dismiss(id)
    expect(t.toasts.value.map((x) => x.title)).toEqual(['B'])
  })

  it('assigns monotonically increasing ids', () => {
    const t = useToast()
    const ids = [t.info('a', undefined, 0), t.info('b', undefined, 0), t.info('c', undefined, 0)]
    expect(ids[1]!).toBeGreaterThan(ids[0]!)
    expect(ids[2]!).toBeGreaterThan(ids[1]!)
  })
})
