export function playerInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return ''
  const first = parts[0] ?? ''
  if (parts.length === 1) return first.slice(0, 2).toUpperCase()
  const last = parts.at(-1) ?? ''
  return (first[0] ?? '').toUpperCase() + (last[0] ?? '').toUpperCase()
}
