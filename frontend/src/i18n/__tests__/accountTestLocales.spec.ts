import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe.each([
  ['en', en],
  ['zh', zh]
])('account test locale keys in %s', (_locale, messages) => {
  it('defines all locally generated errors and SSE progress messages', () => {
    expect(messages.admin.accounts.testError).toMatchObject({
      requestFailed: expect.any(String),
      noResponseBody: expect.any(String),
      unknown: expect.any(String),
      line: expect.any(String)
    })
    expect(messages.admin.accounts.testStatus).toMatchObject({
      chatCompletionsTesting: expect.any(String),
      chatCompletionsVerified: expect.any(String),
      codexImageToolCalling: expect.any(String)
    })
    expect(messages.admin.accounts.bulkActions.batchTest).toEqual(expect.any(String))
    expect(messages.admin.accounts.batchTest).toMatchObject({
      title: expect.any(String),
      menu: expect.any(String),
      loadingAccounts: expect.any(String),
      noAccounts: expect.any(String),
      loadFailed: expect.any(String),
      columns: {
        account: expect.any(String),
        platform: expect.any(String),
        status: expect.any(String),
        result: expect.any(String)
      },
      status: {
        pending: expect.any(String),
        running: expect.any(String),
        success: expect.any(String),
        failed: expect.any(String),
        stopped: expect.any(String)
      },
      filters: {
        all: expect.any(String),
        success: expect.any(String),
        failed: expect.any(String),
        unauthorized: expect.any(String),
        rateLimited: expect.any(String),
        otherFailed: expect.any(String)
      }
    })
  })
})

describe('account test locale language boundaries', () => {
  it('keeps the English SSE progress copy in English', () => {
    expect(en.admin.accounts.testStatus.chatCompletionsTesting).toBe(
      'Testing connection through /v1/chat/completions'
    )
    expect(en.admin.accounts.testStatus.chatCompletionsVerified).toBe(
      'Verified through /v1/chat/completions'
    )
  })

  it('keeps the Chinese SSE progress copy in Chinese', () => {
    expect(zh.admin.accounts.testStatus.chatCompletionsTesting).toBe(
      '正在通过 /v1/chat/completions 测试连接'
    )
    expect(zh.admin.accounts.testStatus.chatCompletionsVerified).toBe(
      '已通过 /v1/chat/completions 验证'
    )
  })
})
