<script setup lang="ts">
definePageMeta({ layout: 'default', middleware: 'admin' })

await navigateTo('/admin/parametres/demandes', { replace: true })
</script>
