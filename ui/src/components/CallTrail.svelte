<script lang="ts">
  import type { Frame } from '../lib/types'

  let { frames }: { frames: Frame[] } = $props()

  function shortFn(fn: string): string {
    const parts = fn.split('.')
    if (parts.length <= 2) return fn.replace('()', '')
    return parts.slice(-2).join('.').replace('()', '')
  }

  function shortFile(path: string): string {
    return path?.split('/').pop()?.replace('.cs', '') || ''
  }

  const trail = $derived(
    frames
      .filter(f => f.file)
      .slice(0, 5)
      .reverse()
  )
</script>

<div class="flex items-center gap-1 flex-wrap font-mono text-xs">
  {#each trail as frame, i}
    {#if i > 0}
      <span class="text-text-dim mx-0.5">→</span>
    {/if}
    <span class="{i === trail.length - 1 ? 'text-neon-orange' : 'text-text-muted'} hover:text-neon-lavender transition-colors">
      {shortFn(frame.function)}
    </span>
  {/each}
</div>
