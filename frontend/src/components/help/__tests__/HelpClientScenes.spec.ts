import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import HelpClientSettings from '../HelpClientSettings.vue'
import HelpOpenCodeTerminal from '../HelpOpenCodeTerminal.vue'
import HelpTerminalClientScene from '../HelpTerminalClientScene.vue'
import CursorGuide from '../guides/CursorGuide.vue'
import { desktopClientFields, isDesktopClient } from '@/utils/desktopClients'
import { extendedClientFiles, type ExtendedClientId } from '@/utils/extendedClients'
import { clientGuideComponents } from '../guides/guideRegistry'
import type { GroupPlatform } from '@/types'

const mocks = vi.hoisted(() => ({ copy: vi.fn() }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copyToClipboard: mocks.copy }) }))
const platforms: GroupPlatform[] = ['anthropic', 'openai', 'grok', 'gemini', 'antigravity', 'deepseek', 'minimax', 'composite']
const fieldsFor = (client: string, platform: GroupPlatform, model = 'selected-model') => isDesktopClient(client)
  ? desktopClientFields(client, 'https://site.example/v1', 'YOUR_API_KEY', model, platform)
  : extendedClientFiles(client as ExtendedClientId, platform, 'https://site.example/v1', 'YOUR_API_KEY', model)
function render(clientId: string, platform: GroupPlatform) {
  return mount(HelpClientSettings, { props: { clientId, clientName: clientId, fields: fieldsFor(clientId, platform) }, global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
}

describe('Client-specific tutorial scenes', () => {
  it.each(Object.keys(clientGuideComponents))('%s shows shared fields exactly once as copy targets across platforms', async client => {
    for (const platform of platforms) {
      const wrapper = render(client, platform)
      expect(wrapper.find(`[data-guide="${client}"]`).exists()).toBe(true)
      expect(wrapper.find('.app-workspace').exists()).toBe(false)
      for (const field of fieldsFor(client, platform)) {
        const buttons = wrapper.findAll('button').filter(button => button.attributes('aria-label') === `helpCenter.copyConfig ${field.path}`)
        expect(buttons).toHaveLength(1)
        await buttons[0]!.trigger('click')
        expect(mocks.copy).toHaveBeenLastCalledWith(field.content, 'helpCenter.copied')
      }
      expect(wrapper.text()).not.toContain('krill')
      expect(wrapper.text()).toContain('helpCenter.walkthrough.caption')
      expect(wrapper.text()).toContain('helpCenter.walkthrough.notTested')
      expect(wrapper.find('input, iframe, form').exists()).toBe(false)
      wrapper.unmount()
    }
  })

  it('keeps TRAE full URL off and leaves the protocol path for the client to append', async () => {
    const wrapper = render('trae', 'openai')
    expect(wrapper.find('[data-field="API 格式"] code').text()).toContain('OpenAI Chat Completions')
    expect(wrapper.find('[data-field="自定义请求地址"] code').text()).toBe('https://site.example/v1')
    expect(wrapper.text()).not.toContain('https://site.example/v1/chat/completions')
    expect(wrapper.find('[data-field="完整 URL"] .scene-switch.is-on').exists()).toBe(false)
    await wrapper.setProps({ fields: fieldsFor('trae', 'antigravity', 'new-model') })
    expect(wrapper.find('[data-field="API 格式"] code').text()).toContain('Anthropic Messages')
    expect(wrapper.find('[data-field="自定义请求地址"] code').text()).toBe('https://site.example/antigravity')
    expect(wrapper.text()).not.toContain('https://site.example/antigravity/v1/messages')
    expect(wrapper.text()).not.toContain('selected-model')
    expect(wrapper.find('[data-field="完整 URL"] .scene-switch.is-on').exists()).toBe(false)
    wrapper.unmount()
    const native = render('trae', 'deepseek')
    expect(native.find('[data-field="自定义请求地址"] code').text()).toBe('https://site.example')
    expect(native.find('[data-field="完整 URL"] .scene-switch.is-on').exists()).toBe(false)
  })

  it('shows distinct provider, editor, translation and workspace layouts', () => {
    expect(render('cherry-studio', 'openai').find('.cherry-add').exists()).toBe(true)
    expect(render('cursor', 'openai').text()).toContain('API Keys')
    expect(clientGuideComponents['read-frog']).toBeUndefined()
    expect(clientGuideComponents['roo-code']).toBeUndefined()
    expect(render('kiss-translator', 'openai').find('.translation-page').exists()).toBe(true)
    expect(render('workbuddy', 'openai').find('.wb-account .is-current').exists()).toBe(true)
    expect(render('workbuddy', 'openai').find('.wb-models aside .current').exists()).toBe(true)
    expect(render('workbuddy', 'openai').find('.workspace-conversation aside').exists()).toBe(true)
    expect(render('trae', 'openai').find('.trae-window').exists()).toBe(true)
    expect(render('trae', 'openai').find('.trae-chat').exists()).toBe(true)
    expect(render('cherry-studio', 'openai').find('.cherry-api').exists()).toBe(true)
    expect(render('cherry-studio', 'openai').find('.cherry-reply').exists()).toBe(true)
    const zcode = render('zcode', 'openai')
    expect(zcode.find('.zcode-window').exists()).toBe(true)
    expect(zcode.text()).toContain('helpCenter.zcodeUi.settings')
    expect(zcode.text()).toContain('helpCenter.zcodeUi.saveProvider')
    expect(zcode.text()).toContain('Choose model')
    expect(zcode.find('[data-field="Base URL"] code').text()).toBe('https://site.example/v1')
    expect(zcode.text()).not.toContain('/chat/completions')
    zcode.unmount()
  })
  it('shows SillyTavern and TavernAI chat fields without appending the protocol path', () => {
    const silly = render('sillytavern', 'anthropic')
    expect(silly.find('[data-step]').exists()).toBe(true)
    expect(silly.findAll('[data-step]').map(step => step.attributes('data-step'))).toEqual(['api', 'endpoint', 'model', 'reply'])
    expect(silly.find('[data-field="Custom Endpoint (Base URL)"] code').text()).toBe('https://site.example/v1')
    expect(silly.text()).toContain('Chat Completion')
    expect(silly.text()).not.toContain('https://site.example/v1/chat/completions')
    expect(silly.find('.st-check .box').exists()).toBe(true)
    silly.unmount()
    const tavern = render('tavernai', 'gemini')
    expect(tavern.findAll('[data-step]').map(step => step.attributes('data-step'))).toEqual(['open', 'custom', 'connect', 'message'])
    expect(tavern.find('[data-field="API Address"] code').text()).toBe('https://site.example/v1')
    expect(tavern.find('[data-field="Use direct API address"] .scene-switch.is-on').exists()).toBe(true)
    expect(tavern.find('[data-field="Model Endpoint"] code').text()).toBe('Chat Completions')
    expect(tavern.text()).not.toContain('/chat/completions')
    tavern.unmount()
  })

  it('renders Cursor settings in the same Network -> API Keys -> Models -> Chat order', () => {
    const wrapper = mount(CursorGuide, { props: { clientId: 'cursor', clientName: 'Cursor', fields: fieldsFor('cursor', 'openai') }, global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
    const steps = wrapper.findAll('[data-step]').map(step => step.attributes('data-step'))
    expect(steps).toEqual(['network', 'api-keys', 'chat'])
    expect(wrapper.findAll('.cursor-toggle.is-on')).toHaveLength(2)
    expect(wrapper.find('[data-field="Provider"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Add Model')
    expect(wrapper.find('.cursor-select').text()).toContain('HTTP/1.1')
    expect(wrapper.text()).toContain('HTTP Compatibility Mode')
    expect(wrapper.text()).toContain('Override OpenAI Base URL')
    expect(wrapper.find('[data-field="Override OpenAI Base URL"] code').text()).toBe('https://site.example/v1')
    wrapper.unmount()
  })

  it('escapes selected model text and updates OpenCode without fixed cost/version', async () => {
    const wrapper = mount(HelpOpenCodeTerminal, { props: { model: 'first-model', provider: 'Site', os: 'macos' } })
    expect(wrapper.text()).toContain('/models')
    expect(wrapper.text()).not.toContain('$0.01')
    expect(wrapper.text()).not.toContain('v0.3.133')
    await wrapper.setProps({ model: '<img src=x onerror=alert(1)>', os: 'windows' })
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.text()).toContain('<img src=x onerror=alert(1)>')
    expect(wrapper.text()).toContain('C:\\projects\\my-app')
    expect(wrapper.text()).not.toContain('first-model')
  })
  it('renders a distinct terminal scene for every CLI client', () => {
    for (const client of ['codex', 'dsh', 'pi', 'grok', 'openclaw', 'hermes'] as const) {
      const wrapper = mount(HelpTerminalClientScene, { props: { client, model: 'selected-model', os: 'macos', shell: 'bash' } })
      expect(wrapper.find(`.scene-${client}`).exists()).toBe(true)
      expect(wrapper.text()).toContain('selected-model')
      expect(wrapper.text()).toContain('helpCenter.walkthrough.notTested')
      wrapper.unmount()
    }
  })
  it('gives each supported GUI client its own component and distinct step sequence', () => {
    const components = Object.values(clientGuideComponents)
    expect(new Set(components).size).toBe(11)
    const sequences = Object.keys(clientGuideComponents).map(client => {
      const wrapper = render(client, 'openai')
      const steps = wrapper.findAll('[data-step]').map(step => step.attributes('data-step'))
      wrapper.unmount()
      return steps.join(',')
    })
    expect(new Set(sequences).size).toBe(11)
  })
  it('matches the live Kiss, Immersive and WorkBuddy screens', () => {
    const kiss = render('kiss-translator', 'anthropic')
    expect(kiss.text()).toContain('Sort Order')
    expect(kiss.text()).toContain('20480')
    expect(kiss.text()).toContain('formal')
    expect(kiss.find('[data-field="API URL"] code').text()).toBe('https://site.example/v1/messages')
    kiss.unmount()
    const kissOpen = render('kiss-translator', 'openai')
    expect(kissOpen.find('[data-field="API URL"] code').text()).toBe('https://site.example/v1/chat/completions')
    kissOpen.unmount()

    const immersive = render('immersive-translate', 'anthropic')
    expect(immersive.text()).toContain('Claude 1')
    expect(immersive.text()).toContain('APIKEY')
    expect(immersive.find('.context-switch.is-on').exists()).toBe(false)
    expect(immersive.find('[data-step="open-page"]').exists()).toBe(true)
    expect(immersive.find('[data-step="select-service"]').exists()).toBe(false)
    immersive.unmount()

    const workbuddy = render('workbuddy', 'openai')
    expect(workbuddy.text()).toContain('helpCenter.scenes.workbuddyToolCalling')
    expect(workbuddy.text()).toContain('helpCenter.scenes.workbuddyAutoReasoning')
    expect(workbuddy.find('[data-field="Custom Protocol"] .scene-switch.is-on').exists()).toBe(false)
    expect(workbuddy.find('[data-field="URL"] code').text()).toBe('https://site.example/v1/chat/completions')
    workbuddy.unmount()
  })

  it('gives dsh a separate reasoning-level screen after the web session screen', () => {
    const wrapper = render('dsh', 'openai')
    expect(wrapper.find('[data-step="new-session"]').exists()).toBe(true)
    expect(wrapper.find('[data-step="reasoning-level"]').exists()).toBe(true)
    expect(wrapper.find('.reasoning-options .selected').text()).toBe('medium')
    wrapper.unmount()
  })
  it('switches provider, endpoint and model together across native and OpenAI channels', async () => {
    for (const client of ['cherry-studio', 'kiss-translator', 'immersive-translate']) {
      const wrapper = render(client, 'anthropic')
      expect(wrapper.text()).toMatch(/Anthropic|Claude/)
      await wrapper.setProps({ fields: fieldsFor(client, 'openai', 'openai-model') })
      expect(wrapper.text()).toContain('OpenAI')
      expect(wrapper.text()).toContain('openai-model')
      expect(wrapper.text()).not.toContain('selected-model')
      expect(wrapper.text()).not.toContain('/messages')
      wrapper.unmount()
    }
  })
})
