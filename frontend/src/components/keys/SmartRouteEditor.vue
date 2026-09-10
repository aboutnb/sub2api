<template>
  <section class="space-y-4" aria-labelledby="smart-route-mode-label">
    <div>
      <label id="smart-route-mode-label" class="input-label">{{ t('smartRouting.mode') }}</label>
      <div class="grid grid-cols-2 rounded-lg bg-surface-muted p-1 dark:bg-surface-muted" role="radiogroup">
        <button
          type="button"
          role="radio"
          :aria-checked="modelValue.mode === 'single'"
          :class="modeButtonClass(modelValue.mode === 'single')"
          @click="setMode('single')"
        >
          {{ t('smartRouting.single') }}
        </button>
        <button
          type="button"
          role="radio"
          :aria-checked="modelValue.mode === 'smart'"
          :disabled="!enabled"
          :class="modeButtonClass(modelValue.mode === 'smart')"
          @click="setMode('smart')"
        >
          <Icon name="sparkles" size="sm" />
          {{ t('smartRouting.smart') }}
        </button>
      </div>
      <p class="input-hint mt-2">
        {{ t(modelValue.mode === 'smart' ? 'smartRouting.smartHint' : 'smartRouting.singleHint') }}
      </p>
    </div>

    <div v-if="modelValue.mode === 'single'">
      <label class="input-label">{{ t('keys.groupLabel') }}</label>
      <Select
        :model-value="modelValue.group_id"
        :options="groupOptions"
        :placeholder="t('keys.selectGroup')"
        :searchable="true"
        :search-placeholder="t('keys.searchGroup')"
        data-tour="key-form-group"
        @update:model-value="setGroup"
      >
        <template #selected="{ option }">
          <GroupBadge
            v-if="option"
            :name="groupOption(option).label"
            :platform="groupOption(option).group.platform"
            :subscription-type="groupOption(option).group.subscription_type"
            :rate-multiplier="groupOption(option).group.rate_multiplier"
            :user-rate-multiplier="groupOption(option).userRate"
            :peak-rate-enabled="groupOption(option).group.peak_rate_enabled"
            :peak-start="groupOption(option).group.peak_start"
            :peak-end="groupOption(option).group.peak_end"
            :peak-rate-multiplier="groupOption(option).group.peak_rate_multiplier"
          />
          <span v-else class="text-ink-muted">{{ t('keys.selectGroup') }}</span>
        </template>
        <template #option="{ option, selected }">
          <GroupOptionItem
            :name="groupOption(option).label"
            :platform="groupOption(option).group.platform"
            :subscription-type="groupOption(option).group.subscription_type"
            :rate-multiplier="groupOption(option).group.rate_multiplier"
            :user-rate-multiplier="groupOption(option).userRate"
            :peak-rate-enabled="groupOption(option).group.peak_rate_enabled"
            :peak-start="groupOption(option).group.peak_start"
            :peak-end="groupOption(option).group.peak_end"
            :peak-rate-multiplier="groupOption(option).group.peak_rate_multiplier"
            :description="groupOption(option).group.description"
            :selected="selected"
          />
        </template>
      </Select>
    </div>

    <template v-else>
      <div>
        <label class="input-label">{{ t('smartRouting.strategy') }}</label>
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <button
            v-for="strategy in strategies"
            :key="strategy.value"
            type="button"
            :class="strategyClass(strategy.value)"
            @click="setStrategy(strategy.value)"
          >
            <span :class="['mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md', strategy.tone]">
              <Icon :name="strategy.icon" size="sm" />
            </span>
            <span class="min-w-0 text-left">
              <span class="block text-sm font-semibold text-ink-strong dark:text-white">{{ t(strategy.label) }}</span>
              <span class="mt-0.5 block text-xs leading-5 text-ink-muted dark:text-ink-muted">{{ t(strategy.hint) }}</span>
              <span class="mt-1 block text-[11px] leading-4 text-ink-muted dark:text-ink-muted">
                {{ t('smartRouting.weightSummary', { ...strategyWeights(strategy.value) }) }}
              </span>
            </span>
          </button>
        </div>
      </div>

      <div v-if="modelValue.strategy === 'custom'" class="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <label v-for="weight in weightFields" :key="weight.key" class="block">
          <span class="input-label">{{ t(weight.label) }}</span>
          <div class="relative">
            <input
              :value="modelValue.weights[weight.key]"
              type="number"
              min="0"
              max="100"
              step="1"
              class="input pr-8 tabular-nums"
              @input="setWeight(weight.key, $event)"
            />
            <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-ink-muted">%</span>
          </div>
        </label>
        <p class="sm:col-span-3 text-right text-xs font-medium" :class="weightTotal === 100 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500'">
          {{ t('smartRouting.weightTotal', { total: weightTotal }) }}
        </p>
      </div>

      <div>
        <div class="mb-2 flex items-center justify-between gap-3">
          <label class="input-label mb-0">{{ t('smartRouting.candidates') }}</label>
          <span class="text-xs tabular-nums text-ink-muted dark:text-ink-muted">
            {{ t('smartRouting.candidateCount', { count: modelValue.candidate_group_ids.length }) }}
          </span>
        </div>
        <div class="space-y-2">
          <label
            v-for="groupID in missingCandidateIDs"
            :key="`missing-${groupID}`"
            class="flex min-h-11 cursor-pointer items-center gap-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 dark:border-amber-900 dark:bg-amber-950/20"
          >
            <input
              type="checkbox"
              checked
              class="h-4 w-4 shrink-0 rounded border-line-strong text-primary-600 focus:ring-primary-500"
              @change="toggleCandidate(groupID)"
            />
            <span class="min-w-0 flex-1 text-sm text-amber-800 dark:text-amber-200">
              {{ t('smartRouting.unavailableCandidate', { id: groupID }) }}
            </span>
          </label>
          <label
            v-for="group in candidateGroups"
            :key="group.id"
            :class="candidateClass(group)"
          >
            <input
              type="checkbox"
              class="h-4 w-4 shrink-0 rounded border-line-strong text-primary-600 focus:ring-primary-500"
              :checked="isSelected(group.id)"
              :disabled="isDisabled(group)"
              @change="toggleCandidate(group.id)"
            />
            <span class="min-w-0 flex-1">
              <GroupBadge
                :name="group.name"
                :platform="group.platform"
                :subscription-type="group.subscription_type"
                :rate-multiplier="group.rate_multiplier"
                :user-rate-multiplier="userGroupRates[group.id] ?? null"
                :peak-rate-enabled="group.peak_rate_enabled"
                :peak-start="group.peak_start"
                :peak-end="group.peak_end"
                :peak-rate-multiplier="group.peak_rate_multiplier"
              />
              <span v-if="disabledReason(group)" class="mt-1 block text-xs text-ink-muted">
                {{ disabledReason(group) }}
              </span>
            </span>
          </label>
          <p v-if="candidateGroups.length === 0 && missingCandidateIDs.length === 0" class="py-3 text-center text-sm text-ink-muted">
            {{ t('smartRouting.noCompatibleGroups') }}
          </p>
        </div>
      </div>

      <div class="rounded-lg border border-emerald-200 bg-emerald-50/60 p-4 dark:border-emerald-800 dark:bg-emerald-950/20">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="text-sm font-semibold text-ink-strong dark:text-white">{{ t('smartRouting.rateGuard') }}</h3>
            <p class="mt-1 text-xs leading-5 text-ink-muted dark:text-ink-muted">{{ t('smartRouting.rateGuardHint') }}</p>
          </div>
          <button
            type="button"
            role="switch"
            :aria-checked="modelValue.rate_guard.enabled"
            :class="[
              'relative mt-0.5 inline-flex h-6 w-11 shrink-0 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
              modelValue.rate_guard.enabled ? 'bg-emerald-600' : 'bg-line-strong dark:bg-line-strong'
            ]"
            @click="setGuardEnabled(!modelValue.rate_guard.enabled)"
          >
            <span :class="['mt-0.5 h-5 w-5 rounded-full bg-white shadow-sm transition-transform', modelValue.rate_guard.enabled ? 'translate-x-5' : 'translate-x-0.5']" />
          </button>
        </div>
        <div v-if="modelValue.rate_guard.enabled" class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <label>
            <span class="input-label">{{ t('smartRouting.maxTextRate') }}</span>
            <input
              :value="modelValue.rate_guard.max_rate_multiplier ?? ''"
              type="number"
              min="0"
              step="0.0001"
              class="input tabular-nums"
              :placeholder="t('smartRouting.noLimit')"
              @input="setRateLimit('max_rate_multiplier', $event)"
            />
          </label>
          <label>
            <span class="input-label">{{ t('smartRouting.maxImageRate') }}</span>
            <input
              :value="modelValue.rate_guard.max_image_rate_multiplier ?? ''"
              type="number"
              min="0"
              step="0.0001"
              class="input tabular-nums"
              :placeholder="t('smartRouting.noLimit')"
              @input="setRateLimit('max_image_rate_multiplier', $event)"
            />
          </label>
        </div>
      </div>

      <div class="flex items-start gap-2 rounded-lg border border-cyan-200 bg-cyan-50 px-3 py-2.5 text-xs leading-5 text-cyan-900 dark:border-cyan-900 dark:bg-cyan-950/25 dark:text-cyan-200">
        <Icon name="clock" size="sm" class="mt-0.5 shrink-0" />
        <span>{{ t('smartRouting.snapshotHint') }}</span>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { smartRoutePresetWeights } from '@/composables/useSmartRouteForm'
import type { Group, SelectOption } from '@/types'
import type { SmartRouteFormState, SmartRouteMode, SmartRouteStrategy, SmartRouteWeights } from '@/types/smart-routing'

const props = defineProps<{
  modelValue: SmartRouteFormState
  groups: Group[]
  userGroupRates: Record<number, number>
  enabled: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: SmartRouteFormState]
}>()

const { t } = useI18n()

type IconName = InstanceType<typeof Icon>['$props']['name']
type WeightKey = keyof SmartRouteWeights
type GroupSelectOption = SelectOption & { group: Group; userRate: number | null }

const strategies: Array<{ value: SmartRouteStrategy; label: string; hint: string; icon: IconName; tone: string }> = [
  { value: 'auto', label: 'smartRouting.auto', hint: 'smartRouting.autoHint', icon: 'sparkles', tone: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300' },
  { value: 'price', label: 'smartRouting.price', hint: 'smartRouting.priceHint', icon: 'dollar', tone: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300' },
  { value: 'speed', label: 'smartRouting.speed', hint: 'smartRouting.speedHint', icon: 'bolt', tone: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300' },
  { value: 'success', label: 'smartRouting.success', hint: 'smartRouting.successHint', icon: 'shield', tone: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300' },
  { value: 'custom', label: 'smartRouting.custom', hint: 'smartRouting.customHint', icon: 'chart', tone: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300' }
]

const weightFields: Array<{ key: WeightKey; label: string }> = [
  { key: 'price', label: 'smartRouting.weightPrice' },
  { key: 'speed', label: 'smartRouting.weightSpeed' },
  { key: 'success', label: 'smartRouting.weightSuccess' }
]

const groupOptions = computed<GroupSelectOption[]>(() => props.groups
  .map((group) => ({
    value: group.id,
    label: group.name,
    group,
    userRate: props.userGroupRates[group.id] ?? null
  })))

const candidateGroups = computed(() => {
  const selected = new Set(props.modelValue.candidate_group_ids)
  return props.groups.filter((group) => group.status === 'active' || selected.has(group.id))
})
const missingCandidateIDs = computed(() => {
  const available = new Set(props.groups.map((group) => group.id))
  return props.modelValue.candidate_group_ids.filter((id) => !available.has(id))
})
const selectedAnchor = computed(() => props.groups.find((group) => group.id === props.modelValue.candidate_group_ids[0]))
const weightTotal = computed(() => props.modelValue.weights.price + props.modelValue.weights.speed + props.modelValue.weights.success)

const strategyWeights = (strategy: SmartRouteStrategy): SmartRouteWeights => (
  strategy === 'custom' ? props.modelValue.weights : smartRoutePresetWeights(strategy)
)

const clone = (): SmartRouteFormState => ({
  ...props.modelValue,
  candidate_group_ids: [...props.modelValue.candidate_group_ids],
  weights: { ...props.modelValue.weights },
  rate_guard: { ...props.modelValue.rate_guard }
})

const groupOption = (option: Record<string, unknown>) => option as unknown as GroupSelectOption
const isSelected = (id: number) => props.modelValue.candidate_group_ids.includes(id)

const disabledReason = (group: Group): string => {
  if (group.platform === 'composite') return t('smartRouting.compositeUnsupported')
  if (group.status !== 'active') return t('smartRouting.groupInactive')
  const anchor = selectedAnchor.value
  if (anchor && !isSelected(group.id) && (group.platform !== anchor.platform || group.subscription_type !== anchor.subscription_type)) {
    return t('smartRouting.incompatible')
  }
  return ''
}

const isDisabled = (group: Group) => {
  if (isSelected(group.id)) return false
  return Boolean(disabledReason(group)) || props.modelValue.candidate_group_ids.length >= 20
}

const modeButtonClass = (active: boolean) => [
  'flex min-h-9 items-center justify-center gap-2 rounded-md px-3 text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:cursor-not-allowed disabled:opacity-45',
  active ? 'bg-white text-primary-700 shadow-sm dark:bg-line-strong dark:text-primary-300' : 'text-ink hover:text-ink-strong dark:text-ink-muted dark:hover:text-white'
]

const strategyClass = (strategy: SmartRouteStrategy) => [
  'flex min-h-[76px] items-start gap-3 rounded-lg border p-3 text-left transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500',
  props.modelValue.strategy === strategy
    ? 'border-emerald-400 bg-emerald-50/70 dark:border-emerald-700 dark:bg-emerald-950/20'
    : 'border-line bg-white hover:border-line-strong dark:border-line-strong dark:bg-surface dark:hover:border-line-strong',
  strategy === 'custom' ? 'sm:col-span-2' : ''
]

const candidateClass = (group: Group) => [
  'flex min-h-11 items-center gap-3 rounded-lg border px-3 py-2 transition-colors',
  isSelected(group.id)
    ? 'border-emerald-300 bg-emerald-50/60 dark:border-emerald-800 dark:bg-emerald-950/20'
    : 'border-line bg-white dark:border-line-strong dark:bg-surface',
  isDisabled(group) ? 'cursor-not-allowed opacity-55' : 'cursor-pointer hover:border-line-strong dark:hover:border-line-strong'
]

const setMode = (mode: SmartRouteMode) => {
  if (mode === 'smart' && !props.enabled) return
  const next = clone()
  next.mode = mode
  if (mode === 'smart') next.group_id = null
  emit('update:modelValue', next)
}

const setGroup = (value: string | number | boolean | null) => {
  const next = clone()
  next.group_id = typeof value === 'number' ? value : null
  emit('update:modelValue', next)
}

const setStrategy = (strategy: SmartRouteStrategy) => {
  const next = clone()
  next.strategy = strategy
  if (strategy !== 'custom') next.weights = smartRoutePresetWeights(strategy)
  emit('update:modelValue', next)
}

const toggleCandidate = (id: number) => {
  const next = clone()
  next.candidate_group_ids = isSelected(id)
    ? next.candidate_group_ids.filter((candidateID) => candidateID !== id)
    : [...next.candidate_group_ids, id]
  emit('update:modelValue', next)
}

const setWeight = (key: WeightKey, event: Event) => {
  const next = clone()
  next.weights[key] = Number((event.target as HTMLInputElement).value)
  emit('update:modelValue', next)
}

const setGuardEnabled = (enabled: boolean) => {
  const next = clone()
  next.rate_guard.enabled = enabled
  emit('update:modelValue', next)
}

const setRateLimit = (key: 'max_rate_multiplier' | 'max_image_rate_multiplier', event: Event) => {
  const next = clone()
  const raw = (event.target as HTMLInputElement).value.trim()
  next.rate_guard[key] = raw === '' ? null : Number(raw)
  emit('update:modelValue', next)
}
</script>
