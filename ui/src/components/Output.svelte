<script lang="ts">
  let { lines }: { lines: string[] } = $props()
</script>

<div class="p-3">
  <h2 class="text-xs font-bold text-neon-pink uppercase tracking-widest mb-2">Output</h2>

  <div class="font-mono text-[11px] leading-relaxed space-y-0">
    {#each lines as line}
      <div class="text-text-muted hover:text-text transition-colors whitespace-pre-wrap break-all">{line}</div>
    {/each}

    {#if lines.length === 0}
      <div class="text-xs text-text-muted italic">no output</div>
    {/if}
  </div>
</div>
