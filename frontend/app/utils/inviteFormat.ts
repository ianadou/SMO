const PARIS = 'Europe/Paris'

export function formatMatchDate(iso: string): string {
  const formatted = new Intl.DateTimeFormat('fr-FR', {
    timeZone: PARIS,
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(iso))
  return formatted.charAt(0).toUpperCase() + formatted.slice(1)
}

export function formatMatchTime(iso: string): string {
  const parts = new Intl.DateTimeFormat('fr-FR', {
    timeZone: PARIS,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).formatToParts(new Date(iso))
  const hour = parts.find(p => p.type === 'hour')?.value ?? '00'
  const minute = parts.find(p => p.type === 'minute')?.value ?? '00'
  return `${hour}h${minute}`
}

export function parseCapacity(capacity: string): { count: string, format: string } {
  const match = capacity.match(/^(\S+)\s*\(([^)]+)\)/)
  if (!match) return { count: capacity.trim(), format: '' }
  return { count: match[1], format: match[2] }
}
