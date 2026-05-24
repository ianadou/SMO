<script setup lang="ts">
import { CheckCircle2, AlertTriangle, Info, X, XCircle } from 'lucide-vue-next'
import { useToast } from '~/composables/useToast'

const { toasts, dismiss } = useToast()

const iconFor = {
  success: CheckCircle2,
  error: XCircle,
  warning: AlertTriangle,
  info: Info,
}
</script>

<template>
  <section class="toast-stack" aria-live="polite" aria-label="Notifications">
    <TransitionGroup name="toast" tag="div" class="toast-stack-inner">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast', `is-${toast.kind}`]"
        role="alert"
      >
        <component :is="iconFor[toast.kind]" :size="18" class="toast-icon" />
        <div class="toast-body">
          <div class="toast-title">{{ toast.title }}</div>
          <div v-if="toast.message" class="toast-message">{{ toast.message }}</div>
        </div>
        <button class="toast-close" aria-label="Fermer" @click="dismiss(toast.id)">
          <X :size="16" />
        </button>
      </div>
    </TransitionGroup>
  </section>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  inset: auto var(--space-3) var(--space-3) var(--space-3);
  display: flex;
  justify-content: center;
  pointer-events: none;
  z-index: 1000;
}
.toast-stack-inner {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  width: 100%;
  max-width: 420px;
}

.toast {
  pointer-events: auto;
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-elevated);
  color: var(--color-fg-default);
  font-size: 14px;
  line-height: 1.4;
}

.toast-icon { flex-shrink: 0; margin-top: 2px; }
.toast.is-success .toast-icon { color: var(--color-success); }
.toast.is-error .toast-icon { color: var(--color-danger); }
.toast.is-warning .toast-icon { color: var(--color-warn); }
.toast.is-info .toast-icon { color: var(--color-action-primary); }

.toast-body { flex: 1; min-width: 0; }
.toast-title { font-weight: 600; }
.toast-message {
  margin-top: 2px;
  color: var(--color-fg-muted);
  font-size: 13px;
}

.toast-close {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: 0;
  border-radius: var(--radius-pill);
  color: var(--color-fg-muted);
  cursor: pointer;
  transition: background var(--motion-default) var(--motion-easing);
}
.toast-close:hover { background: rgba(255, 255, 255, 0.06); }

.toast-enter-from { opacity: 0; transform: translateY(8px); }
.toast-enter-to { opacity: 1; transform: translateY(0); }
.toast-enter-active,
.toast-leave-active { transition: opacity var(--motion-default) var(--motion-easing), transform var(--motion-default) var(--motion-easing); }
.toast-leave-from { opacity: 1; }
.toast-leave-to { opacity: 0; transform: translateY(8px); }
</style>
