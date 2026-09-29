import DOMPurify from 'dompurify'
import { marked } from 'marked'

export function renderHelpMarkdown(content: string) {
  const template = document.createElement('template')
  template.innerHTML = DOMPurify.sanitize(marked.parse(content, { async: false }), {
    USE_PROFILES: { html: true }, FORBID_TAGS: ['form', 'input', 'button', 'style', 'iframe'],
    FORBID_ATTR: ['style'],
    ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto):|\/|#)/i
  })
  const headings = Array.from(template.content.querySelectorAll('h1,h2,h3')).map((heading, index) => {
    const id = `help-section-${index}`
    heading.id = id
    return { id, text: heading.textContent || '' }
  })
  template.content.querySelectorAll('a').forEach(link => { link.rel = 'noopener noreferrer' })
  const blocks = Array.from(template.content.querySelectorAll('pre')).map(block => block.textContent || '')
  return { html: template.innerHTML, headings, blocks }
}
