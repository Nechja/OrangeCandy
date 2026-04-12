<script lang="ts">
  import type { TimelineEntry } from '../lib/types'

  let { entries }: { entries: TimelineEntry[] } = $props()

  function dotColor(entry: TimelineEntry): string {
    switch (entry.type) {
      case 'launch':
      case 'restart':
        return 'bg-neon-purple glow-purple'
      case 'stop':
        return entry.detail?.reason === 'exception'
          ? 'bg-neon-pink glow-pink'
          : 'bg-neon-orange glow-orange'
      case 'action':
        return 'bg-surface-border-bright'
      case 'terminated':
        return 'bg-neon-pink-dim'
      default:
        return 'bg-surface-border'
    }
  }

  function lineColor(entry: TimelineEntry): string {
    switch (entry.type) {
      case 'stop':
        return entry.detail?.reason === 'exception'
          ? 'border-neon-pink/40'
          : 'border-neon-orange/40'
      case 'launch':
      case 'restart':
        return 'border-neon-purple/40'
      default:
        return 'border-surface-border'
    }
  }

  function label(entry: TimelineEntry): string {
    if (entry.type === 'stop') {
      const fn = entry.detail?.function || ''
      const parts = fn.split('.')
      const method = parts.pop()?.replace('()', '') || fn
      return method
    }
    if (entry.type === 'action') return entry.tool || 'action'
    return entry.type
  }

  function sublabel(entry: TimelineEntry): string {
    if (entry.type === 'stop' && entry.detail) {
      const file = (entry.detail.file || '').split('/').pop() || ''
      return `${file}:${entry.detail.line}`
    }
    if (entry.type === 'action' && entry.detail) {
      if (entry.tool === 'set_breakpoint') {
        const file = (entry.detail.file || '').split('/').pop() || ''
        return `${file}:${entry.detail.line}`
      }
    }
    if (entry.type === 'launch' && entry.detail?.args) {
      return (entry.detail.args as string[]).join(' ')
    }
    return ''
  }

  function isSignificant(entry: TimelineEntry): boolean {
    return entry.type === 'stop' || entry.type === 'launch' || entry.type === 'restart' || entry.type === 'terminated'
  }

  function time(ts: string): string {
    return new Date(ts).toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }
</script>

<div class="py-4 px-3">
  <div class="text-[10px] font-bold text-text-dim uppercase tracking-[0.2em] mb-4 px-2">Timeline</div>

  <div class="relative">
    {#each entries as entry, i (entry.index)}
      {@const significant = isSignificant(entry)}
      <div class="anim-slide-left flex items-start gap-3 relative"
           style="animation-delay: {Math.min(i * 30, 300)}ms">

        <div class="flex flex-col items-center flex-shrink-0 w-4">
          {#if i > 0}
            <div class="w-0 border-l {lineColor(entries[i-1])} h-2"></div>
          {:else}
            <div class="h-1"></div>
          {/if}
          <div class="rounded-full flex-shrink-0 {dotColor(entry)}
            {significant ? 'w-2.5 h-2.5' : 'w-1.5 h-1.5'}"></div>
          {#if i < entries.length - 1}
            <div class="w-0 border-l {lineColor(entry)} flex-1 min-h-2"></div>
          {/if}
        </div>

        <div class="flex-1 min-w-0 pb-3 {significant ? '' : 'pb-1.5'}">
          {#if significant}
            <div class="rounded-lg bg-surface-raised/60 border border-surface-border px-3 py-2 hover:bg-surface-hover transition-colors cursor-default">
              <div class="flex items-baseline justify-between gap-2">
                <span class="text-xs font-mono font-medium truncate
                  {entry.type === 'stop' && entry.detail?.reason === 'exception' ? 'text-neon-pink' :
                   entry.type === 'stop' ? 'text-neon-orange' :
                   entry.type === 'terminated' ? 'text-neon-pink-dim' :
                   'text-neon-purple-soft'}">
                  {label(entry)}
                </span>
                <span class="text-[9px] text-text-dim flex-shrink-0 font-mono">{time(entry.timestamp)}</span>
              </div>
              {#if sublabel(entry)}
                <div class="text-[10px] font-mono text-text-muted mt-0.5 truncate">{sublabel(entry)}</div>
              {/if}
            </div>
          {:else}
            <div class="flex items-baseline justify-between gap-2 px-1 py-0.5">
              <span class="text-[10px] font-mono text-text-dim truncate">{label(entry)}</span>
              {#if sublabel(entry)}
                <span class="text-[9px] font-mono text-text-dim truncate">{sublabel(entry)}</span>
              {/if}
            </div>
          {/if}
        </div>
      </div>
    {/each}

    {#if entries.length === 0}
      <div class="text-xs text-text-dim text-center py-12 italic">
        waiting for session...
      </div>
    {/if}
  </div>
</div>
