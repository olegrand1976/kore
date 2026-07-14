<script setup lang="ts">
export type KanbanColumn = {
  id: string
  label: string
  tone?: 'gold' | 'blue' | 'success' | 'warn' | 'muted'
}

const props = defineProps<{
  columns: KanbanColumn[]
  items: unknown[]
  columnKey: (item: unknown) => string
  emptyLabel?: string
}>()

const grouped = computed(() => {
  const map = new Map<string, unknown[]>()
  for (const col of props.columns) {
    map.set(col.id, [])
  }
  for (const item of props.items) {
    const key = props.columnKey(item)
    const bucket = map.get(key)
    if (bucket) {
      bucket.push(item)
    } else {
      const fallback = map.get('unknown') ?? []
      if (!map.has('unknown')) map.set('unknown', fallback)
      fallback.push(item)
    }
  }
  return map
})

function countForColumn(id: string) {
  return grouped.value.get(id)?.length ?? 0
}

function itemsForColumn(id: string) {
  return grouped.value.get(id) ?? []
}
</script>

<template>
  <div class="kanban-board">
    <div
      v-for="col in columns"
      :key="col.id"
      class="kanban-board__column"
      :class="col.tone ? `kanban-board__column--${col.tone}` : undefined"
    >
      <header class="kanban-board__header">
        <h3 class="kanban-board__title">{{ col.label }}</h3>
        <span class="kanban-board__count">{{ countForColumn(col.id) }}</span>
      </header>
      <div class="kanban-board__cards">
        <article
          v-for="(item, idx) in itemsForColumn(col.id)"
          :key="idx"
          class="kanban-board__card"
        >
          <slot name="card" :item="item" :column="col" />
        </article>
        <p v-if="!itemsForColumn(col.id).length" class="kanban-board__empty">
          {{ emptyLabel }}
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.kanban-board {
  display: flex;
  gap: var(--kore-space-md);
  overflow-x: auto;
  padding-bottom: var(--kore-space-sm);
  -webkit-overflow-scrolling: touch;
}

.kanban-board__column {
  flex: 0 0 min(18rem, 85vw);
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-lg);
  background: var(--kore-bg-subtle);
  max-height: 70vh;
}

.kanban-board__column--gold { border-top: 3px solid var(--kore-gold); }
.kanban-board__column--blue { border-top: 3px solid var(--kore-blue, #3b82f6); }
.kanban-board__column--success { border-top: 3px solid var(--kore-success, #16a34a); }
.kanban-board__column--warn { border-top: 3px solid var(--kore-warning, #d97706); }
.kanban-board__column--muted { border-top: 3px solid var(--kore-border); }

.kanban-board__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--kore-space-sm);
}

.kanban-board__title {
  margin: 0;
  font-size: var(--kore-text-small);
  font-weight: 600;
  color: var(--kore-text);
}

.kanban-board__count {
  font-size: var(--kore-text-caption);
  font-weight: 600;
  color: var(--kore-text-muted);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-pill, 999px);
  padding: 0.1rem 0.5rem;
}

.kanban-board__cards {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
  overflow-y: auto;
  flex: 1;
  min-height: 4rem;
}

.kanban-board__card {
  padding: var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
}

.kanban-board__empty {
  margin: 0;
  padding: var(--kore-space-sm);
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
  text-align: center;
}

@media (max-width: 768px) {
  .kanban-board__column {
    flex-basis: min(16rem, 88vw);
    max-height: none;
  }
}
</style>
