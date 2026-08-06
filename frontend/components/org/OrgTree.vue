<script setup lang="ts">
// Arbre d'administration de la hiérarchie organisation.
// Chaque niveau exige son parent : un service exige un site ET un responsable,
// une application exige un service, une équipe exige une application.
import type {
  OrgSociete,
  OrgSite,
  OrgService,
  OrgApplication,
  OrgEquipe
} from '~/composables/useOrganisation'
import type { OrgUserSummary } from '~/composables/useUsers'

type Level = 'site' | 'service' | 'application' | 'equipe'

const { t } = useI18n()
const {
  listSocietes,
  listSites,
  listServices,
  listApplications,
  listEquipes,
  createSite,
  createService,
  createApplication,
  updateApplication,
  deactivateApplication,
  activateApplication,
  createEquipe,
  orgId,
  orgLabel
} = useOrganisation()
const { list: listUsers, pickUserId, pickUserLogin, pickUserEquipeIds } = useUsers()
const { extractFetchError } = useApiError()

const societes = ref<OrgSociete[]>([])
const sites = ref<OrgSite[]>([])
const services = ref<OrgService[]>([])
const applications = ref<OrgApplication[]>([])
const equipes = ref<OrgEquipe[]>([])
const users = ref<OrgUserSummary[]>([])

const pending = ref(true)
const loadError = ref('')

const modalOpen = ref(false)
const modalLevel = ref<Level>('site')
const parentId = ref('')
const parentLabel = ref('')
const submitting = ref(false)
const formError = ref('')
const form = reactive({ libelle: '', type: 'interne', responsableId: '' })

const editOpen = ref(false)
const editSubmitting = ref(false)
const editError = ref('')
const editForm = reactive({ id: '', libelle: '' })
const actionBusyId = ref('')

const isAppActive = (app: OrgApplication) => app.active ?? app.Active ?? true

const load = async () => {
  pending.value = true
  loadError.value = ''
  try {
    const [soc, sit, srv, app, eq, usr] = await Promise.all([
      listSocietes(),
      listSites(),
      listServices(),
      listApplications(),
      listEquipes(),
      listUsers()
    ])
    societes.value = soc
    sites.value = sit
    services.value = srv
    applications.value = app
    equipes.value = eq
    users.value = usr
  } catch (err) {
    loadError.value = extractFetchError(err, t('org.tree.load_error'))
  } finally {
    pending.value = false
  }
}

onMounted(load)

const sitesOf = (societeId: string) =>
  sites.value.filter((s) => (s.societeId ?? s.SocieteID) === societeId)

const servicesOf = (siteId: string) =>
  services.value.filter((s) => (s.siteId ?? s.SiteID) === siteId)

const applicationsOf = (serviceId: string) =>
  applications.value.filter((a) => (a.serviceId ?? a.ServiceID) === serviceId)

const equipesOf = (applicationId: string) =>
  equipes.value.filter((e) => (e.applicationId ?? e.ApplicationID) === applicationId)

const memberCount = (equipeId: string) =>
  users.value.filter((u) => pickUserEquipeIds(u).includes(equipeId)).length

// Un service n'a pas toujours de libellé (colonne ajoutée en 0018) : on retombe
// sur son type pour que le nœud reste identifiable.
const serviceLabel = (service: OrgService) =>
  orgLabel(service) || service.type || t('org.tree.service_unnamed')

const societeLabel = (societe: OrgSociete) =>
  societe.raisonSociale ?? societe.RaisonSociale ?? t('org.tree.societe_unnamed')

const openModal = (level: Level, parent: { id: string; label: string }) => {
  modalLevel.value = level
  parentId.value = parent.id
  parentLabel.value = parent.label
  form.libelle = ''
  form.type = 'interne'
  form.responsableId = ''
  formError.value = ''
  modalOpen.value = true
}

const modalTitle = computed(() => t(`org.tree.add_${modalLevel.value}`))

const submit = async () => {
  submitting.value = true
  formError.value = ''
  try {
    switch (modalLevel.value) {
      case 'site':
        await createSite({ societeId: parentId.value, libelle: form.libelle })
        break
      case 'service':
        await createService({
          siteId: parentId.value,
          libelle: form.libelle,
          type: form.type,
          responsableId: form.responsableId
        })
        break
      case 'application':
        await createApplication({ serviceId: parentId.value, libelle: form.libelle })
        break
      case 'equipe':
        await createEquipe({
          applicationId: parentId.value,
          libelle: form.libelle,
          responsableId: form.responsableId || undefined
        })
        break
      default: {
        const _exhaustive: never = modalLevel.value
        throw new Error(`Unhandled level: ${_exhaustive}`)
      }
    }
    modalOpen.value = false
    await load()
  } catch (err) {
    formError.value = extractFetchError(err, t('org.tree.create_error'))
  } finally {
    submitting.value = false
  }
}

const openEditApplication = (app: OrgApplication) => {
  editForm.id = orgId(app)
  editForm.libelle = orgLabel(app)
  editError.value = ''
  editOpen.value = true
}

const submitEditApplication = async () => {
  editSubmitting.value = true
  editError.value = ''
  try {
    await updateApplication(editForm.id, { libelle: editForm.libelle })
    editOpen.value = false
    await load()
  } catch (err) {
    editError.value = extractFetchError(err, t('org.tree.update_error'))
  } finally {
    editSubmitting.value = false
  }
}

const toggleApplicationActive = async (app: OrgApplication) => {
  const id = orgId(app)
  if (!id || actionBusyId.value) return
  const active = isAppActive(app)
  if (active && !window.confirm(t('org.tree.deactivate_confirm'))) return
  actionBusyId.value = id
  try {
    if (active) {
      await deactivateApplication(id)
    } else {
      await activateApplication(id)
    }
    await load()
  } catch (err) {
    loadError.value = extractFetchError(
      err,
      active ? t('org.tree.deactivate_error') : t('org.tree.activate_error')
    )
  } finally {
    actionBusyId.value = ''
  }
}

defineExpose({ reload: load })
</script>

<template>
  <div class="org-tree">
    <AppCard v-if="pending" padding="lg">
      <p class="muted">{{ $t('org.tree.loading') }}</p>
    </AppCard>

    <AppCard v-else-if="loadError" padding="lg">
      <AppEmptyState icon="error" :title="loadError" />
    </AppCard>

    <AppCard v-else-if="societes.length === 0" padding="lg">
      <AppEmptyState icon="domain_disabled" :title="$t('org.tree.no_societe')" />
    </AppCard>

    <AppCard v-else padding="lg">
      <p class="org-tree__intro">{{ $t('org.tree.intro') }}</p>

      <ul class="org-tree__list">
        <li v-for="societe in societes" :key="orgId(societe)" class="org-tree__node">
          <div class="org-tree__row">
            <AppIcon name="domain" />
            <span class="org-tree__label">{{ societeLabel(societe) }}</span>
            <AppBadge variant="default">{{ $t('org.tree.level_societe') }}</AppBadge>
            <AppButton
              variant="ghost"
              size="sm"
              @click="openModal('site', { id: orgId(societe), label: societeLabel(societe) })"
            >
              <AppIcon name="add" /> {{ $t('org.tree.add_site') }}
            </AppButton>
          </div>

          <p v-if="sitesOf(orgId(societe)).length === 0" class="org-tree__hint">
            {{ $t('org.tree.empty_sites') }}
          </p>

          <ul class="org-tree__list">
            <li v-for="site in sitesOf(orgId(societe))" :key="orgId(site)" class="org-tree__node">
              <div class="org-tree__row">
                <AppIcon name="location_on" />
                <span class="org-tree__label">{{ orgLabel(site) }}</span>
                <AppBadge variant="default">{{ $t('org.tree.level_site') }}</AppBadge>
                <AppButton
                  variant="ghost"
                  size="sm"
                  @click="openModal('service', { id: orgId(site), label: orgLabel(site) })"
                >
                  <AppIcon name="add" /> {{ $t('org.tree.add_service') }}
                </AppButton>
              </div>

              <p v-if="servicesOf(orgId(site)).length === 0" class="org-tree__hint">
                {{ $t('org.tree.empty_services') }}
              </p>

              <ul class="org-tree__list">
                <li
                  v-for="service in servicesOf(orgId(site))"
                  :key="orgId(service)"
                  class="org-tree__node"
                >
                  <div class="org-tree__row">
                    <AppIcon name="account_tree" />
                    <span class="org-tree__label">{{ serviceLabel(service) }}</span>
                    <AppBadge variant="default">{{ $t('org.tree.level_service') }}</AppBadge>
                    <AppButton
                      variant="ghost"
                      size="sm"
                      @click="
                        openModal('application', {
                          id: orgId(service),
                          label: serviceLabel(service)
                        })
                      "
                    >
                      <AppIcon name="add" /> {{ $t('org.tree.add_application') }}
                    </AppButton>
                  </div>

                  <p v-if="applicationsOf(orgId(service)).length === 0" class="org-tree__hint">
                    {{ $t('org.tree.empty_applications') }}
                  </p>

                  <ul class="org-tree__list">
                    <li
                      v-for="application in applicationsOf(orgId(service))"
                      :key="orgId(application)"
                      class="org-tree__node"
                    >
                      <div class="org-tree__row">
                        <AppIcon name="apps" />
                        <span class="org-tree__label" :class="{ 'org-tree__label--inactive': !isAppActive(application) }">
                          {{ orgLabel(application) }}
                        </span>
                        <AppBadge variant="default">{{ $t('org.tree.level_application') }}</AppBadge>
                        <AppBadge v-if="!isAppActive(application)" variant="warning">
                          {{ $t('org.tree.inactive_badge') }}
                        </AppBadge>
                        <AppButton
                          variant="ghost"
                          size="sm"
                          type="button"
                          @click="openEditApplication(application)"
                        >
                          {{ $t('org.tree.edit_application') }}
                        </AppButton>
                        <AppButton
                          variant="ghost"
                          size="sm"
                          type="button"
                          :disabled="actionBusyId === orgId(application)"
                          @click="toggleApplicationActive(application)"
                        >
                          {{
                            isAppActive(application)
                              ? $t('org.tree.deactivate_application')
                              : $t('org.tree.activate_application')
                          }}
                        </AppButton>
                        <AppButton
                          v-if="isAppActive(application)"
                          variant="ghost"
                          size="sm"
                          @click="
                            openModal('equipe', {
                              id: orgId(application),
                              label: orgLabel(application)
                            })
                          "
                        >
                          <AppIcon name="add" /> {{ $t('org.tree.add_equipe') }}
                        </AppButton>
                      </div>

                      <p v-if="equipesOf(orgId(application)).length === 0" class="org-tree__hint">
                        {{ $t('org.tree.empty_equipes') }}
                      </p>

                      <ul class="org-tree__list">
                        <li
                          v-for="equipe in equipesOf(orgId(application))"
                          :key="orgId(equipe)"
                          class="org-tree__node org-tree__node--leaf"
                        >
                          <div class="org-tree__row">
                            <AppIcon name="groups" />
                            <span class="org-tree__label">{{ orgLabel(equipe) }}</span>
                            <AppBadge :variant="memberCount(orgId(equipe)) > 0 ? 'success' : 'default'">
                              {{ $t('org.tree.members', { count: memberCount(orgId(equipe)) }) }}
                            </AppBadge>
                          </div>
                        </li>
                      </ul>
                    </li>
                  </ul>
                </li>
              </ul>
            </li>
          </ul>
        </li>
      </ul>
    </AppCard>

    <AppModal v-model:open="modalOpen" width="sm" :aria-label="modalTitle">
      <form class="org-tree__form" @submit.prevent="submit">
        <h2 class="org-tree__form-title">{{ modalTitle }}</h2>
        <p class="org-tree__form-parent">
          {{ $t('org.tree.parent', { parent: parentLabel }) }}
        </p>

        <AppInput
          id="org-node-libelle"
          v-model="form.libelle"
          :label="$t('org.tree.field_libelle')"
          required
        />

        <div v-if="modalLevel === 'service'" class="org-tree__field">
          <label for="org-service-type">{{ $t('org.tree.field_type') }}</label>
          <select id="org-service-type" v-model="form.type">
            <option value="interne">{{ $t('org.tree.type_interne') }}</option>
            <option value="externe">{{ $t('org.tree.type_externe') }}</option>
          </select>
        </div>

        <div v-if="modalLevel === 'service' || modalLevel === 'equipe'" class="org-tree__field">
          <label for="org-node-responsable">
            {{
              modalLevel === 'service'
                ? $t('org.tree.field_responsable_required')
                : $t('org.tree.field_responsable_optional')
            }}
          </label>
          <select
            id="org-node-responsable"
            v-model="form.responsableId"
            :required="modalLevel === 'service'"
          >
            <option value="">{{ $t('org.tree.responsable_none') }}</option>
            <option v-for="user in users" :key="pickUserId(user)" :value="pickUserId(user)">
              {{ pickUserLogin(user) }}
            </option>
          </select>
          <p v-if="modalLevel === 'service'" class="org-tree__field-hint">
            {{ $t('org.tree.responsable_hint') }}
          </p>
        </div>

        <p v-if="formError" class="org-tree__form-error" role="alert">{{ formError }}</p>

        <div class="org-tree__form-actions">
          <AppButton variant="ghost" type="button" @click="modalOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="submitting">
            {{ submitting ? $t('org.tree.creating') : $t('org.tree.create') }}
          </AppButton>
        </div>
      </form>
    </AppModal>

    <AppModal v-model:open="editOpen" width="sm" :aria-label="$t('org.tree.edit_application_title')">
      <form class="org-tree__form" @submit.prevent="submitEditApplication">
        <h2 class="org-tree__form-title">{{ $t('org.tree.edit_application_title') }}</h2>
        <AppInput
          id="org-edit-app-libelle"
          v-model="editForm.libelle"
          :label="$t('org.tree.field_libelle')"
          required
        />
        <p v-if="editError" class="org-tree__form-error" role="alert">{{ editError }}</p>
        <div class="org-tree__form-actions">
          <AppButton variant="ghost" type="button" @click="editOpen = false">
            {{ $t('common.cancel') }}
          </AppButton>
          <AppButton variant="primary" type="submit" :disabled="editSubmitting">
            {{ editSubmitting ? $t('org.tree.saving') : $t('org.tree.save') }}
          </AppButton>
        </div>
      </form>
    </AppModal>
  </div>
</template>

<style scoped>
.org-tree__intro {
  margin: 0 0 var(--kore-space-lg);
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.org-tree__list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.org-tree__node > .org-tree__list {
  margin-left: var(--kore-space-lg);
  padding-left: var(--kore-space-md);
  border-left: 1px solid var(--kore-border);
}

.org-tree__row {
  display: flex;
  align-items: center;
  gap: var(--kore-space-sm);
  padding: var(--kore-space-sm) 0;
  flex-wrap: wrap;
}

.org-tree__label {
  font-weight: 600;
  font-size: var(--kore-text-small);
  color: var(--kore-text);
}

.org-tree__label--inactive {
  color: var(--kore-text-muted);
  text-decoration: line-through;
}

.org-tree__hint {
  margin: 0 0 var(--kore-space-sm) var(--kore-space-xl);
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.org-tree__form {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-lg);
}

.org-tree__form-title {
  margin: 0;
  font-size: var(--kore-text-h3);
}

.org-tree__form-parent {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-text-muted);
}

.org-tree__field {
  display: flex;
  flex-direction: column;
  gap: var(--kore-space-xs);
}

.org-tree__field label {
  font-size: var(--kore-text-small);
  font-weight: 500;
}

.org-tree__field select {
  padding: var(--kore-space-sm) var(--kore-space-md);
  border: 1px solid var(--kore-border);
  border-radius: var(--kore-radius-md);
  background: var(--kore-bg);
  color: var(--kore-text);
  font-size: var(--kore-text-small);
}

.org-tree__field-hint {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.org-tree__form-error {
  margin: 0;
  font-size: var(--kore-text-small);
  color: var(--kore-error);
}

.org-tree__form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .org-tree__node > .org-tree__list {
    margin-left: var(--kore-space-sm);
    padding-left: var(--kore-space-sm);
  }

  .org-tree__form-actions {
    flex-direction: column-reverse;
  }

  .org-tree__form-actions :deep(.app-button) {
    width: 100%;
  }
}
</style>
