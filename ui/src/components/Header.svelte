<script lang="ts">
  import type { ConnectionState } from '../lib/ws'
  import type { SessionInfo } from '../lib/types'

  let { connectionState, session }: {
    connectionState: ConnectionState
    session: SessionInfo | null
  } = $props()

  const stateColors: Record<string, string> = {
    idle: 'text-text-muted',
    launched: 'text-neon-purple',
    stopped_at_breakpoint: 'text-neon-orange text-glow-orange',
    running: 'text-neon-lavender',
    terminated: 'text-neon-pink text-glow-pink',
  }

  const connectionDot: Record<ConnectionState, string> = {
    connecting: 'bg-neon-lavender-dim animate-pulse',
    connected: 'bg-neon-purple',
    disconnected: 'bg-neon-pink',
  }
</script>

<header class="h-12 bg-surface-deep border-b border-surface-border flex items-center px-4 gap-4 flex-shrink-0">
  <div class="flex items-center gap-2">
    <span class="text-neon-orange font-bold text-lg tracking-tight text-glow-orange">🍊 OrangeCandy</span>
    <span class="text-text-muted text-xs">debug</span>
  </div>

  <div class="flex-1"></div>

  {#if session}
    <div class="flex items-center gap-3 text-xs font-mono">
      <span class={stateColors[session.state] || 'text-text-muted'}>
        {session.state}
      </span>
      <span class="text-text-muted">
        stops: <span class="text-neon-lavender">{session.stop_count}</span>
      </span>
      <span class="text-text-muted">
        events: <span class="text-neon-purple">{session.event_count}</span>
      </span>
    </div>
  {/if}

  <div class="flex items-center gap-2">
    <div class="w-2 h-2 rounded-full {connectionDot[connectionState]}"></div>
    <span class="text-xs text-text-muted">{connectionState}</span>
  </div>
</header>
