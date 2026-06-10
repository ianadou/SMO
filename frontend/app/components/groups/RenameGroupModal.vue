<script setup lang="ts">
import TextField from '~/components/login/TextField.vue'
import type { GroupDTO } from '~/types/groups'
import { ApiError } from '~/composables/useApi'

const props = defineProps<{
  open: boolean
  group: GroupDTO
}>()
const emit = defineEmits<{
  close: []
  renamed: [group: GroupDTO]
}>()

const NAME_MAX = 100

const name = ref(props.group.name)
const submitting = ref(false)
const error = ref('')

const trimmedName = computed(() => name.value.trim())
const nameTooLong = computed(() => name.value.length > NAME_MAX)
const canSubmit = computed(
  () =>
    trimmedName.value.length > 0
    && !nameTooLong.value
    && !submitting.value
    && trimmedName.value !== props.group.name,
)

function close() {
  if (submitting.value) return
  emit('close')
}

async function submit() {
  if (!canSubmit.value) return
  error.value = ''
  submitting.value = true
  try {
    const api = useApi()
    const renamed = await api.patch<GroupDTO>(`/groups/${props.group.id}`, {
      name: trimmedName.value,
    })
    emit('renamed', renamed)
  }
  catch (e) {
    error.value = e instanceof ApiError
      ? e.publicMessage
      : 'Connexion au serveur impossible. Réessaie dans un instant.'
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      name.value = props.group.name
      error.value = ''
    }
  },
)
</script>

<template>
  <BaseModal :open="open" title="Renommer le groupe" :close-disabled="submitting" @close="close">
    <form class="flex flex-col gap-4" @submit.prevent="submit">
      <div class="flex flex-col gap-2">
        <TextField
          id="rename-group-name"
          v-model="name"
          label="Nom du groupe"
          autocomplete="off"
          :has-error="nameTooLong"
        />
        <div class="flex justify-between text-xs text-fg-muted">
          <span v-if="nameTooLong" class="text-team-red">
            Trop long (max {{ NAME_MAX }} caractères)
          </span>
          <span v-else>Visible par tes joueurs sur l'invitation</span>
          <span :class="nameTooLong && 'text-team-red'">
            {{ name.length }}/{{ NAME_MAX }}
          </span>
        </div>
      </div>

      <ModalActions
        :error="error"
        :submitting="submitting"
        :can-submit="canSubmit"
        submit-label="Renommer"
        @cancel="close"
      />
    </form>
  </BaseModal>
</template>
