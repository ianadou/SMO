const STRENGTH_LABELS = ['', 'Faible', 'Moyen', 'Fort', 'Très fort'] as const

export type StrengthLevel = 0 | 1 | 2 | 3 | 4

export function passwordStrength(value: string): StrengthLevel {
  if (!value) return 0
  const classes
    = (/[a-z]/.test(value) ? 1 : 0)
    + (/[A-Z]/.test(value) ? 1 : 0)
    + (/[0-9]/.test(value) ? 1 : 0)
    + (/[^A-Za-z0-9]/.test(value) ? 1 : 0)
  if (value.length < 8) return 1
  if (value.length >= 12 && classes >= 3) return 4
  if (classes >= 2 && value.length >= 8) return 3
  return 2
}

export function passwordStrengthLabel(level: StrengthLevel): string {
  return STRENGTH_LABELS[level]
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function isEmailFormat(value: string): boolean {
  return EMAIL_RE.test(value)
}
