<script lang="ts">
  let { lines }: { lines: string[] } = $props()
</script>

<div class="mt-2 rounded-xl border border-surface-border bg-surface-deep/80 overflow-hidden">
  <div class="p-3 max-h-40 overflow-y-auto font-mono text-[11px] leading-5 text-text-muted space-y-0">
    {#each lines as line}
      <div class="hover:text-text transition-colors whitespace-pre-wrap break-all">{line}</div>
    {/each}
  </div>
</div>
