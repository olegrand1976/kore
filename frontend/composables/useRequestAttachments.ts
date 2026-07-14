export type RequestAttachment = {
  id?: string
  ID?: string
  fileName?: string
  FileName?: string
  mimeType?: string
  MimeType?: string
  sizeBytes?: number
  SizeBytes?: number
}

export const REQUEST_RESOURCE = {
  tma: 'tma_demand',
  support: 'support_ticket',
  maintenance: 'maintenance_work_request'
} as const

export type RequestResourceKey = keyof typeof REQUEST_RESOURCE

export function useRequestAttachments() {
  const pickId = (att: RequestAttachment) => att.id ?? att.ID ?? ''
  const pickFileName = (att: RequestAttachment) => att.fileName ?? att.FileName ?? ''

  const list = async (resourceType: string, resourceId: string) => {
    const res = await $fetch<{ data?: RequestAttachment[] }>('/api/request-attachments', {
      query: { resourceType, resourceId }
    })
    return res?.data ?? []
  }

  const upload = async (resourceType: string, resourceId: string, file: File) => {
    const form = new FormData()
    form.append('resourceType', resourceType)
    form.append('resourceId', resourceId)
    form.append('file', file)
    const res = await $fetch<{ data?: RequestAttachment }>('/api/request-attachments', {
      method: 'POST',
      body: form
    })
    return (res?.data ?? res) as RequestAttachment
  }

  const uploadAll = async (resourceType: string, resourceId: string, files: File[]) => {
    for (const file of files) {
      await upload(resourceType, resourceId, file)
    }
  }

  const downloadUrl = (id: string) => `/api/request-attachments/${id}/download`

  return { list, upload, uploadAll, downloadUrl, pickId, pickFileName }
}
