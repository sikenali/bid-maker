export interface MarkdownSection {
  title: string
  level: number
  children?: MarkdownSection[]
}

export function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

export function serializeOutlineToMarkdown(sections: MarkdownSection[]): string {
  const parts: string[] = []
  const walk = (secs: MarkdownSection[]) => {
    for (const s of secs) {
      parts.push('#'.repeat(s.level) + ' ' + s.title.trim())
      if (s.children && s.children.length) walk(s.children)
    }
  }
  walk(sections)
  return parts.join('\n\n')
}

function stripOwnHeading(content: string, title: string, level: number): string {
  const lines = content.split('\n')
  let i = 0
  while (i < lines.length && lines[i].trim() === '') i++
  if (i >= lines.length) return content
  const headingRe = new RegExp(`^#{${level}}\\s+${escapeRegex(title.trim())}\\s*$`)
  if (headingRe.test(lines[i].trim())) {
    return lines.slice(i + 1).join('\n')
  }
  return content
}

export function setSectionMarkdownContent(
  md: string,
  title: string,
  level: number,
  content: string,
): string {
  const lines = md.split('\n')
  const headingRe = new RegExp(`^#{${level}}\\s+${escapeRegex(title.trim())}\\s*$`)
  let headingIdx = -1
  for (let i = 0; i < lines.length; i++) {
    if (headingRe.test(lines[i].trim())) {
      headingIdx = i
      break
    }
  }

  const body = stripOwnHeading(content, title, level).trim()

  if (headingIdx === -1) {
    if (!body) return md
    const heading = '#'.repeat(level) + ' ' + title.trim()
    return (md ? md.replace(/\s+$/, '') + '\n\n' : '') + heading + '\n\n' + body
  }

  const endHeadingRe = /^(#{1,6})\s+/
  let endIdx = lines.length
  for (let i = headingIdx + 1; i < lines.length; i++) {
    const m = lines[i].trim().match(endHeadingRe)
    if (m && m[1].length <= level) {
      endIdx = i
      break
    }
  }

  const before = lines.slice(0, headingIdx + 1).join('\n')
  const after = lines.slice(endIdx).join('\n').trim()

  let result = before
  if (body) result += '\n\n' + body
  if (after) result += '\n\n' + after
  return result
}