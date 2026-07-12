<template>
  <div>
    <h1>Réserver un entretien</h1>
    <form @submit.prevent="submit">
      <input v-model="form.name" placeholder="Nom" required />
      <input v-model="form.email" type="email" placeholder="Email" required />
      <select v-model="form.slotId">
        <option v-for="slot in slots" :key="slot.id" :value="slot.id">
          {{ slot.label }}
        </option>
      </select>
      <button type="submit">Réserver</button>
    </form>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'public' })
const form = reactive({ name: '', email: '', slotId: '' })
const { data: slotsData } = await useFetch('/api/public/booking/slots')
const slots = computed(() => slotsData.value?.data ?? [])

const submit = async () => {
  await $fetch('/api/public/booking/appointments', {
    method: 'POST',
    body: { ...form, consent: true, website: '' }
  })
  alert('Réservation enregistrée')
}
</script>
