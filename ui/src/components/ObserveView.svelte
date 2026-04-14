<script lang="ts">
  import type { ObserveEvent } from '../lib/types'

  let { events }: { events: ObserveEvent[] } = $props()

  interface CallRow {
    traceId: string
    iface: string
    method: string
    args?: string[]
    returnValue?: string
    exception?: string
    durationMs?: number
    depth: number
    timestamp: string
    status: 'running' | 'completed' | 'failed'
  }

  const rows = $derived.by(() => {
    const map = new Map<string, CallRow>()
    const result: CallRow[] = []

    for (const evt of events) {
      if (evt.event_type === 'method_enter') {
        const row: CallRow = {
          traceId: evt.trace_id,
          iface: evt.interface,
          method: evt.method,
          args: evt.arguments,
          depth: evt.depth,
          timestamp: evt.timestamp,
          status: 'running',
        }
        map.set(evt.trace_id, row)
        result.push(row)
      } else if (evt.event_type === 'method_exit') {
        const existing = map.get(evt.trace_id)
        if (existing) {
          existing.returnValue = evt.return_value
          existing.durationMs = evt.duration_ms
          existing.status = 'completed'
        }
      } else if (evt.event_type === 'method_exception') {
        const existing = map.get(evt.trace_id)
        if (existing) {
          existing.exception = evt.exception
          existing.durationMs = evt.duration_ms
          existing.status = 'failed'
        }
      }
    }

    return result
  })

  function durationBadge(ms?: number): { text: string, cls: string } {
    if (ms === undefined) return { text: '...', cls: 'text-text-dim' }
    if (ms < 1) return { text: '<1ms', cls: 'text-neon-mint' }
    if (ms < 10) return { text: `${ms}ms`, cls: 'text-neon-mint' }
    if (ms < 50) return { text: `${ms}ms`, cls: 'text-neon-lavender' }
    if (ms < 200) return { text: `${ms}ms`, cls: 'text-neon-orange' }
    return { text: `${ms}ms`, cls: 'text-neon-pink' }
  }

  function shortReturn(val?: string): string {
    if (!val) return ''
    try {
      const parsed = JSON.parse(val)
      if (parsed?.Value !== undefined) {
        const v = typeof parsed.Value === 'object' ? JSON.stringify(parsed.Value) : String(parsed.Value)
        return v.length > 60 ? v.slice(0, 57) + '...' : v
      }
    } catch {}
    return val.length > 60 ? val.slice(0, 57) + '...' : val
  }

  function time(ts: string): string {
    return new Date(ts).toLocaleTimeString('en-US', {
      hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit',
      fractionalSecondDigits: 3
    })
  }

  let expandedTrace = $state<string | null>(null)
</script>

<div class="p-6 max-w-5xl mx-auto anim-fade-up">
  <div class="flex items-center gap-4 mb-6">
    <div class="flex items-center gap-2">
      <div class="w-2.5 h-2.5 rounded-full bg-neon-mint anim-pulse"></div>
      <span class="text-xs font-bold text-neon-mint uppercase tracking-[0.2em]">Observe</span>
    </div>
    <span class="text-xs font-mono text-text-muted">{rows.length} calls captured</span>
    <div class="flex-1"></div>
    <div class="flex gap-4 text-[10px] font-mono text-text-dim">
      <span class="flex items-center gap-1.5"><span class="w-1.5 h-1.5 rounded-full bg-neon-mint"></span> fast</span>
      <span class="flex items-center gap-1.5"><span class="w-1.5 h-1.5 rounded-full bg-neon-lavender"></span> normal</span>
      <span class="flex items-center gap-1.5"><span class="w-1.5 h-1.5 rounded-full bg-neon-orange"></span> slow</span>
      <span class="flex items-center gap-1.5"><span class="w-1.5 h-1.5 rounded-full bg-neon-pink"></span> critical</span>
    </div>
  </div>

  {#if rows.length > 0}
    <div class="relative">
      {#each rows as row, i}
        {@const badge = durationBadge(row.durationMs)}
        {@const isExpanded = expandedTrace === row.traceId}
        {@const depthPx = row.depth * 24}

        <div
          class="group relative anim-fade-up"
          style="animation-delay: {Math.min(i * 20, 400)}ms"
        >
          <!-- Depth connector lines -->
          {#if row.depth > 0}
            <div class="absolute top-0 bottom-0" style="left: {depthPx - 12}px">
              <div class="w-px h-full
                {row.status === 'failed' ? 'bg-neon-pink/20' :
                 row.status === 'completed' ? 'bg-neon-purple/15' :
                 'bg-surface-border'}"></div>
            </div>
            <div class="absolute top-4" style="left: {depthPx - 12}px; width: 12px">
              <div class="h-px w-full
                {row.status === 'failed' ? 'bg-neon-pink/30' :
                 row.status === 'completed' ? 'bg-neon-purple/20' :
                 'bg-surface-border'}"></div>
            </div>
          {/if}

          <!-- Main row -->
          <div
            class="relative flex items-center gap-2 py-2 px-3 rounded-lg mb-0.5 cursor-default transition-all duration-150
              {row.status === 'failed' ? 'bg-neon-pink/5 hover:bg-neon-pink/10' :
               isExpanded ? 'bg-surface-raised' :
               'hover:bg-surface-hover/50'}"
            style="margin-left: {depthPx}px"
            onclick={() => expandedTrace = isExpanded ? null : row.traceId}
          >
            <!-- Status dot -->
            <div class="flex-shrink-0 w-2 h-2 rounded-full transition-all
              {row.status === 'running' ? 'bg-neon-lavender-dim anim-pulse' :
               row.status === 'failed' ? 'bg-neon-pink shadow-[0_0_8px_rgba(255,45,138,0.4)]' :
               'bg-neon-mint shadow-[0_0_6px_rgba(52,211,153,0.3)]'}"></div>

            <!-- Interface.Method -->
            <span class="font-mono text-xs flex-shrink-0">
              <span class="text-neon-purple-soft">{row.iface.replace(/^I/, '')}</span>
              <span class="text-text-dim">.</span>
              <span class="{row.status === 'failed' ? 'text-neon-pink' : 'text-neon-orange-soft'} font-medium">{row.method}</span>
            </span>

            <!-- Duration badge -->
            <span class="flex-shrink-0 text-[10px] font-mono px-1.5 py-0.5 rounded-md bg-surface-deep/80 {badge.cls}">
              {badge.text}
            </span>

            <!-- Return value preview -->
            {#if row.exception}
              <span class="text-[10px] font-mono text-neon-pink truncate flex-1">{row.exception}</span>
            {:else if row.returnValue}
              <span class="text-[10px] font-mono text-text-muted truncate flex-1 opacity-60 group-hover:opacity-100 transition-opacity">
                → {shortReturn(row.returnValue)}
              </span>
            {:else}
              <span class="flex-1"></span>
            {/if}

            <!-- Timestamp -->
            <span class="text-[9px] font-mono text-text-dim flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
              {time(row.timestamp)}
            </span>
          </div>

          <!-- Expanded detail -->
          {#if isExpanded}
            <div class="ml-8 mb-2 rounded-lg bg-surface-deep border border-surface-border overflow-hidden anim-fade-up"
                 style="margin-left: {depthPx + 32}px">
              {#if row.args && row.args.length > 0}
                <div class="px-3 py-2 border-b border-surface-border/50">
                  <div class="text-[9px] text-neon-lavender-dim uppercase tracking-widest mb-1">arguments</div>
                  {#each row.args as arg, j}
                    <div class="font-mono text-[11px] text-text py-0.5 break-all">{arg}</div>
                  {/each}
                </div>
              {/if}
              {#if row.returnValue}
                <div class="px-3 py-2 border-b border-surface-border/50">
                  <div class="text-[9px] text-neon-mint uppercase tracking-widest mb-1">return</div>
                  <div class="font-mono text-[11px] text-text break-all">{row.returnValue}</div>
                </div>
              {/if}
              {#if row.exception}
                <div class="px-3 py-2">
                  <div class="text-[9px] text-neon-pink uppercase tracking-widest mb-1">exception</div>
                  <div class="font-mono text-[11px] text-neon-pink-soft break-all">{row.exception}</div>
                </div>
              {/if}
              <div class="px-3 py-1.5 bg-surface/30 flex gap-4 text-[9px] font-mono text-text-dim">
                <span>trace: <span class="text-neon-purple-dim">{row.traceId}</span></span>
                <span>depth: <span class="text-neon-lavender-dim">{row.depth}</span></span>
                {#if row.durationMs !== undefined}
                  <span>duration: <span class="{badge.cls}">{row.durationMs}ms</span></span>
                {/if}
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {:else}
    <div class="flex flex-col items-center justify-center py-24 gap-3">
      <div class="relative">
        <div class="w-16 h-16 rounded-full border border-neon-mint/20 flex items-center justify-center">
          <div class="w-8 h-8 rounded-full border border-neon-mint/40 flex items-center justify-center">
            <div class="w-3 h-3 rounded-full bg-neon-mint/20 anim-pulse"></div>
          </div>
        </div>
      </div>
      <div class="text-sm text-text-muted">Listening for method calls</div>
      <div class="text-xs text-text-dim">Run the app with OrangeCandy.Observe to see the flow</div>
    </div>
  {/if}
</div>
