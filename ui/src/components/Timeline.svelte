<script lang="ts">
  import type { TimelineEntry } from '../lib/types'

  let { entries }: { entries: TimelineEntry[] } = $props()

  function icon(entry: TimelineEntry): string {
    switch (entry.type) {
      case 'launch': return '🚀'
      case 'restart': return '🔄'
      case 'stop': return entry.detail?.reason === 'exception' ? '💥' : '⏸'
      case 'action': return '🔧'
      case 'terminated': return '⏹'
      default: return '•'
    }
  }

  function color(entry: TimelineEntry): string {
    switch (entry.type) {
      case 'launch':
      case 'restart':
        return 'border-neon-purple'
      case 'stop':
        return entry.detail?.reason === 'exception'
          ? 'border-neon-pink'
          : 'border-neon-orange'
      case 'action':
        return 'border-neon-lavender-dim'
      case 'terminated':
        return 'border-neon-pink'
      default:
        return 'border-surface-border'
    }
  }

  function label(entry: TimelineEntry): string {
    if (entry.type === 'stop') {
      const fn = entry.detail?.function || ''
      const short = fn.split('.').pop()?.replace('()', '') || fn
      return short
    }
    if (entry.type === 'action') {
      return entry.tool || 'action'
    }
    return entry.type
  }

  function detail(entry: TimelineEntry): string {
    if (entry.type === 'stop' && entry.detail) {
      const file = (entry.detail.file || '').split('/').pop() || ''
      return `${file}:${entry.detail.line}`
    }
    if (entry.type === 'action' && entry.detail) {
      if (entry.tool === 'set_breakpoint') {
        const file = (entry.detail.file || '').split('/').pop() || ''
        return `${file}:${entry.detail.line}`
      }
      if (entry.tool === 'continue') return ''
    }
    if (entry.type === 'launch' && entry.detail?.args) {
      return entry.detail.args.join(' ')
    }
    return ''
  }

  function time(ts: string): string {
    const d = new Date(ts)
    return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }
</script>

<div class="p-3">
  <h2 class="text-xs font-bold text-neon-purple uppercase tracking-widest mb-3">Timeline</h2>

  <div class="space-y-1">
    {#each entries as entry (entry.index)}
      <div class="animate-slide-in flex items-start gap-2 pl-2 border-l-2 {color(entry)} py-1.5 hover:bg-surface-hover rounded-r transition-colors cursor-default">
        <span class="text-xs flex-shrink-0 w-4">{icon(entry)}</span>
        <div class="min-w-0 flex-1">
          <div class="text-xs font-mono text-text-bright truncate">{label(entry)}</div>
          {#if detail(entry)}
            <div class="text-[10px] font-mono text-text-muted truncate">{detail(entry)}</div>
          {/if}
        </div>
        <span class="text-[10px] text-text-muted flex-shrink-0">{time(entry.timestamp)}</span>
      </div>
    {/each}

    {#if entries.length === 0}
      <div class="text-xs text-text-muted italic py-4 text-center">waiting for debug session...</div>
    {/if}
  </div>
</div>
