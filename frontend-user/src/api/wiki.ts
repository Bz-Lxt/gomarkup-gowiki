import { http, type Api } from './http'

export interface User {
  id: string
  email: string
  displayName: string
  avatarColor: string
}

export interface AuthPayload {
  accessToken: string
  refreshToken: string
  expiresAt: string
  user: User
}

export interface Space {
  id: string
  name: string
  ownerId: string
}

export interface Document {
  id: string
  spaceId: string
  parentId: string | null
  title: string
  path: string
  sortOrder: number
  contentMd: string
  contentJson: string
  editorMode: 'markdown' | 'rich'
  createdAt: string
  updatedAt: string
}

export interface Version {
  id: string
  documentId: string
  layer: string
  label: string
  contentMd: string
  authorId: string
  createdAt: string
}

export const api = {
  login: (email: string, password: string) =>
    http.post('/api/v1/auth/login', { email, password }) as Api<AuthPayload>,
  register: (email: string, password: string, name: string) =>
    http.post('/api/v1/auth/register', { email, password, name }) as Api<AuthPayload>,
  me: () => http.get('/api/v1/auth/me') as Api<{ id: string; name: string; email: string }>,
  spaces: () => http.get('/api/v1/spaces') as Api<Space[]>,
  createSpace: (name: string) => http.post('/api/v1/spaces', { name }) as Api<Space>,
  tree: (spaceId: string) => http.get('/api/v1/documents', { params: { spaceId } }) as Api<Document[]>,
  createDoc: (payload: { spaceId: string; parentId?: string | null; title: string; editorMode?: string }) =>
    http.post('/api/v1/documents', payload) as Api<Document>,
  getDoc: (id: string) => http.get(`/api/v1/documents/${id}`) as Api<{ document: Document; favorite: boolean }>,
  updateDoc: (id: string, payload: Partial<Document>) => http.patch(`/api/v1/documents/${id}`, payload) as Api<Document>,
  deleteDoc: (id: string) => http.delete(`/api/v1/documents/${id}`) as Api<{ ok: boolean }>,
  moveDoc: (id: string, parentId: string | null, sortOrder: number) =>
    http.post(`/api/v1/documents/${id}/move`, { parentId, sortOrder }) as Api<Document>,
  recycle: () => http.get('/api/v1/documents/recycle') as Api<Document[]>,
  restore: (id: string) => http.post(`/api/v1/documents/${id}/restore`) as Api<Document>,
  favorite: (id: string) => http.post(`/api/v1/documents/${id}/favorite`) as Api<{ favorite: boolean }>,
  versions: (id: string) => http.get(`/api/v1/documents/${id}/versions`) as Api<Version[]>,
  saveVersion: (id: string, label: string) => http.post(`/api/v1/documents/${id}/versions`, { label }) as Api<Version>,
  diff: (left: string, right = 'current') => http.get('/api/v1/versions/diff', { params: { left, right } }) as Api<any>,
  rollback: (id: string) => http.post(`/api/v1/versions/${id}/rollback`) as Api<Version>,
  search: (q: string) => http.get('/api/v1/search', { params: { q } }) as Api<{ hits: any[]; analyzer: string }>,
  workbench: () => http.get('/api/v1/workbench') as Api<any>,
  upload: (file: File) => {
    const fd = new FormData()
    fd.append('file', file)
    return http.post('/api/v1/uploads', fd) as Api<{ url: string }>
  },
}
