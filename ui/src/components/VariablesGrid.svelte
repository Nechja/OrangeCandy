<script lang="ts">
  import type { Local } from '../lib/types'

  interface Watch {
    id: number
    expression: string
    current?: { value: string; type: string; error?: string }
  }

  let { locals, watches }: { locals: Local[]; watches: Watch[] } = $props()

  const cleanLocals = $derived(
    locals.filter(l => l.name !== 'this' && !l.name.startsWith('$'))
  )

  const hasWatches = $derived(watches.length > 0)
</script>

<div class="grid {hasWatches ? 'grid-cols-2' : 'grid-cols-1'} gap-4">
  <div class="rounded-xl border border-surface-border bg-surface/50 overflow-hidden">
    <div class="px-4 py-2 border-b border-surface-border/50">
      <span class="text-[10px] font-bold text-neon-lavender uppercase tracking-[0.15em]">Locals</span>
    </div>
    <div class="p-3 space-y-0.5">
      {#each cleanLocals as local}
        <div class="flex items-baseline gap-3 font-mono text-xs py-1 px-2 rounded hover:bg-surface-hover/50 transition-colors">
          <span class="text-neon-orange-soft font-medium flex-shrink-0 min-w-[100px]">{local.name}</span>
          <span class="text-text-bright flex-1 truncate">{local.value}</span>
          <span class="text-text-dim text-[10px] flex-shrink-0">{local.type}</span>
        </div>
      {/each}
      {#if cleanLocals.length === 0}
        <div class="text-[11px] text-text-dim italic px-2">no locals in scope</div>
      {/if}
    </div>
  </div>

  {#if hasWatches}
    <div class="rounded-xl border border-surface-border bg-surface/50 overflow-hidden">
      <div class="px-4 py-2 border-b border-surface-border/50">
        <span class="text-[10px] font-bold text-neon-orange uppercase tracking-[0.15em]">Watches</span>
      </div>
      <div class="p-3 space-y-0.5">
        {#each watches as watch}
          <div class="flex items-baseline gap-3 font-mono text-xs py-1 px-2 rounded hover:bg-surface-hover/50 transition-colors">
            <span class="text-neon-lavender font-medium flex-shrink-0 min-w-[100px]">{watch.expression}</span>
            {#if watch.current?.error}
              <span class="text-neon-pink-dim text-[10px] truncate italic">{watch.current.error}</span>
            {:else if watch.current}
              <span class="text-text-bright flex-1 truncate">{watch.current.value}</span>
              <span class="text-text-dim text-[10px] flex-shrink-0">{watch.current.type}</span>
            {:else}
              <span class="text-text-dim italic">—</span>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
