<script setup lang="ts">
import { Eye, EyeOff } from 'lucide-vue-next'
import TextField from '~/components/login/TextField.vue'
import PrimaryButton from '~/components/login/PrimaryButton.vue'
import InlineError from '~/components/login/InlineError.vue'

definePageMeta({ layout: false })

useHead({ title: 'Connexion organisateur — SMO' })

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const loading = ref(false)
const error = ref('')

const STUB_RESPONSE_DELAY_MS = 1400

async function submit() {
  if (!email.value || !password.value) return
  error.value = ''
  loading.value = true
  await new Promise((resolve) => setTimeout(resolve, STUB_RESPONSE_DELAY_MS))
  loading.value = false
  error.value = 'Identifiants incorrects.'
}
</script>

<template>
  <main class="min-h-dvh flex flex-col">
    <div class="flex-1 flex items-center justify-center py-8">
      <form
        class="flex flex-col w-full max-w-[380px] mx-auto px-6"
        novalidate
        @submit.prevent="submit"
      >
        <div class="mb-10">
          <Wordmark :size="40" />
        </div>

        <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] text-fg-default mb-6">
          Connexion organisateur
        </h1>

        <div class="flex flex-col gap-4 mb-5">
          <TextField
            id="email"
            v-model="email"
            label="Email"
            type="email"
            inputmode="email"
            autocomplete="email"
            placeholder="toi@exemple.fr"
            :has-error="!!error"
          />

          <TextField
            id="password"
            v-model="password"
            label="Mot de passe"
            :type="showPassword ? 'text' : 'password'"
            :autocomplete="showPassword ? 'off' : 'current-password'"
            placeholder="••••••••"
            :has-error="!!error"
          >
            <template #right>
              <button
                type="button"
                class="inline-flex items-center justify-center w-10 h-10 mr-1 bg-transparent border-0 text-fg-muted rounded-sm cursor-pointer transition-colors hover:bg-white/[0.06] hover:text-fg-default active:bg-white/[0.10] focus-visible:outline-none focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-elevated),0_0_0_4px_rgba(32,128,255,0.45)]"
                :aria-label="showPassword ? 'Masquer le mot de passe' : 'Afficher le mot de passe'"
                @click="showPassword = !showPassword"
              >
                <EyeOff v-if="showPassword" :size="18" />
                <Eye v-else :size="18" />
              </button>
            </template>
          </TextField>
        </div>

        <PrimaryButton :loading="loading">
          Se connecter
        </PrimaryButton>

        <InlineError v-if="error">{{ error }}</InlineError>

        <div class="mt-5 text-[15px] text-fg-muted text-center">
          Pas encore de compte ?
          <NuxtLink
            to="/register"
            class="text-fg-default underline underline-offset-[3px] decoration-fg-muted transition-colors hover:decoration-fg-default focus-visible:outline-none focus-visible:rounded-sm focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
          >
            S'inscrire
          </NuxtLink>
        </div>

        <div class="mt-10 text-[13px] leading-[1.4] text-fg-muted text-center">
          Les joueurs n'ont pas besoin de compte — ils accèdent par lien d'invitation.
        </div>
      </form>
    </div>
  </main>
</template>
