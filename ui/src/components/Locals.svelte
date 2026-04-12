<script lang="ts">
  import type { Snapshot } from '../lib/types'

  let { snapshot }: { snapshot: Snapshot | null } = $props()
</script>

<div class="p-3">
  <h2 class="text-xs font-bold text-neon-lavender uppercase tracking-widest mb-2">Locals</h2>

  {#if snapshot && snapshot.locals.length > 0}
    <div class="space-y-0.5">
      {#each snapshot.locals as local}
        <div class="flex items-baseline gap-2 font-mono text-xs py-1 px-2 rounded hover:bg-surface-hover transition-colors">
          <span class="text-neon-orange font-medium flex-shrink-0">{local.name}</span>
          <span class="text-neon-lavender-dim text-[10px] flex-shrink-0">{local.type}</span>
          <span class="text-text-muted">=</span>
          <span class="text-text-bright truncate">{local.value}</span>
        </div>
      {/each}
    </div>
  {:else}
    <div class="text-xs text-text-muted italic">no locals</div>
  {/if}
</div>
