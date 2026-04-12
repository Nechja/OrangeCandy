<script lang="ts">
  import { onMount } from 'svelte'
  import type { Snapshot } from '../lib/types'
  import { highlightLines } from '../lib/highlighter'

  let { snapshot }: { snapshot: Snapshot | null } = $props()

  let highlightedLines = $state<string[]>([])

  $effect(() => {
    if (snapshot?.source_window) {
      highlightLines(snapshot.source_window).then(lines => {
        highlightedLines = lines
      })
    } else {
      highlightedLines = []
    }
  })

  function fileName(path: string): string {
    return path.split('/').pop() || path
  }
</script>

<div class="p-4 h-full flex flex-col">
  {#if snapshot}
    <div class="flex items-center gap-2 mb-3">
      <span class="text-xs font-bold text-neon-orange uppercase tracking-widest">Source</span>
      <span class="text-xs font-mono text-neon-lavender">{fileName(snapshot.stopped_at.file)}</span>
      <span class="text-xs text-text-muted">:</span>
      <span class="text-xs font-mono text-neon-orange text-glow-orange">{snapshot.stopped_at.line}</span>
      {#if snapshot.reason === 'exception'}
        <span class="text-xs font-mono text-neon-pink text-glow-pink ml-2">exception</span>
      {/if}
    </div>

    <div class="flex-1 overflow-auto bg-surface-deep rounded-lg border border-surface-border">
      <div class="text-xs leading-relaxed font-mono">
        {#each snapshot.source_window as line, i}
          {@const lineNum = snapshot.stopped_at.line - Math.floor(snapshot.source_window.length / 2) + i}
          {@const isCurrentLine = lineNum === snapshot.stopped_at.line}
          {@const isException = isCurrentLine && snapshot.reason === 'exception'}
          <div class="flex hover:bg-surface-hover transition-colors
            {isException ? 'bg-neon-pink/10 border-l-2 border-neon-pink' : isCurrentLine ? 'bg-neon-orange/10 border-l-2 border-neon-orange' : 'border-l-2 border-transparent'}">
            <span class="w-12 text-right pr-3 select-none flex-shrink-0 py-0.5
              {isException ? 'text-neon-pink' : isCurrentLine ? 'text-neon-orange' : 'text-text-muted'}">{lineNum}</span>
            <span class="flex-1 py-0.5 {isCurrentLine ? 'text-text-bright' : ''}">
              {#if highlightedLines[i]}
                {@html highlightedLines[i]}
              {:else}
                {line}
              {/if}
            </span>
          </div>
        {/each}
      </div>
    </div>

    <div class="mt-2 text-xs font-mono text-neon-purple truncate">
      {snapshot.stopped_at.function}
    </div>
  {:else}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="text-4xl mb-3 opacity-20">🍊</div>
        <div class="text-sm text-text-muted">No debug session active</div>
        <div class="text-xs text-text-muted mt-1">Launch a project to start debugging</div>
      </div>
    </div>
  {/if}
</div>
