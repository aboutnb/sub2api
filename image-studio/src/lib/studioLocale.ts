import { useSyncExternalStore } from 'react'

let locale = 'zh'
const listeners = new Set<() => void>()
const subscribe = (listener: () => void) => { listeners.add(listener); return () => { listeners.delete(listener) } }
const snapshot = () => locale
export function useStudioLocale() { return useSyncExternalStore(subscribe, snapshot) }
export function setStudioLocale(value: string) {
  const next = value?.startsWith('en') ? 'en' : 'zh'
  if (next === locale) return
  locale = next
  for (const listener of listeners) listener()
}

const english: Record<string, string> = {
  '绘图参数': 'Image parameters',
  '生图模型': 'Image model', '加载中...': 'Loading...', 'AI 绘图暂未开放': 'Image studio is unavailable',
  '当前账号没有可用的生图模型': 'No image models are available for this account',
  '收藏夹': 'Collections', '返回收藏夹': 'Back to collections', '退出收藏夹': 'Exit collections', '管理收藏夹': 'Manage collections',
  '全部': 'All', '已完成': 'Completed', '生成中': 'Generating', '失败': 'Failed', '没有失败记录': 'No failed tasks',
  '搜索收藏夹名称...': 'Search collections...', '搜索提示词、参数...': 'Search prompts and parameters...',
  '清除失败记录': 'Clear failed tasks', '清除': 'Clear', '取消': 'Cancel',
  '尺寸': 'Size', '质量': 'Quality', '格式': 'Format', '透明背景': 'Transparent', '压缩率': 'Compression', '审核': 'Moderation', '数量': 'Count',
  '停止生成': 'Stop generation', '遮罩编辑': 'Mask edit', '生成图像': 'Generate image', '请先配置 API': 'Select a model first',
  '描述你想生成的图片，可输入 @ 来指定参考图...': 'Describe your image; use @ to reference an input image...',
  '上传图片': 'Upload image', '清空文本': 'Clear prompt', '清空': 'Clear', '清空全部': 'Clear all', '查看参考图': 'View reference image',
  '生成中...': 'Generating...', '正在生成……': 'Generating...', '重连中': 'Reconnecting', '已停止': 'Stopped', '预览': 'Preview',
  '(无提示词)': '(No prompt)', '局部重绘': 'Inpainting', '重试任务': 'Retry task', '编辑收藏夹': 'Edit collections', '收藏任务': 'Favorite task',
  '复用配置': 'Reuse settings', '编辑输出': 'Edit output', '删除任务': 'Delete task', '关闭': 'Close',
  '下载图片': 'Download image', '下载全部': 'Download all', '下载原图': 'Download original', '下载中间步骤图': 'Download intermediate images',
  '下载成功': 'Downloaded', '原图下载成功': 'Original downloaded', '下载失败': 'Download failed',
  '输入内容': 'Prompt', '参考图': 'Reference images', '参数配置': 'Parameters', '来源': 'Source',
  '复制提示词': 'Copy prompt', '提示词已复制': 'Prompt copied', '复制参考图': 'Copy reference image', '参考图已复制': 'Reference image copied',
  '复制完整报错': 'Copy error', '完整报错已复制': 'Error copied', '查看原始响应': 'View raw response',
  '复制图片链接': 'Copy image URL', '图片链接已复制': 'Image URL copied', '复制链接': 'Copy URL', '复制': 'Copy', '全部复制': 'Copy all',
  '复制成功': 'Copied', '复制失败': 'Copy failed', '原始响应数据': 'Raw response', '生成失败': 'Generation failed', '未知': 'Unknown', '未设置': 'Not set',
  '替换图片': 'Replace image', '编辑图片': 'Edit image', '只能有一张遮罩图': 'Only one mask is supported',
  '请选择有效图片': 'Select a valid image', '原参考图已不存在': 'Reference image no longer exists', '参考图未变化': 'Reference image is unchanged',
  '这张图片已在参考图中': 'This image is already included', '参考图已替换': 'Reference image replaced',
  '编辑遮罩': 'Edit mask', '遮罩编辑说明': 'About mask editing', '移除遮罩': 'Remove mask', '已移除遮罩': 'Mask removed',
  '保存': 'Save', '保存中...': 'Saving...', '遮罩已保存': 'Mask saved', '正在载入图片...': 'Loading image...',
  '画笔': 'Brush', '橡皮': 'Eraser', '调节笔刷大小': 'Brush size', '撤销': 'Undo', '重做': 'Redo', '重置视图': 'Reset view', '清空遮罩': 'Clear mask',
  '当前浏览器不支持 Canvas': 'Canvas is unavailable', '遮罩尺寸与当前图片不一致': 'Mask dimensions do not match the image',
  '图片已不存在，无法编辑遮罩': 'Image no longer exists', '确定要撤销对这张图片的所有涂抹并移除遮罩吗？': 'Remove all strokes and the mask?',
  '确定要删除这个任务吗？关联的图片资源也会被清理（如果没有其他任务引用）。': 'Delete this task and images not used by other tasks?',
  '输入内容将在响应完成时接收': 'The prompt will arrive with the completed response',
  '根据官方文档说明，此功能仅基于提示词，无法完全控制模型编辑区域。': 'The model may also change areas outside the mask.',
  '建议附加类似“只编辑遮罩区域”的提示词以提升模型指令遵循程度。': 'You can specify that only the masked area should change.',
}

export function studioText(text: string) { return locale === 'en' ? english[text] ?? text : text }
