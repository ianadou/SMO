<script setup lang="ts">
import { Eye, EyeOff } from 'lucide-vue-next'
import TextField from '~/components/login/TextField.vue'
import PrimaryButton from '~/components/login/PrimaryButton.vue'
import InlineError from '~/components/login/InlineError.vue'
import StrengthMeter from '~/components/register/StrengthMeter.vue'
import CguCheckbox from '~/components/register/CguCheckbox.vue'
import { isEmailFormat, passwordStrength } from '~/utils/password'

definePageMeta({ layout: false })

useHead({ title: 'Créer un compte — SMO' })

const STUB_RESPONSE_DELAY_MS = 1400
const NAME_MIN = 2
const NAME_MAX = 50
const PASSWORD_MIN = 8

const name = ref('')
const email = ref('')
const password = ref('')
const showPassword = ref(false)
const cgu = ref(false)
const submitted = ref(false)
const loading = ref(false)
const apiError = ref('')

const trimmedName = computed(() => name.value.trim())
const nameValid = computed(() => trimmedName.value.length >= NAME_MIN && trimmedName.value.length <= NAME_MAX)
const emailFormatValid = computed(() => isEmailFormat(email.value))
const passwordLongEnough = computed(() => password.value.length >= PASSWORD_MIN)
const formValid = computed(() => nameValid.value && emailFormatValid.value && passwordLongEnough.value && cgu.value)

const showEmailFormatError = computed(() => submitted.value && email.value.length > 0 && !emailFormatValid.value)
const showCguError = computed(() => submitted.value && !cgu.value)

const strength = computed(() => passwordStrength(password.value))

async function submit() {
  submitted.value = true
  if (!formValid.value) return
  apiError.value = ''
  loading.value = true
  await new Promise((resolve) => setTimeout(resolve, STUB_RESPONSE_DELAY_MS))
  loading.value = false
  apiError.value = 'Un compte existe déjà pour cette adresse.'
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
        <div class="mb-8">
          <Wordmark :size="32" />
        </div>

        <h1 class="text-[28px] leading-[1.2] font-semibold tracking-[-0.01em] text-fg-default mb-2">
          Créer un compte
        </h1>
        <p class="text-[15px] leading-[1.5] text-fg-muted mb-6">
          Pour organiser tes matchs et inviter tes joueurs.
        </p>

        <div class="flex flex-col gap-5 mb-5">
          <div>
            <TextField
              id="reg-name"
              v-model="name"
              label="Comment te présenter à tes joueurs ?"
              autocomplete="name"
              placeholder="Alex L."
            />
            <p class="text-[13px] leading-[1.4] text-fg-muted mt-2">
              C'est ce que verront tes joueurs dans les invitations (ex : « Alex L. t'invite… »).
            </p>
          </div>

          <div>
            <TextField
              id="reg-email"
              v-model="email"
              label="Email"
              type="email"
              autocomplete="email"
              inputmode="email"
              placeholder="alex@exemple.fr"
              :has-error="showEmailFormatError"
            />
            <p v-if="showEmailFormatError" class="text-[13px] leading-[1.4] text-team-red mt-2">
              Format d'email invalide.
            </p>
          </div>

          <div>
            <TextField
              id="reg-password"
              v-model="password"
              label="Mot de passe"
              :type="showPassword ? 'text' : 'password'"
              :autocomplete="showPassword ? 'off' : 'new-password'"
              placeholder="••••••••"
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
            <StrengthMeter :level="strength" />
          </div>

          <div>
            <CguCheckbox v-model="cgu" :has-error="showCguError" />
            <p v-if="showCguError" class="text-[13px] leading-[1.4] text-team-red mt-1">
              Tu dois accepter les conditions pour créer un compte.
            </p>
          </div>
        </div>

        <PrimaryButton :loading="loading" :disabled="!loading && !formValid" loading-label="Création…">
          {{ formValid ? 'Créer mon compte' : 'Compléter le formulaire' }}
        </PrimaryButton>

        <InlineError v-if="apiError">{{ apiError }}</InlineError>

        <div class="mt-5 text-[15px] text-fg-muted text-center">
          Déjà un compte ?
          <NuxtLink
            to="/login"
            class="text-fg-default underline underline-offset-[3px] decoration-fg-muted transition-colors hover:decoration-fg-default focus-visible:outline-none focus-visible:rounded-sm focus-visible:[box-shadow:0_0_0_2px_var(--color-bg-base),0_0_0_4px_rgba(32,128,255,0.45)]"
          >
            Se connecter
          </NuxtLink>
        </div>
      </form>
    </div>
  </main>
</template>
