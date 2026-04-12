<script lang="ts">
  import type { Snapshot } from '../lib/types'

  let { snapshot }: { snapshot: Snapshot | null } = $props()

  function shortFunction(fn: string): string {
    const parts = fn.split('.')
    return parts.length > 2 ? parts.slice(-2).join('.') : fn
  }

  function shortFile(path: string): string {
    return path.split('/').pop() || path
  }
</script>

<div class="p-3">
  <h2 class="text-xs font-bold text-neon-purple uppercase tracking-widest mb-2">Call Stack</h2>

  {#if snapshot && snapshot.call_stack.length > 0}
    <div class="space-y-0.5">
      {#each snapshot.call_stack as frame, i}
        <div class="font-mono text-xs py-1.5 px-2 rounded hover:bg-surface-hover transition-colors
          {i === 0 ? 'border-l-2 border-neon-orange bg-surface-raised' : 'border-l-2 border-surface-border'}">
          <div class="truncate {i === 0 ? 'text-neon-orange' : 'text-neon-lavender'}">
            {shortFunction(frame.function)}
          </div>
          {#if frame.file}
            <div class="text-[10px] text-text-muted truncate">
              {shortFile(frame.file)}:{frame.line}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {:else}
    <div class="text-xs text-text-muted italic">no call stack</div>
  {/if}
</div>
