import { baseCompile } from '@intlify/message-compiler'
import { createI18n } from 'vue-i18n'
import zh from '@/i18n/locales/zh'
import en from '@/i18n/locales/en'

// The test runtime omits the message compiler; use compiled functions with real t().
function compileMessages<T>(value: T): T {
  if (typeof value === 'string') return new Function(`return ${baseCompile(value).code}`)() as T
  if (value && typeof value === 'object') {
    return Object.fromEntries(Object.entries(value).map(([key, child]) => [key, compileMessages(child)])) as T
  }
  return value
}

export function createTicketI18n(locale = 'zh') {
  return createI18n({
    legacy: false, locale, missingWarn: false, fallbackWarn: false,
    messages: {
      zh: compileMessages({ supportTickets: zh.supportTickets, common: zh.common, pagination: zh.pagination }),
      en: compileMessages({ supportTickets: en.supportTickets, common: en.common, pagination: en.pagination })
    }
  })
}
