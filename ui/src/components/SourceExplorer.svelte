<script lang="ts">
  import { highlightLines } from '../lib/highlighter'

  interface SourceView {
    file: string
    start_line: number
    lines: string[]
    center_line: number
  }

  let { source, onClose }: { source: SourceView; onClose: () => void } = $props()

  let highlighted = $state<string[]>([])

  $effect(() => {
    if (source?.lines) {
      highlightLines(source.lines).then(lines => { highlighted = lines })
    }
  })

  function fileName(path: string): string {
    return path.split('/').pop() || path
  }
</script>

<div class="p-6 max-w-5xl mx-auto space-y-4 anim-fade-up">
  <!-- Header with close button -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <span class="text-[10px] font-bold text-neon-purple uppercase tracking-[0.15em]">Exploring</span>
      <span class="font-mono text-sm text-neon-lavender">{fileName(source.file)}</span>
    </div>
    <button
      onclick={onClose}
      class="text-xs font-mono text-text-dim hover:text-neon-orange transition-colors cursor-pointer px-2 py-1 rounded border border-surface-border hover:border-neon-orange/30">
      ← back to debug
    </button>
  </div>

  <!-- Source block -->
  <div class="rounded-xl overflow-hidden border border-surface-border bg-surface-deep/80 glow-purple">
    <div class="px-4 py-2 border-b border-surface-border/50 flex items-center gap-2">
      <div class="w-2 h-2 rounded-full bg-neon-purple"></div>
      <span class="text-[11px] font-mono text-text-muted">{fileName(source.file)}</span>
      <span class="text-[11px] font-mono text-text-dim">· line {source.center_line}</span>
    </div>

    <div class="font-mono text-[13px] leading-6">
      {#each source.lines as line, i}
        {@const lineNum = source.start_line + i}
        {@const isCenter = lineNum === source.center_line}

        <div class="flex group transition-colors duration-100
          {isCenter ? 'bg-neon-purple/8 border-l-2 border-neon-purple' : 'border-l-2 border-transparent hover:bg-surface-hover/40'}">
          <span class="w-14 text-right pr-4 select-none flex-shrink-0
            {isCenter ? 'text-neon-purple' : 'text-text-dim group-hover:text-text-muted'} transition-colors">
            {lineNum}
          </span>
          <span class="w-1 flex-shrink-0 {isCenter ? 'bg-neon-purple' : 'bg-transparent'}"></span>
          <span class="flex-1 pl-3 pr-4 {isCenter ? 'text-text-bright' : ''}">
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

  <div class="text-[10px] font-mono text-text-dim truncate">
    {source.file}
  </div>
</div>
