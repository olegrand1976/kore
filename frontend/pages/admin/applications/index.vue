<template>
  <div>
    <AppPageHeader :title="$t('applications.title')" :subtitle="$t('applications.subtitle')">
      <template #actions>
        <AppButton
          variant="secondary"
          size="sm"
          type="button"
          :disabled="!mergeSelectionReady"
          @click="openMerge"
        >
          <AppIcon name="merge" /> {{ $t('applications.merge_action') }}
        </AppButton>
        <AppButton
          variant="secondary"
          size="sm"
          type="button"
          :disabled="!taigaAvailable"
          @click="openTaigaImport"
        >
          <AppIcon name="cloud_download" /> {{ $t('applications.taiga_import_title') }}
        </AppButton>
        <AppButton variant="primary" size="sm" type="button" @click="openCreate">
          <AppIcon name="add" /> {{ $t('applications.create_title') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <p
      v-if="flash"
      class="apps-flash"
      :class="{ 'apps-flash--error': flashError }"
      role="status"
    >
      {{ flash }}
    </p>

    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('applications.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="forbidden" padding="lg">
      <AppEmptyState icon="lock" :title="$t('applications.forbidden')" />
    </AppCard>

    <template v-else>
      <AppListToolbar
        :filters="listFilters"
        :filter-values="filterValues"
        :sort-keys="sortKeys"
        :sort-key="sortKey"
        :sort-dir="sortDir"
        :has-active-filters="hasActiveFilters"
        @update:filter="setFilter"
        @update:sort-key="setSort($event)"
        @update:sort-dir="setSortDir"
        @reset="resetFilters"
      />
      <AppButton
        class="apps-merge-mobile"
        variant="secondary"
        size="sm"
        type="button"
        :disabled="!mergeSelectionReady"
        @click="openMerge"
      >
        <AppIcon name="merge" /> {{ $t('applications.merge_action') }}
      </AppButton>
      <AppCard padding="lg">
        <AppTable
          :columns="columns"
          :rows="displayRows"
          :empty-title="hasActiveFilters ? $t('common.list.no_results') : $t('applications.empty')"
          row-key="id"
        >
          <template #cell-select="{ row }">
            <label class="apps-select">
              <input
                type="checkbox"
                :checked="selectedMergeIds.includes(row.id as string)"
                :disabled="!canSelectForMerge(row.id as string)"
                @change="toggleMergeSelect(row.id as string)"
              />
              <span class="apps-form__sr-only">{{ row.libelle }}</span>
            </label>
          </template>
          <template #cell-libelle="{ row }">
            <div class="apps-libelle">
              <span>{{ row.libelle }}</span>
              <AppBadge v-if="row.taigaLinked" variant="gold">{{ $t('applications.taiga_badge') }}</AppBadge>
            </div>
          </template>
          <template #cell-active="{ value }">
            <AppBadge :variant="value ? 'success' : 'warning'">
              {{ value ? $t('applications.active') : $t('applications.inactive') }}
            </AppBadge>
          </template>
          <template #cell-actions="{ row }">
            <div class="apps-actions">
              <AppButton variant="ghost" size="sm" type="button" @click="openEdit(row)">
                {{ $t('common.edit') }}
              </AppButton>
              <AppButton
                v-if="row.active"
                variant="ghost"
                size="sm"
                type="button"
                :disabled="actionBusyId === row.id"
                @click="deactivateRow(row)"
              >
                {{ $t('applications.deactivate') }}
              </AppButton>
              <AppButton
                v-else
                variant="ghost"
                size="sm"
                type="button"
                :disabled="actionBusyId === row.id"
                @click="activateRow(row)"
              >
                {{ $t('applications.activate') }}
              </AppButton>
            </div>
          </template>
        </AppTable>
      </AppCard>
    </template>

    <AppModal
      v-model:open="showForm"
      width="lg"
      title-id="apps-form-title"
    >
      <form class="apps-form" novalidate @submit.prevent="submitForm">
        <h2 id="apps-form-title" class="apps-form__title">
          {{ editingId ? $t('applications.edit_title') : $t('applications.create_title') }}
        </h2>

        <div class="apps-form__body">
          <section
            v-if="!editingId && taigaAvailable"
            class="apps-form__section"
            aria-labelledby="apps-section-source"
          >
            <h3 id="apps-section-source">{{ $t('applications.taiga_source_title') }}</h3>
            <fieldset class="apps-form__radiogroup">
              <legend class="apps-form__sr-only">{{ $t('applications.taiga_source_title') }}</legend>
              <label class="apps-form__radio">
                <input v-model="createMode" type="radio" value="manual" />
                <span>{{ $t('applications.taiga_source_manual') }}</span>
              </label>
              <label class="apps-form__radio">
                <input v-model="createMode" type="radio" value="taiga" />
                <span>{{ $t('applications.taiga_source_taiga') }}</span>
              </label>
            </fieldset>
            <label v-if="createMode === 'taiga'" class="apps-form__field">
              <span>{{ $t('applications.taiga_project_label') }}</span>
              <select
                id="app-taiga-project"
                v-model="selectedTaigaProjectId"
                class="apps-form__input"
                :disabled="taigaProjectsLoading"
              >
                <option value="">{{ $t('applications.taiga_project_none') }}</option>
                <option v-for="p in taigaProjects" :key="p.id" :value="String(p.id)">
                  {{ p.name }} ({{ p.slug }})
                </option>
              </select>
              <span v-if="taigaProjectsLoading" class="apps-form__hint">{{ $t('applications.loading') }}</span>
              <span v-else-if="!taigaProjects.length" class="apps-form__hint">
                {{ $t('applications.taiga_no_projects') }}
              </span>
            </label>
          </section>

          <section
            v-if="editingId && taigaAvailable"
            class="apps-form__section"
            aria-labelledby="apps-section-taiga-link"
          >
            <h3 id="apps-section-taiga-link">{{ $t('applications.taiga_link_title') }}</h3>
            <p v-if="taigaLinkLoading" class="apps-form__hint">{{ $t('applications.loading') }}</p>
            <template v-else-if="existingTaigaLink">
              <p class="apps-form__hint">{{ taigaLinkDisplayName }}</p>
              <a
                v-if="taigaLinkExternalUrl"
                class="apps-form__link"
                :href="taigaLinkExternalUrl"
                target="_blank"
                rel="noopener noreferrer"
              >
                {{ $t('applications.taiga_link_open') }}
              </a>
              <p class="apps-form__hint">{{ $t('applications.taiga_link_readonly') }}</p>
            </template>
            <template v-else>
              <label class="apps-form__field">
                <span>{{ $t('applications.taiga_project_label') }}</span>
                <select
                  id="app-taiga-project-edit"
                  v-model="selectedTaigaProjectId"
                  class="apps-form__input"
                  :disabled="taigaProjectsLoading"
                >
                  <option value="">{{ $t('applications.taiga_project_none') }}</option>
                  <option v-for="p in taigaProjects" :key="p.id" :value="String(p.id)">
                    {{ p.name }} ({{ p.slug }})
                  </option>
                </select>
                <span v-if="taigaProjectsLoading" class="apps-form__hint">{{ $t('applications.loading') }}</span>
                <span v-else-if="!taigaProjects.length" class="apps-form__hint">
                  {{ $t('applications.taiga_no_projects') }}
                </span>
                <span v-else class="apps-form__hint">{{ $t('applications.taiga_link_on_save') }}</span>
              </label>
            </template>
          </section>

          <section class="apps-form__section" aria-labelledby="apps-section-identity">
            <h3 id="apps-section-identity">{{ $t('applications.section_identity') }}</h3>
            <AppInput
              id="app-libelle"
              v-model="form.libelle"
              :label="$t('applications.field_libelle')"
              required
            />
            <AppInput
              id="app-proprietaire"
              v-model="form.proprietaire"
              :label="$t('applications.field_proprietaire')"
            />
            <label class="apps-form__field">
              <span>{{ $t('applications.field_chef') }}</span>
              <select id="app-chef" v-model="form.chefUtilisateurId" class="apps-form__input">
                <option value="">{{ $t('applications.field_chef_none') }}</option>
                <option v-for="u in userOptions" :key="u.value" :value="u.value">{{ u.label }}</option>
              </select>
            </label>
            <label class="apps-form__field">
              <span>{{ $t('applications.field_methodology') }}</span>
              <select id="app-methodology" v-model="form.methodologyProfile" class="apps-form__input">
                <option value="psa">{{ $t('applications.methodology_psa') }}</option>
                <option value="agile_scrum">{{ $t('applications.methodology_scrum') }}</option>
                <option value="agile_kanban">{{ $t('applications.methodology_kanban') }}</option>
              </select>
              <span class="apps-form__hint">{{ $t('applications.field_methodology_hint') }}</span>
              <span class="apps-form__hint">{{ methodologyPreview }}</span>
            </label>
          </section>

          <section
            class="apps-form__section"
            :class="{ 'apps-form__section--invalid': sharesInvalid }"
            aria-labelledby="apps-section-shares"
            :aria-invalid="sharesInvalid ? 'true' : undefined"
          >
            <h3 id="apps-section-shares">{{ $t('applications.section_shares') }}</h3>
            <p class="apps-form__hint">{{ $t('applications.shares_hint') }}</p>
            <p class="apps-form__shares-summary" aria-live="polite">{{ formSharesSummary }}</p>

            <fieldset class="apps-form__checkgroup">
              <legend>{{ $t('applications.field_share_sites') }}</legend>
              <p v-if="!siteOptions.length" class="apps-form__hint">{{ $t('applications.shares_empty_sites') }}</p>
              <div v-else class="apps-form__checklist">
                <label v-for="s in siteOptions" :key="s.value" class="apps-form__check">
                  <input v-model="form.siteIds" type="checkbox" :value="s.value" />
                  <span>{{ s.label }}</span>
                </label>
              </div>
            </fieldset>

            <fieldset class="apps-form__checkgroup">
              <legend>{{ $t('applications.field_share_services') }}</legend>
              <p v-if="!serviceOptions.length" class="apps-form__hint">{{ $t('applications.shares_empty_services') }}</p>
              <div v-else class="apps-form__checklist">
                <label v-for="s in serviceOptions" :key="s.value" class="apps-form__check">
                  <input v-model="form.serviceIds" type="checkbox" :value="s.value" />
                  <span>{{ s.label }}</span>
                </label>
              </div>
            </fieldset>

            <fieldset class="apps-form__checkgroup">
              <legend>{{ $t('applications.field_share_equipes') }}</legend>
              <p v-if="!equipeShareOptions.length" class="apps-form__hint">{{ $t('applications.shares_empty_equipes') }}</p>
              <div v-else class="apps-form__checklist">
                <label v-for="e in equipeShareOptions" :key="e.value" class="apps-form__check">
                  <input v-model="form.equipeIds" type="checkbox" :value="e.value" />
                  <span>{{ e.label }}</span>
                </label>
              </div>
            </fieldset>
          </section>

          <section class="apps-form__section" aria-labelledby="apps-section-billing">
            <h3 id="apps-section-billing">{{ $t('applications.section_billing') }}</h3>
            <label class="apps-form__field">
              <span>{{ $t('applications.field_mode') }}</span>
              <select id="app-mode" v-model="form.modeFacturation" class="apps-form__input">
                <option value="temps_passe">{{ $t('applications.mode_temps_passe') }}</option>
                <option value="forfait">{{ $t('applications.mode_forfait') }}</option>
                <option value="non">{{ $t('applications.mode_non') }}</option>
              </select>
            </label>
            <AppInput
              id="app-default-tjm"
              v-model="defaultTjmModel"
              type="number"
              min="0"
              step="1"
              :label="$t('applications.field_default_tjm')"
            />
            <p class="apps-form__hint">{{ $t('applications.field_default_tjm_hint') }}</p>
            <label class="apps-form__toggle">
              <input v-model="form.uoActivee" type="checkbox" />
              <span>{{ $t('applications.field_uo') }}</span>
            </label>
          </section>

          <template v-if="editingId">
            <section class="apps-form__section" aria-labelledby="apps-section-equipes">
              <h3 id="apps-section-equipes">{{ $t('applications.section_equipes') }}</h3>
              <ul v-if="editEquipes.length" class="apps-form__list">
                <li v-for="e in editEquipes" :key="e.id">{{ e.label }}</li>
              </ul>
              <p v-else class="apps-form__hint">{{ $t('applications.empty_equipes') }}</p>
              <div class="apps-form__inline">
                <AppInput
                  id="app-new-equipe"
                  v-model="newEquipeLibelle"
                  :label="$t('applications.new_equipe_placeholder')"
                  :placeholder="$t('applications.new_equipe_placeholder')"
                />
                <AppButton
                  variant="secondary"
                  size="sm"
                  type="button"
                  :disabled="!newEquipeLibelle.trim()"
                  @click="addEquipe"
                >
                  {{ $t('applications.add_equipe') }}
                </AppButton>
              </div>
            </section>

            <section class="apps-form__section" aria-labelledby="apps-section-users">
              <h3 id="apps-section-users">{{ $t('applications.section_users') }}</h3>
              <p class="apps-form__hint">{{ $t('applications.users_hint') }}</p>
              <template v-if="editEquipes.length">
                <label class="apps-form__field">
                  <span>{{ $t('applications.membership_equipe') }}</span>
                  <select v-model="membershipEquipeId" class="apps-form__input">
                    <option v-for="e in editEquipes" :key="e.id" :value="e.id">{{ e.label }}</option>
                  </select>
                </label>
                <fieldset v-if="membershipEquipeId" class="apps-form__checkgroup">
                  <legend>{{ $t('applications.membership_users') }}</legend>
                  <div class="apps-form__checklist">
                    <label v-for="u in userOptions" :key="u.value" class="apps-form__check">
                      <input
                        type="checkbox"
                        :checked="userInMembershipEquipe(u.value)"
                        :disabled="membershipBusyId === u.value"
                        @change="toggleUserMembership(u.value, ($event.target as HTMLInputElement).checked)"
                      />
                      <span>{{ u.label }}</span>
                    </label>
                  </div>
                </fieldset>
              </template>
              <p v-else class="apps-form__hint">{{ $t('applications.membership_need_equipe') }}</p>
              <NuxtLink class="apps-form__link" to="/admin/users">{{ $t('applications.manage_users') }}</NuxtLink>
            </section>

            <section class="apps-form__section" aria-labelledby="apps-section-budgets">
              <h3 id="apps-section-budgets">{{ $t('applications.section_budgets') }}</h3>
              <ul v-if="editBudgets.length" class="apps-form__list">
                <li v-for="b in editBudgets" :key="b.id" class="apps-form__list-item">
                  <span>{{ b.label }}</span>
                  <AppBadge v-if="b.isDefault || b.id === form.budgetDefautId" variant="success">
                    {{ $t('applications.budget_default') }}
                  </AppBadge>
                </li>
              </ul>
              <p v-else class="apps-form__hint">{{ $t('applications.empty_budgets') }}</p>
              <label class="apps-form__field">
                <span>{{ $t('applications.field_budget_defaut') }}</span>
                <select v-model="form.budgetDefautId" class="apps-form__input">
                  <option value="">{{ $t('applications.field_budget_defaut_none') }}</option>
                  <option v-for="b in defaultBudgetOptions" :key="b.id" :value="b.id">{{ b.label }}</option>
                </select>
                <span class="apps-form__hint">{{ $t('applications.field_budget_defaut_hint') }}</span>
              </label>
              <AppButton variant="secondary" size="sm" type="button" @click="goCreateBudget">
                {{ $t('applications.create_budget') }}
              </AppButton>
            </section>
          </template>
        </div>

        <div class="apps-form__footer">
          <p v-if="formError" id="apps-form-error" class="apps-form__error" role="alert">{{ formError }}</p>
          <div class="apps-form__actions">
            <AppButton variant="ghost" type="button" @click="closeForm">{{ $t('common.cancel') }}</AppButton>
            <AppButton variant="primary" type="submit" :disabled="saving">
              {{ saving ? $t('common.saving') : $t('common.save') }}
            </AppButton>
          </div>
        </div>
      </form>
    </AppModal>

    <AppModal
      v-model:open="showTaigaImport"
      width="lg"
      title-id="apps-taiga-import-title"
    >
      <form class="apps-form" novalidate @submit.prevent="submitTaigaImport">
        <h2 id="apps-taiga-import-title" class="apps-form__title">
          {{ $t('applications.taiga_import_title') }}
        </h2>
        <div class="apps-form__body">
          <section class="apps-form__section" aria-labelledby="apps-taiga-import-projects">
            <h3 id="apps-taiga-import-projects">{{ $t('applications.taiga_import_projects') }}</h3>
            <p v-if="taigaProjectsLoading" class="apps-form__hint">{{ $t('applications.loading') }}</p>
            <p v-else-if="!taigaProjects.length" class="apps-form__hint">
              {{ $t('applications.taiga_no_projects') }}
            </p>
            <fieldset v-else class="apps-form__checkgroup apps-form__checklist--scroll">
              <legend class="apps-form__sr-only">{{ $t('applications.taiga_import_projects') }}</legend>
              <div class="apps-form__checklist">
                <label v-for="p in taigaProjects" :key="p.id" class="apps-form__check">
                  <input v-model="taigaImportSelected" type="checkbox" :value="p.id" />
                  <span>{{ p.name }} ({{ p.slug }})</span>
                </label>
              </div>
            </fieldset>
          </section>

          <section
            class="apps-form__section"
            :class="{ 'apps-form__section--invalid': taigaImportSharesInvalid }"
            aria-labelledby="apps-taiga-import-shares"
          >
            <h3 id="apps-taiga-import-shares">{{ $t('applications.section_shares') }}</h3>
            <p class="apps-form__hint">{{ $t('applications.shares_hint') }}</p>
            <p class="apps-form__shares-summary" aria-live="polite">{{ taigaImportSharesSummary }}</p>
            <fieldset class="apps-form__checkgroup">
              <legend>{{ $t('applications.field_share_sites') }}</legend>
              <div v-if="siteOptions.length" class="apps-form__checklist">
                <label v-for="s in siteOptions" :key="s.value" class="apps-form__check">
                  <input v-model="taigaImportForm.siteIds" type="checkbox" :value="s.value" />
                  <span>{{ s.label }}</span>
                </label>
              </div>
            </fieldset>
            <fieldset class="apps-form__checkgroup">
              <legend>{{ $t('applications.field_share_services') }}</legend>
              <div v-if="serviceOptions.length" class="apps-form__checklist">
                <label v-for="s in serviceOptions" :key="s.value" class="apps-form__check">
                  <input v-model="taigaImportForm.serviceIds" type="checkbox" :value="s.value" />
                  <span>{{ s.label }}</span>
                </label>
              </div>
            </fieldset>
            <fieldset class="apps-form__checkgroup">
              <legend>{{ $t('applications.field_share_equipes') }}</legend>
              <div v-if="equipeShareOptions.length" class="apps-form__checklist">
                <label v-for="e in equipeShareOptions" :key="e.value" class="apps-form__check">
                  <input v-model="taigaImportForm.equipeIds" type="checkbox" :value="e.value" />
                  <span>{{ e.label }}</span>
                </label>
              </div>
            </fieldset>
          </section>

          <section class="apps-form__section" aria-labelledby="apps-taiga-import-billing">
            <h3 id="apps-taiga-import-billing">{{ $t('applications.section_billing') }}</h3>
            <label class="apps-form__field">
              <span>{{ $t('applications.field_mode') }}</span>
              <select v-model="taigaImportForm.modeFacturation" class="apps-form__input">
                <option value="temps_passe">{{ $t('applications.mode_temps_passe') }}</option>
                <option value="forfait">{{ $t('applications.mode_forfait') }}</option>
                <option value="non">{{ $t('applications.mode_non') }}</option>
              </select>
            </label>
            <label class="apps-form__field">
              <span>{{ $t('applications.field_methodology') }}</span>
              <select v-model="taigaImportForm.methodologyProfile" class="apps-form__input">
                <option value="psa">{{ $t('applications.methodology_psa') }}</option>
                <option value="agile_scrum">{{ $t('applications.methodology_scrum') }}</option>
                <option value="agile_kanban">{{ $t('applications.methodology_kanban') }}</option>
              </select>
            </label>
          </section>
        </div>
        <div class="apps-form__footer">
          <p v-if="taigaImportError" class="apps-form__error" role="alert">{{ taigaImportError }}</p>
          <div class="apps-form__actions">
            <AppButton variant="ghost" type="button" @click="closeTaigaImport">
              {{ $t('common.cancel') }}
            </AppButton>
            <AppButton variant="primary" type="submit" :disabled="taigaImportSaving">
              {{ taigaImportSaving ? $t('common.saving') : $t('applications.taiga_import_submit') }}
            </AppButton>
          </div>
        </div>
      </form>
    </AppModal>

    <AppModal
      v-model:open="showMerge"
      width="lg"
      title-id="apps-merge-title"
    >
      <form class="apps-form" novalidate @submit.prevent="submitMerge">
        <h2 id="apps-merge-title" class="apps-form__title">
          {{ $t('applications.merge_title') }}
        </h2>
        <div class="apps-form__body">
          <p
            v-if="mergeMethodologyMismatch"
            class="apps-merge-warning"
            role="status"
          >
            {{ $t('applications.merge_methodology_warning') }}
          </p>
          <p v-if="mergeReferenceLocked" class="apps-form__hint">
            {{ $t('applications.merge_reference_taiga_locked') }}
          </p>
          <section class="apps-form__section" aria-labelledby="apps-merge-reference">
            <h3 id="apps-merge-reference">{{ $t('applications.merge_reference_label') }}</h3>
            <fieldset class="apps-form__radiogroup">
              <legend class="apps-form__sr-only">{{ $t('applications.merge_reference_label') }}</legend>
              <label
                v-for="row in mergeSelectedRows"
                :key="row.id"
                class="apps-form__radio"
              >
                <input
                  v-model="mergeReferenceId"
                  type="radio"
                  :value="row.id"
                  :disabled="mergeReferenceLocked"
                />
                <span>
                  {{ row.libelle }}
                  <AppBadge v-if="row.taigaLinked" variant="gold" class="apps-merge-badge">
                    {{ $t('applications.taiga_badge') }}
                  </AppBadge>
                </span>
              </label>
            </fieldset>
          </section>
          <p class="apps-form__hint">{{ $t('applications.merge_confirm') }}</p>
        </div>
        <div class="apps-form__footer">
          <p v-if="mergeError" class="apps-form__error" role="alert">{{ mergeError }}</p>
          <div class="apps-form__actions">
            <AppButton variant="ghost" type="button" @click="closeMerge">
              {{ $t('common.cancel') }}
            </AppButton>
            <AppButton variant="primary" type="submit" :disabled="mergeSaving">
              {{ mergeSaving ? $t('common.saving') : $t('applications.merge_action') }}
            </AppButton>
          </div>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import type { MethodologyProfile } from '~/composables/useMethodologyTerms'
import { applyTextSearch, useListControls } from '~/composables/useListControls'
import {
  coerceBudgetDefautId,
  defaultBudgetsForApplication,
  isDefaultBudgetType
} from '~/composables/useApplications'
import {
  defaultMergeReferenceId,
  canMergeApplications,
  hasMergeMethodologyMismatch,
  isMergeReferenceLocked,
  mergeAbsorbedId,
  type MergeApplicationRow
} from '~/utils/applicationMerge'
import type { OrgEquipe } from '~/composables/useOrganisation'
import type { BudgetItem } from '~/composables/useBudget'
import type { OrgUserSummary } from '~/composables/useUsers'

definePageMeta({ layout: 'default', middleware: 'admin' })

const { t } = useI18n()
const route = useRoute()
const { extractFetchError, extractFetchErrorCode } = useApiError()
const { apiFetch } = useApiFetch()
const {
  list,
  create,
  update,
  deactivate,
  activate,
  merge,
  pickAppId,
  pickAppLabel,
  pickAppClient,
  pickAppActive,
  pickAppSiteIds,
  pickAppServiceIds,
  pickAppEquipeIds,
  pickAppMode,
  pickAppChefId,
  pickAppBudgetDefautId,
  pickAppMethodologyProfile
} = useApplications()
const { listSites, listServices, listEquipes, createEquipe, orgId, orgLabel } = useOrganisation()
const { list: listUsers, update: updateUser, pickUserId, pickUserLogin, pickUserEquipeIds } = useUsers()
const { list: listBudgets, pickId: pickBudgetId } = useBudget()

type AppRow = {
  id: string
  libelle: string
  serviceIds: string[]
  sharesLabel: string
  proprietaire: string
  mode: string
  modeLabel: string
  defaultTjmCents: number
  uoActivee: boolean
  active: boolean
  equipeCount: number
  chefUtilisateurId: string
  budgetDefautId: string
  siteIds: string[]
  equipeIds: string[]
  methodologyProfile: MethodologyProfile
  taigaLinked: boolean
}

const rows = ref<AppRow[]>([])
const pending = ref(true)
const forbidden = ref(false)
const saving = ref(false)
const showForm = ref(false)
const editingId = ref('')
const formError = ref('')
const sharesInvalid = ref(false)
const flash = ref('')
const flashError = ref(false)
const actionBusyId = ref('')
const membershipBusyId = ref('')
const membershipEquipeId = ref('')
const newEquipeLibelle = ref('')

type CreateMode = 'manual' | 'taiga'

type TaigaApplicationLink = {
  externalId?: string
  ExternalID?: string
  externalUrl?: string
  ExternalURL?: string
  metadata?: Record<string, unknown>
  Metadata?: Record<string, unknown>
}

const createMode = ref<CreateMode>('manual')
const taigaAvailable = ref(true)
const taigaProjects = ref<{ id: number; name: string; slug: string }[]>([])
const taigaProjectsLoading = ref(false)
const selectedTaigaProjectId = ref('')
const existingTaigaLink = ref<TaigaApplicationLink | null>(null)
const taigaLinkLoading = ref(false)
const showTaigaImport = ref(false)
const taigaImportSelected = ref<number[]>([])
const taigaImportSaving = ref(false)
const taigaImportError = ref('')
const taigaImportSharesInvalid = ref(false)
const taigaImportForm = reactive({
  siteIds: [] as string[],
  serviceIds: [] as string[],
  equipeIds: [] as string[],
  modeFacturation: 'temps_passe',
  methodologyProfile: 'psa' as MethodologyProfile,
  proprietaire: '',
  defaultTjmEuros: 0,
  uoActivee: false,
  chefUtilisateurId: ''
})
const taigaLinkedIds = ref<Set<string>>(new Set())
const selectedMergeIds = ref<string[]>([])
const showMerge = ref(false)
const mergeReferenceId = ref('')
const mergeSaving = ref(false)
const mergeError = ref('')

const serviceOptions = ref<{ value: string; label: string }[]>([])
const siteOptions = ref<{ value: string; label: string }[]>([])
const equipeShareOptions = ref<{ value: string; label: string }[]>([])
const equipes = ref<OrgEquipe[]>([])
const users = ref<OrgUserSummary[]>([])
const budgets = ref<BudgetItem[]>([])

const form = reactive({
  siteIds: [] as string[],
  serviceIds: [] as string[],
  equipeIds: [] as string[],
  libelle: '',
  proprietaire: '',
  modeFacturation: 'temps_passe',
  defaultTjmEuros: 0,
  uoActivee: false,
  chefUtilisateurId: '',
  budgetDefautId: '',
  methodologyProfile: 'psa' as MethodologyProfile
})

const methodologyTerms = computed(() => useMethodologyTerms(form.methodologyProfile).value)
const methodologyPreview = computed(() =>
  t('applications.methodology_preview', {
    workItem: methodologyTerms.value.workItem,
    backlog: methodologyTerms.value.backlog
  })
)

const modeLabel = (mode: string) => {
  if (mode === 'non') return t('applications.mode_non')
  if (mode === 'forfait') return t('applications.mode_forfait')
  return t('applications.mode_temps_passe')
}

const sharesLabel = (sites: number, services: number, equipesCount: number) =>
  t('applications.shares_summary', { sites, services, equipes: equipesCount })

const columns = computed(() => [
  { key: 'select', label: t('applications.col_select') },
  { key: 'libelle', label: t('applications.col_libelle') },
  { key: 'sharesLabel', label: t('applications.col_shares') },
  { key: 'proprietaire', label: t('applications.col_proprietaire') },
  { key: 'modeLabel', label: t('applications.col_mode') },
  { key: 'equipeCount', label: t('applications.col_equipes') },
  { key: 'active', label: t('applications.col_status') },
  { key: 'actions', label: t('common.actions') }
])

const listFilters = computed(() => ({
  q: {
    type: 'search' as const,
    label: t('common.list.search'),
    placeholder: t('applications.search_placeholder'),
    match: (item: AppRow, query: string) =>
      applyTextSearch(query, item.libelle, item.proprietaire, item.sharesLabel)
  },
  active: {
    type: 'select' as const,
    label: t('applications.col_status'),
    options: [
      { value: 'true', label: t('applications.active') },
      { value: 'false', label: t('applications.inactive') }
    ],
    match: (item: AppRow, value: string) => {
      if (!value) return true
      return String(item.active) === value
    }
  },
  serviceId: {
    type: 'select' as const,
    label: t('applications.col_service'),
    options: serviceOptions.value,
    match: (item: AppRow, value: string) => !value || item.serviceIds.includes(value)
  }
}))

const sortKeys = computed(() => [
  {
    key: 'libelle',
    label: t('applications.col_libelle'),
    type: 'string' as const,
    accessor: (item: AppRow) => item.libelle
  },
  {
    key: 'sharesLabel',
    label: t('applications.col_shares'),
    type: 'string' as const,
    accessor: (item: AppRow) => item.sharesLabel
  },
  {
    key: 'equipeCount',
    label: t('applications.col_equipes'),
    type: 'number' as const,
    accessor: (item: AppRow) => item.equipeCount
  }
])

const {
  filterValues,
  sortKey,
  sortDir,
  hasActiveFilters,
  setFilter,
  setSort,
  setSortDir,
  resetFilters,
  sortedItems
} = useListControls(rows, {
  storageKey: 'admin-applications',
  filters: listFilters,
  sortKeys,
  defaultSort: { key: 'libelle', dir: 'asc' }
})

const displayRows = computed(() => sortedItems.value)

const mergeSelectedRows = computed((): MergeApplicationRow[] =>
  rows.value
    .filter((row) => selectedMergeIds.value.includes(row.id))
    .map((row) => ({
      id: row.id,
      libelle: row.libelle,
      methodologyProfile: row.methodologyProfile,
      taigaLinked: row.taigaLinked
    }))
)

const mergeReferenceLocked = computed(() => isMergeReferenceLocked(mergeSelectedRows.value))

const mergeSelectionReady = computed(() => canMergeApplications(mergeSelectedRows.value))

const mergeMethodologyMismatch = computed(() =>
  hasMergeMethodologyMismatch(mergeSelectedRows.value, mergeReferenceId.value)
)

const canSelectForMerge = (id: string) =>
  selectedMergeIds.value.includes(id) || selectedMergeIds.value.length < 2

const toggleMergeSelect = (id: string) => {
  if (selectedMergeIds.value.includes(id)) {
    selectedMergeIds.value = selectedMergeIds.value.filter((item) => item !== id)
    return
  }
  if (selectedMergeIds.value.length >= 2) return
  selectedMergeIds.value = [...selectedMergeIds.value, id]
}

const openMerge = () => {
  mergeError.value = ''
  if (mergeSelectedRows.value.length !== 2) {
    mergeError.value = t('applications.merge_select_two')
    return
  }
  if (!canMergeApplications(mergeSelectedRows.value)) {
    mergeError.value = t('applications.merge_error_both_taiga')
    return
  }
  mergeReferenceId.value = defaultMergeReferenceId(mergeSelectedRows.value)
  showMerge.value = true
}

const closeMerge = () => {
  showMerge.value = false
  mergeError.value = ''
}

const submitMerge = async () => {
  mergeError.value = ''
  const selected = mergeSelectedRows.value
  if (selected.length !== 2 || !mergeReferenceId.value) {
    mergeError.value = t('applications.merge_select_two')
    return
  }
  const absorbedId = mergeAbsorbedId(selected, mergeReferenceId.value)
  if (!absorbedId) {
    mergeError.value = t('applications.merge_select_two')
    return
  }
  mergeSaving.value = true
  try {
    await merge(absorbedId, mergeReferenceId.value)
    flash.value = t('applications.merge_success')
    flashError.value = false
    selectedMergeIds.value = []
    closeMerge()
    await loadAll()
  } catch (err) {
    const code = extractFetchErrorCode(err)
    if (code === 'APPLICATIONS_MERGE_BOTH_TAIGA') {
      mergeError.value = t('applications.merge_error_both_taiga')
    } else if (code === 'APPLICATIONS_MERGE_ACTIVE_SPRINT') {
      mergeError.value = t('applications.merge_error_active_sprint')
    } else if (code === 'APPLICATIONS_MERGE_METHODOLOGY') {
      mergeError.value = t('applications.merge_error_methodology')
    } else if (code === 'APPLICATIONS_MERGE_DEFAULT_BUDGET') {
      mergeError.value = t('applications.merge_error_default_budget')
    } else {
      mergeError.value = extractFetchError(err)
    }
  } finally {
    mergeSaving.value = false
  }
}

const formSharesSummary = computed(() =>
  sharesLabel(form.siteIds.length, form.serviceIds.length, form.equipeIds.length)
)

const taigaImportSharesSummary = computed(() =>
  sharesLabel(
    taigaImportForm.siteIds.length,
    taigaImportForm.serviceIds.length,
    taigaImportForm.equipeIds.length
  )
)

const taigaLinkExternalUrl = computed(() => {
  const raw = existingTaigaLink.value?.externalUrl ?? existingTaigaLink.value?.ExternalURL ?? ''
  return typeof raw === 'string' ? raw.trim() : ''
})

const taigaLinkDisplayName = computed(() => {
  const link = existingTaigaLink.value
  if (!link) return ''
  const meta = (link.metadata ?? link.Metadata ?? {}) as Record<string, string>
  const name = meta.name?.trim()
  const slug = meta.slug?.trim()
  if (name && slug) return `${name} (${slug})`
  if (name) return name
  if (slug) return slug
  return link.externalId ?? link.ExternalID ?? ''
})

let taigaLinkLoadGeneration = 0

const loadTaigaLinkForEdit = async (applicationId: string) => {
  const generation = ++taigaLinkLoadGeneration
  taigaLinkLoading.value = true
  existingTaigaLink.value = null
  selectedTaigaProjectId.value = ''
  try {
    const res = await apiFetch<{ data?: TaigaApplicationLink }>(
      `/api/integrations/taiga/links/by-application/${applicationId}`
    )
    if (generation !== taigaLinkLoadGeneration) return
    existingTaigaLink.value = res?.data ?? null
  } catch (err) {
    if (generation !== taigaLinkLoadGeneration) return
    if ((err as { statusCode?: number })?.statusCode === 404) {
      existingTaigaLink.value = null
      if (taigaAvailable.value) {
        try {
          await loadTaigaProjects()
        } catch (loadErr) {
          if (generation !== taigaLinkLoadGeneration) return
          formError.value = extractFetchError(loadErr)
        }
      }
      return
    }
    if ((err as { statusCode?: number })?.statusCode === 503) {
      taigaAvailable.value = false
      return
    }
    formError.value = extractFetchError(err)
  } finally {
    if (generation === taigaLinkLoadGeneration) {
      taigaLinkLoading.value = false
    }
  }
}

const loadTaigaProjects = async () => {
  taigaProjectsLoading.value = true
  try {
    const res = await apiFetch<{ data?: { id: number; name: string; slug: string }[] }>(
      '/api/integrations/taiga/projects/unlinked'
    )
    taigaProjects.value = res?.data ?? []
    taigaAvailable.value = true
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 503) {
      taigaAvailable.value = false
      taigaProjects.value = []
      return
    }
    throw err
  } finally {
    taigaProjectsLoading.value = false
  }
}

watch(createMode, async (mode) => {
  if (mode === 'taiga' && !taigaProjects.value.length && taigaAvailable.value) {
    try {
      await loadTaigaProjects()
    } catch (err) {
      formError.value = extractFetchError(err)
    }
  }
})

watch(selectedTaigaProjectId, (id) => {
  if (!id || form.libelle.trim()) return
  const project = taigaProjects.value.find((p) => String(p.id) === id)
  if (project) form.libelle = project.name
})

const openTaigaImport = async () => {
  taigaImportError.value = ''
  taigaImportSharesInvalid.value = false
  taigaImportSelected.value = []
  taigaImportForm.siteIds = []
  taigaImportForm.serviceIds = []
  taigaImportForm.equipeIds = []
  taigaImportForm.modeFacturation = form.modeFacturation
  taigaImportForm.methodologyProfile = form.methodologyProfile
  taigaImportForm.proprietaire = form.proprietaire
  taigaImportForm.defaultTjmEuros = form.defaultTjmEuros
  taigaImportForm.uoActivee = form.uoActivee
  taigaImportForm.chefUtilisateurId = form.chefUtilisateurId
  showTaigaImport.value = true
  try {
    await loadTaigaProjects()
  } catch (err) {
    taigaImportError.value = extractFetchError(err)
  }
}

const closeTaigaImport = () => {
  showTaigaImport.value = false
  taigaImportError.value = ''
  taigaImportSharesInvalid.value = false
}

const submitTaigaImport = async () => {
  taigaImportError.value = ''
  const hasShare =
    taigaImportForm.siteIds.length > 0 ||
    taigaImportForm.serviceIds.length > 0 ||
    taigaImportForm.equipeIds.length > 0
  taigaImportSharesInvalid.value = !hasShare
  if (!taigaImportSelected.value.length || !hasShare) {
    taigaImportError.value = t('applications.taiga_import_validation')
    return
  }
  taigaImportSaving.value = true
  try {
    const res = await apiFetch<{
      data?: {
        created?: { applicationId: string; libelle: string; taigaProjectId: number }[]
        errors?: { taigaProjectId: number; message: string }[]
      }
    }>('/api/integrations/taiga/applications/import', {
      method: 'POST',
      body: {
        projects: taigaImportSelected.value.map((id) => {
          const project = taigaProjects.value.find((p) => p.id === id)
          return { taigaProjectId: id, libelle: project?.name ?? '' }
        }),
        proprietaire: taigaImportForm.proprietaire.trim(),
        modeFacturation: taigaImportForm.modeFacturation,
        defaultTjmCents: Math.max(0, Math.round((taigaImportForm.defaultTjmEuros || 0) * 100)),
        uoActivee: taigaImportForm.uoActivee,
        chefUtilisateurId: taigaImportForm.chefUtilisateurId || undefined,
        siteIds: taigaImportForm.siteIds,
        serviceIds: taigaImportForm.serviceIds,
        equipeIds: taigaImportForm.equipeIds,
        methodologyProfile: taigaImportForm.methodologyProfile
      }
    })
    const created = res?.data?.created?.length ?? 0
    const errors = res?.data?.errors?.length ?? 0
    flash.value =
      errors > 0
        ? t('applications.taiga_import_partial', { created, errors })
        : t('applications.taiga_import_done', { count: created })
    flashError.value = errors > 0 && created === 0
    closeTaigaImport()
    await loadAll()
    await loadTaigaProjects()
  } catch (err) {
    taigaImportError.value = extractFetchError(err)
  } finally {
    taigaImportSaving.value = false
  }
}

const defaultTjmModel = computed({
  get: () => String(form.defaultTjmEuros || 0),
  set: (value: string) => {
    form.defaultTjmEuros = Math.max(0, Math.round(Number(value) || 0))
  }
})

const userOptions = computed(() =>
  users.value.map((u) => ({ value: pickUserId(u), label: pickUserLogin(u) }))
)

const editEquipes = computed(() => {
  if (!editingId.value) return []
  const shared = new Set(
    rows.value.find((r) => r.id === editingId.value)?.equipeIds ?? []
  )
  return equipes.value
    .filter(
      (e) =>
        (e.applicationId ?? e.ApplicationID) === editingId.value || shared.has(orgId(e))
    )
    .map((e) => ({ id: orgId(e), label: orgLabel(e) }))
})

watch(
  () => [form.siteIds.length, form.serviceIds.length, form.equipeIds.length],
  () => {
    if (
      sharesInvalid.value &&
      (form.siteIds.length > 0 || form.serviceIds.length > 0 || form.equipeIds.length > 0)
    ) {
      sharesInvalid.value = false
    }
  }
)

watch(
  editEquipes,
  (list) => {
    if (!list.length) {
      membershipEquipeId.value = ''
      return
    }
    if (!list.some((e) => e.id === membershipEquipeId.value)) {
      membershipEquipeId.value = list[0].id
    }
  },
  { immediate: true }
)

const userInMembershipEquipe = (userId: string) => {
  const user = users.value.find((u) => pickUserId(u) === userId)
  if (!user || !membershipEquipeId.value) return false
  return pickUserEquipeIds(user).includes(membershipEquipeId.value)
}

const toggleUserMembership = async (userId: string, checked: boolean) => {
  const equipeId = membershipEquipeId.value
  if (!equipeId) return
  const user = users.value.find((u) => pickUserId(u) === userId)
  if (!user) return
  const current = pickUserEquipeIds(user)
  const next = checked
    ? Array.from(new Set([...current, equipeId]))
    : current.filter((id) => id !== equipeId)
  membershipBusyId.value = userId
  try {
    await updateUser(userId, { equipeIds: next })
    users.value = await listUsers()
    flash.value = t('applications.membership_updated')
    flashError.value = false
  } catch (err) {
    formError.value = extractFetchError(err)
  } finally {
    membershipBusyId.value = ''
  }
}

const editBudgets = computed(() => {
  if (!editingId.value) return []
  return budgets.value
    .filter((b) => (b.applicationId ?? b.ApplicationID) === editingId.value)
    .map((b) => {
      const type = String(b.type ?? b.Type ?? '').toLowerCase()
      return {
        id: pickBudgetId(b),
        label: type || pickBudgetId(b),
        isDefault: isDefaultBudgetType(type)
      }
    })
})

const defaultBudgetOptions = computed(() => {
  if (!editingId.value) return []
  return defaultBudgetsForApplication(budgets.value, editingId.value).map((b) => {
    const type = String(b.type ?? b.Type ?? '').toLowerCase()
    const id = pickBudgetId(b)
    const short = id.length > 8 ? id.slice(0, 8) : id
    return {
      id,
      label: type ? `${type} (${short})` : short
    }
  })
})

/** Drop stale FK values that are not a default-type budget for this app. */
const sanitizeFormBudgetDefaut = () => {
  if (!editingId.value) return
  form.budgetDefautId = coerceBudgetDefautId(
    form.budgetDefautId,
    defaultBudgetOptions.value.map((b) => b.id)
  )
}

const loadAll = async () => {
  pending.value = true
  forbidden.value = false
  try {
    const [siteList, svcList, appList, equipeList, userList, budgetList] = await Promise.all([
      listSites(),
      listServices(),
      list({ active: 'all' }),
      listEquipes(),
      listUsers(),
      listBudgets()
    ])
    siteOptions.value = siteList.map((s) => ({
      value: orgId(s),
      label: orgLabel(s) || orgId(s)
    }))
    serviceOptions.value = svcList.map((s) => ({
      value: orgId(s),
      label: orgLabel(s) || orgId(s)
    }))
    equipeShareOptions.value = equipeList.map((e) => ({
      value: orgId(e),
      label: orgLabel(e) || orgId(e)
    }))
    equipes.value = equipeList
    users.value = userList
    budgets.value = budgetList
    try {
      const taigaRes = await apiFetch<{ data?: { applicationIds?: string[] } }>(
        '/api/integrations/taiga/links/applications'
      )
      taigaLinkedIds.value = new Set(taigaRes?.data?.applicationIds ?? [])
    } catch {
      taigaLinkedIds.value = new Set()
    }
    rows.value = appList.map((app) => {
      const id = pickAppId(app)
      const serviceIds = pickAppServiceIds(app)
      const siteIds = pickAppSiteIds(app)
      const equipeIds = pickAppEquipeIds(app)
      const mode = pickAppMode(app)
      return {
        id,
        libelle: pickAppLabel(app),
        serviceIds,
        siteIds,
        equipeIds,
        sharesLabel: sharesLabel(siteIds.length, serviceIds.length, equipeIds.length),
        proprietaire: pickAppClient(app),
        mode,
        modeLabel: modeLabel(mode),
        defaultTjmCents: Number(app.defaultTjmCents ?? app.DefaultTJMCents ?? 0),
        uoActivee: app.uoActivee ?? app.UOActivee ?? false,
        active: pickAppActive(app),
        equipeCount: equipeList.filter(
          (e) =>
            (e.applicationId ?? e.ApplicationID) === id || equipeIds.includes(orgId(e))
        ).length,
        chefUtilisateurId: pickAppChefId(app),
        budgetDefautId: pickAppBudgetDefautId(app),
        methodologyProfile: pickAppMethodologyProfile(app),
        taigaLinked: taigaLinkedIds.value.has(id)
      }
    })
    sanitizeFormBudgetDefaut()
  } catch (err) {
    if ((err as { statusCode?: number })?.statusCode === 403) {
      forbidden.value = true
      rows.value = []
    } else {
      flash.value = extractFetchError(err)
      flashError.value = true
    }
  } finally {
    pending.value = false
  }
}

const openCreate = () => {
  editingId.value = ''
  createMode.value = 'manual'
  selectedTaigaProjectId.value = ''
  existingTaigaLink.value = null
  taigaLinkLoading.value = false
  form.siteIds = []
  form.serviceIds = []
  form.equipeIds = []
  form.libelle = ''
  form.proprietaire = ''
  form.modeFacturation = 'temps_passe'
  form.defaultTjmEuros = 0
  form.uoActivee = false
  form.chefUtilisateurId = ''
  form.budgetDefautId = ''
  form.methodologyProfile = 'psa'
  formError.value = ''
  sharesInvalid.value = false
  newEquipeLibelle.value = ''
  showForm.value = true
}

const openEdit = async (row: AppRow) => {
  editingId.value = row.id
  selectedTaigaProjectId.value = ''
  existingTaigaLink.value = null
  taigaLinkLoading.value = taigaAvailable.value
  form.siteIds = [...row.siteIds]
  form.serviceIds = [...row.serviceIds]
  form.equipeIds = [...row.equipeIds]
  form.libelle = row.libelle
  form.proprietaire = row.proprietaire
  form.modeFacturation = row.mode || 'temps_passe'
  form.defaultTjmEuros = Math.round((row.defaultTjmCents || 0) / 100)
  form.uoActivee = row.uoActivee
  form.chefUtilisateurId = row.chefUtilisateurId
  form.budgetDefautId = row.budgetDefautId
  form.methodologyProfile = row.methodologyProfile
  formError.value = ''
  sharesInvalid.value = false
  newEquipeLibelle.value = ''
  showForm.value = true
  sanitizeFormBudgetDefaut()
  if (taigaAvailable.value) {
    await loadTaigaLinkForEdit(row.id)
  } else {
    taigaLinkLoading.value = false
  }
}

const closeForm = () => {
  taigaLinkLoadGeneration++
  showForm.value = false
  editingId.value = ''
  existingTaigaLink.value = null
  taigaLinkLoading.value = false
  formError.value = ''
  sharesInvalid.value = false
}

const submitForm = async () => {
  formError.value = ''
  const hasShare =
    form.siteIds.length > 0 || form.serviceIds.length > 0 || form.equipeIds.length > 0
  sharesInvalid.value = !hasShare
  if (!editingId.value && createMode.value === 'taiga' && !selectedTaigaProjectId.value) {
    formError.value = t('applications.taiga_project_required')
    return
  }
  if (!form.libelle.trim() || !hasShare) {
    formError.value = t('applications.validation_required')
    return
  }
  saving.value = true
  try {
    const chef = form.chefUtilisateurId || null
    if (editingId.value) {
      const originalBudget =
        rows.value.find((r) => r.id === editingId.value)?.budgetDefautId ?? ''
      const nextBudget = form.budgetDefautId || null
      const body: {
        libelle: string
        proprietaire: string
        modeFacturation: string
        defaultTjmCents: number
        uoActivee: boolean
        chefUtilisateurId: string | null
        siteIds: string[]
        serviceIds: string[]
        equipeIds: string[]
        budgetDefautId?: string | null
        methodologyProfile?: MethodologyProfile
        taigaProjectId?: number
      } = {
        libelle: form.libelle.trim(),
        proprietaire: form.proprietaire.trim(),
        modeFacturation: form.modeFacturation,
        defaultTjmCents: Math.max(0, Math.round((form.defaultTjmEuros || 0) * 100)),
        uoActivee: form.uoActivee,
        chefUtilisateurId: chef,
        siteIds: form.siteIds,
        serviceIds: form.serviceIds,
        equipeIds: form.equipeIds,
        methodologyProfile: form.methodologyProfile
      }
      if ((form.budgetDefautId || '') !== originalBudget) {
        body.budgetDefautId = nextBudget
      }
      if (!existingTaigaLink.value && selectedTaigaProjectId.value) {
        body.taigaProjectId = Number(selectedTaigaProjectId.value)
      }
      await update(editingId.value, body)
      flash.value = t('applications.updated')
    } else {
      const body: Record<string, unknown> = {
        siteIds: form.siteIds,
        serviceIds: form.serviceIds,
        equipeIds: form.equipeIds,
        libelle: form.libelle.trim(),
        proprietaire: form.proprietaire.trim(),
        modeFacturation: form.modeFacturation,
        defaultTjmCents: Math.max(0, Math.round((form.defaultTjmEuros || 0) * 100)),
        uoActivee: form.uoActivee,
        chefUtilisateurId: chef || undefined,
        methodologyProfile: form.methodologyProfile
      }
      if (createMode.value === 'taiga' && selectedTaigaProjectId.value) {
        body.taigaProjectId = Number(selectedTaigaProjectId.value)
      }
      await create(body as Parameters<typeof create>[0])
      flash.value = t('applications.created')
    }
    flashError.value = false
    closeForm()
    await loadAll()
  } catch (err) {
    formError.value = extractFetchError(err)
  } finally {
    saving.value = false
  }
}

const addEquipe = async () => {
  if (!editingId.value || !newEquipeLibelle.value.trim()) return
  try {
    await createEquipe({
      applicationId: editingId.value,
      libelle: newEquipeLibelle.value.trim()
    })
    newEquipeLibelle.value = ''
    flash.value = t('applications.equipe_created')
    flashError.value = false
    await loadAll()
    const row = rows.value.find((r) => r.id === editingId.value)
    if (row) {
      form.equipeIds = [...row.equipeIds]
      form.siteIds = [...row.siteIds]
      form.serviceIds = [...row.serviceIds]
    }
  } catch (err) {
    formError.value = extractFetchError(err)
  }
}

const goCreateBudget = () => {
  if (!editingId.value) return
  navigateTo(`/budget?create=1&applicationId=${editingId.value}`)
}

const deactivateRow = async (row: AppRow) => {
  if (!confirm(t('applications.deactivate_confirm', { name: row.libelle }))) return
  actionBusyId.value = row.id
  try {
    await deactivate(row.id)
    flash.value = t('applications.deactivated')
    flashError.value = false
    await loadAll()
  } catch (err) {
    flash.value = extractFetchError(err)
    flashError.value = true
  } finally {
    actionBusyId.value = ''
  }
}

const activateRow = async (row: AppRow) => {
  actionBusyId.value = row.id
  try {
    await activate(row.id)
    flash.value = t('applications.activated')
    flashError.value = false
    await loadAll()
  } catch (err) {
    flash.value = extractFetchError(err)
    flashError.value = true
  } finally {
    actionBusyId.value = ''
  }
}

onMounted(async () => {
  await loadAll()
  const id = typeof route.query.id === 'string' ? route.query.id : ''
  if (id) {
    const row = rows.value.find((a) => a.id === id)
    if (row) openEdit(row)
  }
})
</script>

<style scoped>
.apps-flash {
  margin: 0 0 var(--kore-space-md);
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
  background: color-mix(in srgb, var(--kore-success) 18%, transparent);
  color: var(--kore-text);
}
.apps-flash--error {
  background: color-mix(in srgb, var(--kore-error) 18%, transparent);
}
.apps-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-xs);
}
.apps-libelle {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-xs);
}
.apps-select {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.apps-merge-mobile {
  display: none;
  width: 100%;
  margin-bottom: var(--kore-space-md);
}
.apps-merge-warning {
  margin: 0;
  padding: var(--kore-space-sm) var(--kore-space-md);
  border-radius: var(--kore-radius-md);
  background: color-mix(in srgb, var(--kore-warning) 18%, transparent);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}
.apps-merge-badge {
  margin-left: var(--kore-space-xs);
}
.apps-form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
}
.apps-form__title {
  margin: 0;
  font-size: var(--kore-text-h3);
}
.apps-form__body {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}
.apps-form__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}
.apps-form__field > span:first-child {
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}
.apps-form__input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-elevated);
  color: var(--kore-text);
  font-family: var(--kore-font);
  font-size: var(--kore-text-body);
}
.apps-form__toggle {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}
.apps-form__section {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-md);
  padding-top: var(--kore-space-md);
  border-top: 1px solid var(--kore-border);
}
.apps-form__section:first-child {
  padding-top: 0;
  border-top: none;
}
.apps-form__section--invalid .apps-form__checklist {
  border-color: var(--kore-error);
}
.apps-form__section h3 {
  margin: 0;
  font-size: var(--kore-text-body);
  font-weight: 600;
}
.apps-form__hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}
.apps-form__shares-summary {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
  font-weight: 500;
}
.apps-form__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: var(--kore-space-xs);
}
.apps-form__list-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--kore-space-sm);
}
.apps-form__inline {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: end;
  gap: var(--kore-space-sm);
}
.apps-form__checkgroup {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
  margin: 0;
  padding: 0;
  border: none;
}
.apps-form__checkgroup legend {
  margin: 0 0 var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
  font-weight: 500;
}
.apps-form__checklist {
  display: grid;
  gap: var(--kore-space-xs);
  max-height: 10rem;
  overflow: auto;
  padding: var(--kore-space-sm);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg-subtle);
}
.apps-form__checklist--scroll .apps-form__checklist {
  max-height: 14rem;
}
.apps-form__radiogroup {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-md);
  margin: 0;
  padding: 0;
  border: none;
}
.apps-form__radio {
  display: inline-flex;
  align-items: center;
  gap: var(--kore-space-xs);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}
.apps-form__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.apps-form__check {
  display: flex;
  align-items: flex-start;
  gap: var(--kore-space-sm);
  font-size: var(--kore-text-small);
  color: var(--kore-text);
  cursor: pointer;
}
.apps-form__link {
  color: var(--kore-link);
  text-decoration: underline;
  font-size: var(--kore-text-small);
  width: fit-content;
}
.apps-form__error {
  color: var(--kore-error);
  margin: 0;
  font-size: var(--kore-text-small);
}
.apps-form__footer {
  position: sticky;
  bottom: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-sm);
  padding-top: var(--kore-space-md);
  margin-top: var(--kore-space-xs);
  border-top: 1px solid var(--kore-border);
  background: var(--kore-bg-elevated);
  box-shadow: 0 -6px 12px color-mix(in srgb, var(--kore-bg-elevated) 85%, transparent);
}
.apps-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--kore-space-sm);
  justify-content: flex-end;
}
.muted {
  color: var(--kore-text-muted);
  margin: 0;
}
@media (max-width: 768px) {
  .apps-merge-mobile {
    display: flex;
  }
  .apps-form__inline {
    grid-template-columns: 1fr;
  }
  .apps-form__actions {
    flex-direction: column;
  }
  .apps-form__actions :deep(.app-btn) {
    width: 100%;
  }
}
</style>
