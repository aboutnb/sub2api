import { computed } from 'vue'
export interface GuideField { path: string; content: string }
export interface ClientGuideProps { fields: GuideField[]; clientId: string; clientName: string }
export function useGuideFields(props: ClientGuideProps) {
  const field = (path: string) => props.fields.find(item => item.path === path)!
  const modelField = computed(() => props.fields.find(item => /^(Model( ID)?|模型( ID)?|模型名称)$/i.test(item.path)))
  const model = computed(() => modelField.value?.content || 'YOUR_MODEL_ID')
  const provider = computed(() => props.fields.find(item => item.path === 'Provider')?.content || '')
  return { field, modelField, model, provider }
}
