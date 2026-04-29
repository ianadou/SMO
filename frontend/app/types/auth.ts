export interface OrganizerDTO {
  id: string
  email: string
  display_name: string
  created_at: string
}

export interface LoginResponseDTO {
  token: string
  organizer: OrganizerDTO
}

export interface RegisterPayload {
  email: string
  password: string
  display_name: string
}

export interface LoginPayload {
  email: string
  password: string
}

export interface ApiErrorResponse {
  error: string
}
