<template>
  <div class="page-shell">
    <PublicPageHero
      :eyebrow="$t('book.eyebrow')"
      :title="$t('book.title')"
      :subtitle="$t('book.subtitle')"
    />

    <div class="split-layout">
      <PublicCard padding="lg" class="book-form">
        <form @submit.prevent="submit">
          <PublicInput id="name" v-model="form.name" :label="$t('book.name')" required />
          <PublicInput id="email" v-model="form.email" type="email" :label="$t('book.email')" required />
          <div class="book-form__select">
            <label for="slot">{{ $t('book.slot') }}</label>
            <select id="slot" v-model="form.slotId" required>
              <option v-for="slot in slots" :key="slot.id" :value="slot.id">{{ slot.label }}</option>
            </select>
          </div>
          <PublicButton variant="primary" type="submit">{{ $t('book.submit') }}</PublicButton>
        </form>
        <p v-if="success" class="book-form__success" role="status">{{ $t('book.success') }}</p>
      </PublicCard>

      <PublicCard padding="lg" class="book-aside">
        <h3>{{ $t('book.aside_title') }}</h3>
        <ul>
          <li><AppIcon name="done" /> {{ $t('book.aside_1') }}</li>
          <li><AppIcon name="done" /> {{ $t('book.aside_2') }}</li>
          <li><AppIcon name="done" /> {{ $t('book.aside_3') }}</li>
        </ul>
      </PublicCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })

const form = reactive({ name: '', email: '', slotId: '' })
const success = ref(false)

const { data: slotsData } = await useFetch('/api/public/booking/slots')
const slots = computed(() => (slotsData.value as { data?: Array<{ id: string; label: string }> })?.data ?? [])

watch(slots, (list) => {
  if (list.length && !form.slotId) form.slotId = list[0].id
}, { immediate: true })

const submit = async () => {
  await $fetch('/api/public/booking/appointments', {
    method: 'POST',
    body: { ...form, consent: true, website: '' }
  })
  success.value = true
}
</script>

<style scoped>
.book-form form { display: flex; flex-direction: column; gap: var(--kore-space-md); }

.book-form__select { display: flex; flex-direction: column; gap: var(--kore-space-xs); }

.book-form__select label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.book-form__select select {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.book-form__success {
  margin: var(--kore-space-md) 0 0;
  padding: var(--kore-space-sm);
  text-align: center;
  color: var(--kore-success);
  background: rgba(74, 222, 128, 0.08);
  border-radius: var(--kore-radius-md);
  font-size: var(--kore-text-small);
}

.book-aside h3 {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-h3);
}

.book-aside ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}

.book-aside li {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  line-height: 1.5;
}

.book-aside :deep(.material-symbols-outlined) {
  color: var(--kore-brand-gold);
  font-size: 1.125rem !important;
  flex-shrink: 0;
}
</style>
