export interface TocItem {
  id: string
  text: string
  level: number
}

function slugify(text: string) {
  const slug = text
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}\s-]/gu, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')

  return slug || 'heading'
}

export function extractTocItems(markdown: string): TocItem[] {
  const items: TocItem[] = []
  const used = new Map<string, number>()
  let inFence = false

  for (const line of markdown.split(/\r?\n/)) {
    if (/^\s*```/.test(line)) {
      inFence = !inFence
      continue
    }
    if (inFence) continue

    const match = /^(#{1,6})\s+(.+?)\s*#*\s*$/.exec(line)
    if (!match) continue

    const text = match[2].trim()
    if (!text) continue

    const base = slugify(text)
    const count = used.get(base) ?? 0
    used.set(base, count + 1)
    items.push({
      id: count === 0 ? base : `${base}-${count + 1}`,
      text,
      level: match[1].length,
    })
  }

  return items
}
