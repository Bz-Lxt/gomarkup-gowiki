import type { Op } from './types'

export const START = { site: 0, clock: 0 }

export interface Atom {
  id: { site: number; clock: number }
  after: { site: number; clock: number }
  value: string
  deleted: boolean
}

function eq(a: { site: number; clock: number }, b: { site: number; clock: number }) {
  return a.site === b.site && a.clock === b.clock
}
function greater(a: { site: number; clock: number }, b: { site: number; clock: number }) {
  if (a.clock !== b.clock) return a.clock > b.clock
  return a.site > b.site
}
function key(id: { site: number; clock: number }) {
  return `${id.site}:${id.clock}`
}

export class LocalRGA {
  site = 1
  clock = 0
  atoms = new Map<string, Atom>()
  constructor() {
    this.atoms.set(key(START), { id: START, after: START, value: '', deleted: false })
  }
  load(site: number, clock: number, atoms: Atom[]) {
    this.site = site
    this.clock = clock
    this.atoms = new Map([[key(START), { id: START, after: START, value: '', deleted: false }]])
    for (const a of atoms || []) this.atoms.set(key(a.id), { ...a })
  }
  apply(op: Op) {
    if (op.type === 'insert' && op.value && op.after) {
      this.atoms.set(key(op.id), { id: op.id, after: op.after, value: op.value, deleted: false })
      if (op.id.clock > this.clock) this.clock = op.id.clock
    }
    if (op.type === 'delete' && op.target) {
      const a = this.atoms.get(key(op.target))
      if (a) a.deleted = true
      if (op.id.clock > this.clock) this.clock = op.id.clock
    }
  }
  visible(): Atom[] {
    const out: Atom[] = []
    const walk = (parent: { site: number; clock: number }) => {
      const kids = [...this.atoms.values()].filter((a) => eq(a.after, parent) && !eq(a.id, parent))
      kids.sort((a, b) => (greater(a.id, b.id) ? -1 : 1))
      for (const k of kids) {
        if (!k.deleted) out.push(k)
        walk(k.id)
      }
    }
    walk(START)
    return out
  }
  text() {
    return this.visible().map((a) => a.value).join('')
  }
  localInsert(index: number, s: string): Op[] {
    const vis = this.visible()
    let after = START
    if (index > 0) after = vis[index - 1].id
    const ops: Op[] = []
    for (const ch of Array.from(s)) {
      this.clock += 1
      const id = { site: this.site, clock: this.clock }
      const op: Op = { type: 'insert', id, after, value: ch }
      this.apply(op)
      ops.push(op)
      after = id
    }
    return ops
  }
  localDelete(index: number, length: number): Op[] {
    const vis = this.visible()
    const ops: Op[] = []
    for (let i = 0; i < length; i++) {
      this.clock += 1
      const op: Op = { type: 'delete', id: { site: this.site, clock: this.clock }, target: vis[index + i].id }
      this.apply(op)
      ops.push(op)
    }
    return ops
  }
}
