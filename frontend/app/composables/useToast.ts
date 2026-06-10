import { ref } from 'vue'

export type ToastKind = 'success' | 'error' | 'info' | 'warning'

export interface Toast {
  id: number
  kind: ToastKind
  title: string
  message?: string
  duration: number
}

const toasts = ref<Toast[]>([])
let seq = 0

function show(toast: Omit<Toast, 'id'>) {
  const id = ++seq
  toasts.value = [...toasts.value, { id, ...toast }]
  if (toast.duration > 0) {
    setTimeout(() => dismiss(id), toast.duration)
  }
  return id
}

function dismiss(id: number) {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

export function useToast() {
  return {
    toasts,
    dismiss,
    success: (title: string, message?: string, duration = 4000) =>
      show({ kind: 'success', title, message, duration }),
    error: (title: string, message?: string, duration = 6000) =>
      show({ kind: 'error', title, message, duration }),
    info: (title: string, message?: string, duration = 4000) =>
      show({ kind: 'info', title, message, duration }),
    warning: (title: string, message?: string, duration = 5000) =>
      show({ kind: 'warning', title, message, duration }),
  }
}
