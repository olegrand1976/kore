export type RequestChannel = 'tma' | 'support' | 'maintenance'

export type ChannelsEnabled = {
  tma: boolean
  support: boolean
  maintenance: boolean
}

export type TenantRequestSettings = {
  tenantId?: string
  channelsEnabled: ChannelsEnabled
  guidesEnabled: boolean
  updatedAt?: string
}

const defaultChannels: ChannelsEnabled = {
  tma: true,
  support: true,
  maintenance: true
}

export function useRequestSettings() {
  const settings = useState<TenantRequestSettings | null>('request-settings', () => null)
  const loaded = useState('request-settings-loaded', () => false)

  const fetchSettings = async () => {
    const { apiFetch } = useApiFetch()
    try {
      const res = await apiFetch<{ data?: TenantRequestSettings }>('/api/request-settings')
      const data = res.data
      settings.value = {
        channelsEnabled: {
          tma: data?.channelsEnabled?.tma ?? defaultChannels.tma,
          support: data?.channelsEnabled?.support ?? defaultChannels.support,
          maintenance: data?.channelsEnabled?.maintenance ?? defaultChannels.maintenance
        },
        guidesEnabled: data?.guidesEnabled ?? true,
        updatedAt: data?.updatedAt
      }
      loaded.value = true
    } catch {
      settings.value = { channelsEnabled: { ...defaultChannels }, guidesEnabled: true }
      loaded.value = true
    }
  }

  const saveSettings = async (payload: Pick<TenantRequestSettings, 'channelsEnabled' | 'guidesEnabled'>) => {
    const { apiFetch } = useApiFetch()
    const res = await apiFetch<{ data?: TenantRequestSettings }>('/api/admin/request-settings', {
      method: 'PUT',
      body: payload
    })
    const data = res.data
    if (data) {
      settings.value = {
        channelsEnabled: data.channelsEnabled,
        guidesEnabled: data.guidesEnabled,
        updatedAt: data.updatedAt
      }
    }
    return settings.value
  }

  const isChannelEnabled = (channel: RequestChannel) => {
    const ch = settings.value?.channelsEnabled
    if (!ch) return true
    return ch[channel] ?? false
  }

  const activeChannelCount = computed(() => {
    const ch = settings.value?.channelsEnabled
    if (!ch) return 3
    return Number(ch.tma) + Number(ch.support) + Number(ch.maintenance)
  })

  const guidesEnabled = computed(() => settings.value?.guidesEnabled ?? true)

  return {
    settings,
    loaded,
    fetchSettings,
    saveSettings,
    isChannelEnabled,
    activeChannelCount,
    guidesEnabled
  }
}
