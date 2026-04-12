<script lang="ts">
  interface Watch {
    id: number
    expression: string
    current?: {
      value: string
      type: string
      error?: string
      stop_index: number
    }
  }

  let { watches }: { watches: Watch[] } = $props()
</script>

<div class="p-3">
  <h2 class="text-xs font-bold text-neon-orange uppercase tracking-widest mb-2">Watches</h2>

  {#if watches.length > 0}
    <div class="space-y-1">
      {#each watches as watch}
        <div class="font-mono text-xs py-1.5 px-2 rounded bg-surface-deep border border-surface-border hover:border-neon-orange/30 transition-colors">
          <div class="flex items-baseline gap-2">
            <span class="text-neon-lavender">{watch.expression}</span>
            {#if watch.current?.error}
              <span class="text-neon-pink text-[10px] truncate">{watch.current.error}</span>
            {:else if watch.current}
              <span class="text-text-muted">=</span>
              <span class="text-text-bright truncate">{watch.current.value}</span>
            {:else}
              <span class="text-text-muted italic">not evaluated</span>
            {/if}
          </div>
          {#if watch.current?.type && !watch.current?.error}
            <div class="text-[10px] text-neon-purple-dim mt-0.5">{watch.current.type}</div>
          {/if}
        </div>
      {/each}
    </div>
  {:else}
    <div class="text-xs text-text-muted italic">no watches — use add_watch to track expressions</div>
  {/if}
</div>
