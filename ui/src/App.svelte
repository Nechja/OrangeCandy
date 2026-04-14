<script lang="ts">
  import { onMount } from 'svelte'
  import { createWebSocket, type ConnectionState } from './lib/ws'
  import type { SessionInfo, TimelineEntry, Snapshot, ObserveEvent } from './lib/types'
  import StatusBar from './components/StatusBar.svelte'
  import TimelineRail from './components/TimelineRail.svelte'
  import StopHero from './components/StopHero.svelte'
  import SourceExplorer from './components/SourceExplorer.svelte'
  import ObserveView from './components/ObserveView.svelte'

  interface SourceView {
    file: string
    start_line: number
    lines: string[]
    center_line: number
  }

  let connectionState = $state<ConnectionState>('connecting')
  let session = $state<SessionInfo | null>(null)
  let timeline = $state<TimelineEntry[]>([])
  let currentSnapshot = $state<Snapshot | null>(null)
  let watches = $state<any[]>([])
  let output = $state<string[]>([])
  let sourceView = $state<SourceView | null>(null)
  let observeEvents = $state<ObserveEvent[]>([])
  let observing = $state(false)

  let viewMode = $state<'stop' | 'source' | 'observe'>('stop')

  onMount(() => {
    fetch('/api/session').then(r => r.json()).then(data => { session = data })
    fetch('/api/timeline').then(r => r.json()).then(data => { timeline = data || [] })

    const ws = createWebSocket(
      (event) => {
        switch (event.type) {
          case 'session':
            session = event.data
            break
          case 'stop':
            if (event.data?.snapshot) {
              currentSnapshot = event.data.snapshot
              viewMode = 'stop'
            }
            break
          case 'timeline':
            timeline = [...timeline, event.data]
            if (event.data.snapshot) {
              currentSnapshot = event.data.snapshot
              viewMode = 'stop'
            }
            break
          case 'output':
            output = [...output, event.data.text]
            break
          case 'watches':
            watches = event.data
            break
          case 'show_source':
            sourceView = event.data
            viewMode = 'source'
            break
          case 'show_stop':
            if (event.data?.snapshot) {
              currentSnapshot = event.data.snapshot
              viewMode = 'stop'
            }
            break
          case 'observe':
            observeEvents = [...observeEvents, event.data]
            if (viewMode !== 'source') viewMode = 'observe'
            break
          case 'observe_status':
            observing = event.data.observing
            if (!event.data.observing && viewMode === 'observe' && observeEvents.length === 0) {
              viewMode = 'stop'
            }
            break
        }
      },
      (state) => { connectionState = state }
    )

    return () => ws.disconnect()
  })
</script>

<div class="h-screen flex flex-col overflow-hidden relative noise">
  <StatusBar {connectionState} {session} {observing} snapshot={viewMode === 'stop' ? currentSnapshot : null} />

  <div class="flex-1 flex overflow-hidden relative z-10">
    <div class="w-64 flex-shrink-0 overflow-y-auto border-r border-surface-border bg-surface-deep/50">
      <TimelineRail entries={timeline} />
    </div>

    <div class="flex-1 overflow-y-auto">
      {#if viewMode === 'observe'}
        <ObserveView events={observeEvents} />
      {:else if viewMode === 'source' && sourceView}
        <SourceExplorer source={sourceView} onClose={() => { viewMode = 'stop' }} />
      {:else}
        <StopHero snapshot={currentSnapshot} {watches} {output} />
      {/if}
    </div>
  </div>
</div>
