import { describe, expect, it } from 'vitest'
import { parseStudioPromptOptimization } from './promptOptimization'

describe('parseStudioPromptOptimization', () => {
  it('keeps the rewritten prompt and the optimizer model', () => {
    expect(parseStudioPromptOptimization('一只猫', 200, {
      prompt: '一只橘色猫咪坐在窗边，柔和晨光',
      optimized: true,
      model: 'gpt-5.4-mini',
      group: '智能分组',
    })).toEqual({
      prompt: '一只橘色猫咪坐在窗边，柔和晨光',
      optimized: true,
      model: 'gpt-5.4-mini',
      group: '智能分组',
    })
  })

  it('does not invent model metadata when the response contains none', () => {
    expect(parseStudioPromptOptimization('一只猫', 200, { prompt: '一只猫', optimized: false })).toEqual({
      prompt: '一只猫',
      optimized: false,
    })
    expect(() => parseStudioPromptOptimization('一只猫', 502, { error: { message: '上游失败' } })).toThrow('上游失败')
  })
  it('retains the model that was used even when the optimizer keeps the original text', () => {
    expect(parseStudioPromptOptimization('一只猫', 200, { prompt: '一只猫', optimized: false, model: 'gpt-5.6-sol', group: '生图智能路由' })).toEqual({
      prompt: '一只猫', optimized: false, model: 'gpt-5.6-sol', group: '生图智能路由',
    })
  })
  it('preserves permission and session errors returned in the standard envelope', () => {
    expect(() => parseStudioPromptOptimization('一只猫', 401, { message: '登录已过期' })).toThrow('登录已过期')
    expect(() => parseStudioPromptOptimization('一只猫', 403, { error: { message: '当前分组无可用模型' } })).toThrow('当前分组无可用模型')
  })
})
