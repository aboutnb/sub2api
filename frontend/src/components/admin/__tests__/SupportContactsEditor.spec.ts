import { describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { mount } from '@vue/test-utils'
import SupportContactsEditor from '../SupportContactsEditor.vue'
import SupportContactList from '@/components/common/SupportContactList.vue'
import { parseSupportContacts, serializeSupportContacts } from '@/utils/supportContacts'
const copy = vi.hoisted(() => vi.fn().mockResolvedValue(true))
vi.mock('@/composables/useClipboard', () => ({ useClipboard: () => ({ copied: false, copyToClipboard: copy }) }))
const entries = [
  { name: '在线客服', tag: '微信', icon: 'chat' as const, account: 'hello', url: '' },
  { name: '社群', tag: '交流', icon: 'users' as const, account: '', url: 'https://example.com' }
]
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
const global = { stubs: { BaseDialog: true } }
describe('support contacts', () => {
  it('edits, reorders, adds and removes entries through v-model', async () => {
    const host = mount(defineComponent({
      components: { SupportContactsEditor },
      setup() { return { value: ref(serializeSupportContacts(entries)) } },
      template: '<SupportContactsEditor v-model="value" />'
    }), { global })
    await host.find('input').setValue('专属客服')
    expect(parseSupportContacts(host.vm.value)[0].name).toBe('专属客服')
    await host.find('button[aria-label="common.support.moveDown"]').trigger('click')
    expect(parseSupportContacts(host.vm.value)[0].name).toBe('社群')
    await host.findAll('button').find(button => button.text().includes('common.support.add'))!.trigger('click')
    expect(parseSupportContacts(host.vm.value)).toHaveLength(3)
    await host.find('button[aria-label="common.delete"]').trigger('click')
    expect(parseSupportContacts(host.vm.value)).toHaveLength(2)
  })
  it('only shows link actions when present and safe, and copies the selected account', async () => {
    const wrapper = mount(SupportContactList, { props: { contacts: [...entries, { ...entries[0], url: 'javascript:alert(1)' }] }, global })
    expect(wrapper.findAll('a')).toHaveLength(1)
    expect(wrapper.find('a').attributes('rel')).toBe('noopener noreferrer')
    expect(wrapper.findAll('li')[0].find('a').exists()).toBe(false)
    await wrapper.find('button').trigger('click')
    expect(copy).toHaveBeenCalledWith('hello')
  })
})
