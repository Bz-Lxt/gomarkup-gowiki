/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'fast-diff' {
  const diff: (t1: string, t2: string) => Array<[number, string]>
  export default diff
}

declare module 'lowlight' {
  export const common: any
  export function createLowlight(langs?: any): any
}
