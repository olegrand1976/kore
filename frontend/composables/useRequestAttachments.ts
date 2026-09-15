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

export type PreviewKind = 'image' | 'pdf' | 'unsupported' | 'none'

const IMAGE_EXT = new Set(['.png', '.jpg', '.jpeg', '.gif', '.webp'])
const PDF_EXT = new Set(['.pdf'])

export function pickMimeType(att: RequestAttachment): string {
  return (att.mimeType ?? att.MimeType ?? '').toLowerCase()
}

export function pickFileExtension(fileName: string): string {
  const i = fileName.lastIndexOf('.')
  return i >= 0 ? fileName.slice(i).toLowerCase() : ''
}

export function previewKindFrom(mime: string, fileName: string): PreviewKind {
  const m = mime.toLowerCase()
  if (m.startsWith('image/')) return 'image'
  if (m === 'application/pdf' || m === 'application/x-pdf') return 'pdf'
  const ext = pickFileExtension(fileName)
  if (IMAGE_EXT.has(ext)) return 'image'
  if (PDF_EXT.has(ext)) return 'pdf'
  if (!fileName && !mime) return 'none'
  return 'unsupported'
}

export function isPreviewable(mime: string, fileName: string): boolean {
  const kind = previewKindFrom(mime, fileName)
  return kind === 'image' || kind === 'pdf'
}

export function isIndexableAttachment(fileName: string): boolean {
  const ext = pickFileExtension(fileName)
  return ['.txt', '.md', '.csv', '.log', '.pdf'].includes(ext)
}

export function useRequestAttachments() {
  const { apiFetch } = useApiFetch()
  const pickId = (att: RequestAttachment) => att.id ?? att.ID ?? ''
  const pickFileName = (att: RequestAttachment) => att.fileName ?? att.FileName ?? ''

  const list = async (resourceType: string, resourceId: string) => {
    const res = await apiFetch<{ data?: RequestAttachment[] }>('/api/request-attachments', {
      query: { resourceType, resourceId }
    })
    return res?.data ?? []
  }

  const upload = async (resourceType: string, resourceId: string, file: File) => {
    const form = new FormData()
    form.append('resourceType', resourceType)
    form.append('resourceId', resourceId)
    form.append('file', file)
    const res = await apiFetch<{ data?: RequestAttachment }>('/api/request-attachments', {
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

  const fetchContent = async (id: string) => {
    return apiFetch<Blob>(downloadUrl(id), { responseType: 'blob' })
  }

  const remove = async (id: string) => {
    await apiFetch(`/api/request-attachments/${id}`, { method: 'DELETE' })
  }

  return {
    list,
    upload,
    uploadAll,
    downloadUrl,
    fetchContent,
    remove,
    pickId,
    pickFileName,
    pickMimeType
  }
}
