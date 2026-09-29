import { describe, expect, it } from 'vitest'
import { renderHelpMarkdown } from '../helpMarkdown'

describe('renderHelpMarkdown', () => {
  it('removes executable markup and unsafe URL protocols', () => {
    const rendered = renderHelpMarkdown('<script>alert(1)</script> [bad](javascript:alert(1)) [good](https://example.com)')
    expect(rendered.html).not.toContain('<script')
    expect(rendered.html).not.toMatch(/href=["']javascript:/i)
    expect(rendered.html).toContain('https://example.com')
  })

  it('extracts headings and copies only preformatted blocks', () => {
    const rendered = renderHelpMarkdown('# Title\n\n```bash\necho ok\n```')
    expect(rendered.headings).toEqual([{ id: 'help-section-0', text: 'Title' }])
    expect(rendered.blocks).toEqual(['echo ok\n'])
  })

  it('keeps status tables and inline codes', () => {
    const rendered = renderHelpMarkdown('| 状态 | 优先检查 |\n| --- | --- |\n| `400` | 请求结构 |')
    expect(rendered.html).toContain('<table>')
    expect(rendered.html).toContain('<code>400</code>')
    expect(rendered.html).toContain('请求结构')
  })
})
