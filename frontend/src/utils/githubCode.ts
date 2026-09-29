import { createHighlighterCore, type HighlighterCore } from 'shiki/core'
import { createJavaScriptRegexEngine } from 'shiki/engine/javascript'
import githubDark from '@shikijs/themes/github-dark-default'
import githubLight from '@shikijs/themes/github-light-default'
import bash from '@shikijs/langs/bash'
import batch from '@shikijs/langs/batch'
import css from '@shikijs/langs/css'
import diff from '@shikijs/langs/diff'
import dotenv from '@shikijs/langs/dotenv'
import html from '@shikijs/langs/html'
import ini from '@shikijs/langs/ini'
import javascript from '@shikijs/langs/javascript'
import json from '@shikijs/langs/json'
import markdown from '@shikijs/langs/markdown'
import powershell from '@shikijs/langs/powershell'
import python from '@shikijs/langs/python'
import sql from '@shikijs/langs/sql'
import toml from '@shikijs/langs/toml'
import typescript from '@shikijs/langs/typescript'
import vue from '@shikijs/langs/vue'
import xml from '@shikijs/langs/xml'
import yaml from '@shikijs/langs/yaml'

const themes = {
  light: 'github-light-default',
  dark: 'github-dark-default',
} as const

const languages = [
  json, bash, powershell, python, toml, yaml, dotenv, batch,
  javascript, typescript, html, css, markdown, xml, diff, sql, vue, ini,
]

const shikiLanguage: Record<string, string> = {
  shell: 'bash',
  env: 'dotenv',
  cmd: 'cmd',
  text: 'text',
}

let highlighter: HighlighterCore | null = null
let ready: Promise<void> | null = null
const listeners = new Set<() => void>()

export function escapeGithubCode(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

export function inferCodeLanguage(label?: string): string {
  const value = (label || '').trim().toLowerCase()
  if (!value || value === 'code' || value === 'text' || value === 'plaintext' || value === 'txt') return 'text'
  if (value.includes('powershell') || value.endsWith('.ps1')) return 'powershell'
  if (value.includes('python') || value.endsWith('.py')) return 'python'
  if (value.includes('command prompt') || value.endsWith('.bat') || value.endsWith('.cmd') || /(^|[^a-z])(cmd|bat|batch)([^a-z]|$)/.test(value)) return 'cmd'
  if (/(^|[^a-z])(bash|shell|zsh|terminal|curl)([^a-z]|$)/.test(value) || value === 'sh' || value.endsWith('.sh')) return 'shell'
  if (value === 'json' || value.endsWith('.json')) return 'json'
  if (value === 'toml' || value.endsWith('.toml')) return 'toml'
  if (value === 'yaml' || value === 'yml' || value.endsWith('.yaml') || value.endsWith('.yml')) return 'yaml'
  if (value === 'env' || value.endsWith('.env') || value.includes('environment')) return 'env'
  if (value === 'javascript' || value === 'js' || value.endsWith('.js') || value.endsWith('.mjs') || value.endsWith('.cjs')) return 'javascript'
  if (value === 'typescript' || value === 'ts' || value.endsWith('.ts')) return 'typescript'
  if (value === 'html' || value.endsWith('.html')) return 'html'
  if (value === 'css' || value.endsWith('.css')) return 'css'
  if (value === 'markdown' || value === 'md' || value.endsWith('.md')) return 'markdown'
  if (value === 'xml' || value.endsWith('.xml')) return 'xml'
  if (value === 'sql' || value.endsWith('.sql')) return 'sql'
  if (value === 'diff' || value === 'patch' || value.endsWith('.diff')) return 'diff'
  if (value === 'vue' || value.endsWith('.vue')) return 'vue'
  if (value === 'ini' || value.endsWith('.ini')) return 'ini'
  return 'text'
}

function shikiLang(language?: string): string {
  const inferred = inferCodeLanguage(language)
  return shikiLanguage[inferred] || inferred
}

export function githubHighlighterReady(): Promise<void> {
  if (!ready) {
    ready = createHighlighterCore({
      themes: [githubLight, githubDark],
      langs: languages,
      engine: createJavaScriptRegexEngine(),
    }).then((instance) => {
      highlighter = instance
      listeners.forEach((listener) => listener())
      listeners.clear()
    }).catch((error: unknown) => {
      listeners.clear()
      throw error
    })
  }
  return ready
}

export function onGithubHighlighter(listener: () => void): () => void {
  if (highlighter) {
    listener()
    return () => {}
  }
  listeners.add(listener)
  void githubHighlighterReady()
  return () => listeners.delete(listener)
}

function innerCode(html: string): string {
  return html.match(/<code[^>]*>([\s\S]*)<\/code>/)?.[1] ?? ''
}

export function highlightGithubCode(source: string, language?: string): string {
  if (!highlighter) {
    void githubHighlighterReady()
    return escapeGithubCode(source)
  }
  const render = (lang: string) => highlighter!.codeToHtml(source, {
    lang,
    themes,
    defaultColor: false,
  })
  try {
    let html = innerCode(render(shikiLang(language)))
    if (source.endsWith('\n') && !source.endsWith('\n\n')) {
      html = html.replace(/<span class="line"><\/span>$/, '')
    }
    return html
  } catch {
    return innerCode(render('text')) || escapeGithubCode(source)
  }
}

void githubHighlighterReady()
