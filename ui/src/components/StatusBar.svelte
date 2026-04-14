<script lang="ts">
  import type { ConnectionState } from '../lib/ws'
  import type { SessionInfo, Snapshot } from '../lib/types'

  let { connectionState, session, snapshot, observing }: {
    connectionState: ConnectionState
    session: SessionInfo | null
    snapshot: Snapshot | null
    observing: boolean
  } = $props()

  function shortFile(path: string): string {
    return path?.split('/').pop() || ''
  }

  function shortFn(fn: string): string {
    const parts = fn?.split('.') || []
    return parts.length > 1 ? parts.slice(-2).join('.').replace('()', '') : fn || ''
  }

  const connectionColors: Record<ConnectionState, string> = {
    connecting: 'bg-neon-lavender-dim anim-pulse',
    connected: 'bg-neon-mint',
    disconnected: 'bg-neon-pink',
  }

  const stateLabels: Record<string, { text: string, class: string }> = {
    idle: { text: 'idle', class: 'text-text-muted' },
    launched: { text: 'launched', class: 'text-neon-purple' },
    stopped_at_breakpoint: { text: 'stopped', class: 'text-neon-orange text-glow-orange' },
    running: { text: 'running', class: 'text-neon-mint' },
    terminated: { text: 'terminated', class: 'text-neon-pink' },
  }
</script>

<header class="h-10 flex items-center px-5 gap-6 border-b border-surface-border bg-surface-deep/80 backdrop-blur-sm flex-shrink-0 relative z-20">
  <!-- Brand -->
  <div class="flex items-center gap-2">
    <span class="text-base font-bold tracking-tight text-neon-orange text-glow-orange">OrangeCandy</span>
    <span class="text-[10px] text-text-dim font-mono uppercase tracking-widest">debug</span>
  </div>

  <!-- Location breadcrumb -->
  {#if snapshot}
    <div class="flex items-center gap-1.5 font-mono text-xs">
      <span class="text-neon-purple-soft">{shortFn(snapshot.stopped_at.function)}</span>
      <span class="text-text-dim">·</span>
      <span class="text-neon-lavender-dim">{shortFile(snapshot.stopped_at.file)}</span>
      <span class="text-text-dim">:</span>
      <span class="text-neon-orange font-bold">{snapshot.stopped_at.line}</span>
      {#if snapshot.reason === 'exception'}
        <span class="ml-1 px-1.5 py-0.5 rounded text-[10px] bg-neon-pink/15 text-neon-pink font-bold uppercase tracking-wider">exception</span>
      {/if}
    </div>
  {/if}

  {#if observing}
    <span class="px-1.5 py-0.5 rounded text-[10px] bg-neon-mint/15 text-neon-mint font-bold uppercase tracking-wider anim-pulse">observing</span>
  {/if}

  <div class="flex-1"></div>
  {#if session}
    {@const state = stateLabels[session.state] || stateLabels.idle}
    <span class="font-mono text-xs {state.class}">{state.text}</span>
    <div class="flex gap-3 text-[10px] font-mono text-text-dim">
      <span>stops <span class="text-neon-lavender-dim">{session.stop_count}</span></span>
      <span>events <span class="text-neon-purple-dim">{session.event_count}</span></span>
    </div>
  {/if}

  <!-- Connection -->
  <div class="flex items-center gap-1.5">
    <div class="w-1.5 h-1.5 rounded-full {connectionColors[connectionState]}"></div>
  </div>
</header>
