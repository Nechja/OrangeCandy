<script lang="ts">
  import type { Snapshot } from '../lib/types'
  import SourceBlock from './SourceBlock.svelte'
  import VariablesGrid from './VariablesGrid.svelte'
  import CallTrail from './CallTrail.svelte'
  import OutputLog from './OutputLog.svelte'

  let { snapshot, watches, output }: {
    snapshot: Snapshot | null
    watches: any[]
    output: string[]
  } = $props()

  let showOutput = $state(false)
</script>

{#if snapshot}
  <div class="p-6 max-w-5xl mx-auto space-y-5 anim-fade-up">

    <!-- Call stack as breadcrumb trail -->
    <CallTrail frames={snapshot.call_stack} />

    <!-- Source block — the code context -->
    <SourceBlock {snapshot} />

    <!-- Variables — locals and watches merged -->
    <VariablesGrid locals={snapshot.locals} {watches} />

    <!-- Output toggle -->
    {#if output.length > 0}
      <div>
        <button
          onclick={() => showOutput = !showOutput}
          class="text-[10px] font-mono uppercase tracking-widest text-text-dim hover:text-neon-pink transition-colors cursor-pointer">
          output ({output.length}) {showOutput ? '▾' : '▸'}
        </button>
        {#if showOutput}
          <OutputLog lines={output} />
        {/if}
      </div>
    {/if}
  </div>

{:else}
  <!-- Empty state -->
  <div class="h-full flex items-center justify-center">
    <div class="text-center space-y-3">
      <div class="text-5xl opacity-10">🍊</div>
      <div class="text-sm text-text-muted">No debug session</div>
      <div class="text-xs text-text-dim">The AI will launch a session — watch it unfold here</div>
    </div>
  </div>
{/if}
