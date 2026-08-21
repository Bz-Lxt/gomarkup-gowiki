export interface Op {
  type: 'insert' | 'delete'
  id: { site: number; clock: number }
  after?: { site: number; clock: number }
  target?: { site: number; clock: number }
  value?: string
}
