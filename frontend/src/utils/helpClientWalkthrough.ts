export interface ClientWalkthrough {
  layout: 'provider' | 'editor' | 'extension' | 'dialog'
  accent: string
  entry: string
  navigation: string[]
  section: string
  add: string
  form: string
  save: string
  modelAction: string
  modelLocation: string
}

export const detailedSceneClients = ['cherry-studio', 'cursor', 'trae', 'workbuddy', 'zcode', 'read-frog', 'kiss-translator', 'immersive-translate']

// UI labels describe the client, while all connection values come from the shared generator.
export const clientWalkthroughs: Record<string, ClientWalkthrough> = {
  'cherry-studio': {
    layout: 'provider', accent: '#d65370', entry: '设置 → 模型服务',
    navigation: ['常规设置', '模型服务', '默认模型', '网络设置'], section: '模型服务',
    add: '添加供应商', form: '供应商设置', save: '保存并启用', modelAction: '获取模型列表 / 添加模型', modelLocation: '新对话顶部 → 模型选择器'
  },
  cursor: {
    layout: 'editor', accent: '#526174', entry: 'Cursor Settings → Models',
    navigation: ['General', 'Models', 'Rules', 'Network'], section: 'Models',
    add: 'OpenAI API Key', form: 'API Keys', save: 'Save', modelAction: 'Chat model', modelLocation: 'Chat → Model picker'
  },
  cline: {
    layout: 'editor', accent: '#526174', entry: 'VS Code → Cline → Settings',
    navigation: ['API Configuration', 'Feature Settings', 'Browser', 'Terminal'], section: 'API Configuration',
    add: 'API Provider', form: 'OpenAI Compatible', save: 'Done', modelAction: 'Model ID', modelLocation: 'Cline → Plan / Act → API Configuration'
  },
  'roo-code': {
    layout: 'editor', accent: '#557c64', entry: 'VS Code → Roo Code → Settings',
    navigation: ['Providers', 'Auto-Approve', 'Browser', 'Terminal'], section: 'Providers',
    add: 'API Configuration', form: 'OpenAI Compatible', save: 'Save', modelAction: 'Model ID', modelLocation: 'Roo Code → Active API Configuration'
  },
  dsh: {
    layout: 'provider', accent: '#4777ee', entry: 'dsh Web → Settings → Models',
    navigation: ['General', 'Models', 'Tools'], section: 'Models',
    add: 'Add model provider', form: 'Custom model API', save: 'Save', modelAction: 'Fetch available models / Model ID', modelLocation: 'Chat → Model → site / model'
  },
  workbuddy: {
    layout: 'dialog', accent: '#088c74', entry: '头像 → 设置 → 模型',
    navigation: ['通用', '模型', '外观'], section: '模型',
    add: '添加模型', form: '编辑模型', save: '保存', modelAction: '模型名称', modelLocation: '任务界面 → 输入框右下角 → 模型'
  },
  zcode: {
    layout: 'provider', accent: '#525252', entry: 'Model selector → Manage models → Model settings',
    navigation: ['General', 'Model settings', 'Tools'], section: 'Model settings',
    add: 'Add provider', form: 'Provider settings', save: 'Save', modelAction: 'Add model', modelLocation: 'Workspace → Choose model → Provider'
  },
  trae: {
    layout: 'dialog', accent: '#168b63', entry: '模型设置 → 添加模型 → 自定义配置',
    navigation: ['账号', '模型', '偏好设置'], section: '模型',
    add: '添加模型', form: '自定义配置', save: '提交', modelAction: '模型 ID', modelLocation: 'SOLO → 输入框 → 模型选择器'
  },
  'read-frog': {
    layout: 'extension', accent: '#64843c', entry: '扩展选项 → API 服务商',
    navigation: ['翻译设置', 'API 服务商', 'AI 功能'], section: 'API 服务商',
    add: '添加服务商', form: 'OpenAI 兼容服务商', save: '保存', modelAction: '模型标识', modelLocation: '翻译设置 → 当前翻译服务商 / 模型'
  },
  'kiss-translator': {
    layout: 'extension', accent: '#3786b0', entry: '扩展设置 → 翻译服务 → OpenAI',
    navigation: ['翻译服务', '翻译设置', '规则'], section: '翻译服务',
    add: '添加翻译服务', form: 'OpenAI', save: '保存', modelAction: '自定义模型', modelLocation: '翻译面板 → 当前翻译服务'
  },
  'immersive-translate': {
    layout: 'extension', accent: '#bd4d85', entry: '设置 → 翻译服务 → OpenAI → 更多设置',
    navigation: ['基本设置', '翻译服务', '高级设置'], section: '翻译服务',
    add: 'OpenAI', form: '更多设置', save: '保存', modelAction: '自定义模型', modelLocation: '翻译面板 → 翻译服务 → OpenAI'
  },
  sillytavern: {
    layout: 'dialog', accent: '#7c6af7', entry: 'API → Chat Completion',
    navigation: ['API', 'Chat Completion', 'Custom'], section: 'Chat Completion',
    add: 'Custom (OpenAI-compatible)', form: 'Custom Endpoint', save: 'Connect', modelAction: 'Model ID', modelLocation: 'Chat → selected model'
  },
  tavernai: {
    layout: 'dialog', accent: '#e0a100', entry: 'AI Provider Connection',
    navigation: ['Provider', 'Custom', 'Connect'], section: 'AI Provider',
    add: 'Custom', form: 'API Address', save: 'Connect', modelAction: 'Model ID', modelLocation: 'Chat → connected provider'
  }
}
