export interface Snapshot {
  stopped_at: Location
  source_window: string[]
  locals: Local[]
  call_stack: Frame[]
  reason: string
}

export interface Location {
  file: string
  line: number
  function: string
}

export interface Local {
  name: string
  type: string
  value: string
}

export interface Frame {
  function: string
  file: string
  line: number
}

export interface StopRecord {
  index: number
  timestamp: string
  snapshot: Snapshot
  thread_id: number
}

export interface TimelineEntry {
  index: number
  timestamp: string
  type: string // 'launch' | 'restart' | 'stop' | 'action' | 'terminated'
  tool?: string
  detail?: Record<string, any>
  snapshot?: Snapshot
}

export interface SessionInfo {
  state: string
  project_path?: string
  args?: string[]
  stop_count: number
  event_count: number
  current?: StopRecord
  breakpoints: BreakpointInfo[]
}

export interface BreakpointInfo {
  file: string
  line: number
  hit_count: number
}

export interface OutputLine {
  timestamp: string
  category: string
  text: string
}

export interface ObserveEvent {
  trace_id: string
  event_type: 'method_enter' | 'method_exit' | 'method_exception'
  interface: string
  method: string
  arguments?: string[]
  return_value?: string
  exception?: string
  duration_ms?: number
  depth: number
  timestamp: string
}
