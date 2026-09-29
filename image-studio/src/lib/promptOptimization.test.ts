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

  it('does not report a model when the prompt is unchanged or the request fails', () => {
    expect(parseStudioPromptOptimization('一只猫', 200, { prompt: '一只猫', optimized: false })).toEqual({
      prompt: '一只猫',
      optimized: false,
    })
    expect(() => parseStudioPromptOptimization('一只猫', 502, { error: { message: '上游失败' } })).toThrow('上游失败')
  })
})
