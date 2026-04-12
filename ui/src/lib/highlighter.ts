import { createHighlighter, type Highlighter } from 'shiki'

let highlighter: Highlighter | null = null

export async function getHighlighter(): Promise<Highlighter> {
  if (!highlighter) {
    highlighter = await createHighlighter({
      themes: ['vitesse-dark'],
      langs: ['csharp'],
    })
  }
  return highlighter
}

export async function highlightLines(lines: string[], lang = 'csharp'): Promise<string[]> {
  const h = await getHighlighter()
  return lines.map(line => {
    const html = h.codeToHtml(line || ' ', { lang, theme: 'vitesse-dark' })
    const match = html.match(/<code[^>]*>([\s\S]*)<\/code>/)
    return match ? match[1].replace(/<span class="line">([\s\S]*?)<\/span>/, '$1') : line
  })
}
