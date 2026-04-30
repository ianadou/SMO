export interface GroupDTO {
  id: string
  name: string
  organizer_id: string
  has_webhook: boolean
  created_at: string
}

export interface CreateGroupPayload {
  name: string
  discord_webhook_url?: string
}
