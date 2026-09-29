export type HelpSelectorValue = string | undefined | null

export interface HelpSelection {
  platform?: HelpSelectorValue
  client?: HelpSelectorValue
  system?: HelpSelectorValue
  shell?: HelpSelectorValue
  protocol?: HelpSelectorValue
  version?: HelpSelectorValue
}

export type HelpSelectorSchema = Record<string, unknown>

/** Returns false only when a document explicitly excludes the current context. */
export function matchesHelpSelectors(schema: HelpSelectorSchema | undefined, selection: HelpSelection): boolean {
  if (!schema) return true
  const values: Record<string, HelpSelectorValue> = {
    platforms: selection.platform,
    clients: selection.client,
    systems: selection.system,
    shells: selection.shell,
    protocols: selection.protocol,
    versions: selection.version
  }
  return Object.entries(values).every(([key, selected]) => {
    const allowed = schema[key]
    if (!Array.isArray(allowed) || allowed.length === 0 || !selected) return true
    return allowed.some(value => typeof value === 'string' && value === selected)
  })
}
