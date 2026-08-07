<template>
  <div class="client-billing">
    <AppInput
      :id="`${idPrefix}-raison-sociale`"
      v-model="model.raisonSociale"
      :label="$t('clients.field_name')"
      :required="requireName"
    />
    <AppInput
      :id="`${idPrefix}-adresse`"
      v-model="model.adresse"
      :label="$t('org.address')"
    />
    <div class="client-billing__row">
      <AppInput
        :id="`${idPrefix}-adresse-numero`"
        v-model="model.adresseNumero"
        :label="$t('org.address_number')"
      />
      <AppInput
        :id="`${idPrefix}-adresse-boite`"
        v-model="model.adresseBoite"
        :label="addressBoxLabel"
      />
    </div>
    <div class="client-billing__row">
      <AppInput
        :id="`${idPrefix}-code-postal`"
        v-model="model.codePostal"
        :label="$t('org.postal_code')"
      />
      <AppInput
        :id="`${idPrefix}-ville`"
        v-model="model.ville"
        :label="$t('org.city')"
      />
    </div>
    <div class="client-billing__field">
      <label :for="`${idPrefix}-pays`" class="client-billing__label">{{ $t('org.country') }}</label>
      <select :id="`${idPrefix}-pays`" v-model="model.pays" class="client-billing__select">
        <option value="">{{ $t('clients.country_unset') }}</option>
        <option value="FR">{{ $t('org.country_fr') }}</option>
        <option value="BE">{{ $t('org.country_be') }}</option>
        <option value="MG">{{ $t('org.country_mg') }}</option>
        <option value="MA">{{ $t('org.country_ma') }}</option>
        <option value="TN">{{ $t('org.country_tn') }}</option>
        <option value="CA">{{ $t('org.country_ca') }}</option>
      </select>
    </div>
    <div>
      <AppInput
        :id="`${idPrefix}-siret`"
        v-model="model.siret"
        :label="registryLabel"
        :placeholder="registryPlaceholder"
      />
      <p class="client-billing__hint">{{ registryHint }}</p>
    </div>
    <AppInput
      :id="`${idPrefix}-tva`"
      v-model="model.tva"
      :label="$t('clients.field_tva')"
    />
  </div>
</template>

<script setup lang="ts">
import {
  useBillingCountryLabels,
  type ClientBillingFields
} from '~/composables/useBillingCountry'

withDefaults(
  defineProps<{
    idPrefix?: string
    requireName?: boolean
  }>(),
  {
    idPrefix: 'client',
    requireName: true
  }
)

const model = defineModel<ClientBillingFields>({ required: true })

const { registryLabel, registryPlaceholder, registryHint, addressBoxLabel } = useBillingCountryLabels(
  computed(() => model.value.pays)
)
</script>

<style scoped>
.client-billing {
  display: grid;
  gap: var(--kore-space-md);
}

.client-billing__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--kore-space-md);
}

.client-billing__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.client-billing__label {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}

.client-billing__select {
  padding: 0.75rem 1rem;
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
  color: var(--kore-text);
  background: var(--kore-bg-elevated);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
}

.client-billing__hint {
  margin: var(--kore-space-xs) 0 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

@media (max-width: 768px) {
  .client-billing__row {
    grid-template-columns: 1fr;
  }
}
</style>
