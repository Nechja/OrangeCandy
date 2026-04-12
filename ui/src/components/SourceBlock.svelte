<script lang="ts">
  import type { Snapshot } from '../lib/types'
  import { highlightLines } from '../lib/highlighter'

  let { snapshot }: { snapshot: Snapshot } = $props()

  let highlighted = $state<string[]>([])

  $effect(() => {
    if (snapshot?.source_window) {
      highlightLines(snapshot.source_window).then(lines => {
        highlighted = lines
      })
    } else {
      highlighted = []
    }
  })

  const startLine = $derived(
    snapshot ? snapshot.stopped_at.line - Math.floor(snapshot.source_window.length / 2) : 0
  )
</script>

<div class="rounded-xl overflow-hidden border border-surface-border bg-surface-deep/80
  {snapshot.reason === 'exception' ? 'glow-pink' : 'glow-orange'}">

  <!-- File header -->
  <div class="px-4 py-2 border-b border-surface-border/50 flex items-center gap-2">
    <div class="w-2 h-2 rounded-full {snapshot.reason === 'exception' ? 'bg-neon-pink' : 'bg-neon-orange'}"></div>
    <span class="text-[11px] font-mono text-text-muted">
      {snapshot.stopped_at.file.split('/').pop()}
    </span>
    <span class="text-[11px] font-mono text-text-dim">:</span>
    <span class="text-[11px] font-mono {snapshot.reason === 'exception' ? 'text-neon-pink' : 'text-neon-orange'} font-bold">
      {snapshot.stopped_at.line}
    </span>
  </div>

  <!-- Source lines -->
  <div class="font-mono text-[13px] leading-6">
    {#each snapshot.source_window as line, i}
      {@const lineNum = startLine + i}
      {@const isCurrent = lineNum === snapshot.stopped_at.line}
      {@const isException = isCurrent && snapshot.reason === 'exception'}

      <div class="flex group transition-colors duration-100
        {isException ? 'bg-neon-pink/8 glow-pink-line' :
         isCurrent ? 'bg-neon-orange/8 glow-orange-line' :
         'hover:bg-surface-hover/40'}">

        <!-- Line number -->
        <span class="w-14 text-right pr-4 select-none flex-shrink-0
          {isException ? 'text-neon-pink' :
           isCurrent ? 'text-neon-orange' :
           'text-text-dim group-hover:text-text-muted'} transition-colors">
          {lineNum}
        </span>

        <!-- Gutter mark -->
        <span class="w-1 flex-shrink-0
          {isException ? 'bg-neon-pink' :
           isCurrent ? 'bg-neon-orange' :
           'bg-transparent'}">
        </span>

        <!-- Code -->
        <span class="flex-1 pl-3 pr-4 {isCurrent ? 'text-text-bright' : ''}">
          {#if highlighted[i]}
            {@html highlighted[i]}
          {:else}
            {line || ' '}
          {/if}
        </span>
      </div>
    {/each}
  </div>
</div>
