import CherryStudioGuide from './CherryStudioGuide.vue'
import WorkBuddyGuide from './WorkBuddyGuide.vue'
import ZCodeGuide from './ZCodeGuide.vue'
import CursorGuide from './CursorGuide.vue'
import TraeGuide from './TraeGuide.vue'
import KissTranslatorGuide from './KissTranslatorGuide.vue'
import ImmersiveTranslateGuide from './ImmersiveTranslateGuide.vue'
import DshGuide from './DshGuide.vue'
import ClineGuide from './ClineGuide.vue'
import SillyTavernGuide from './SillyTavernGuide.vue'
import TavernAIGuide from './TavernAIGuide.vue'
import type { Component } from 'vue'

export const clientGuideComponents: Record<string, Component> = {
  'cherry-studio': CherryStudioGuide, workbuddy: WorkBuddyGuide, zcode: ZCodeGuide, cursor: CursorGuide,
  trae: TraeGuide, 'kiss-translator': KissTranslatorGuide, 'immersive-translate': ImmersiveTranslateGuide, dsh: DshGuide,
  cline: ClineGuide, sillytavern: SillyTavernGuide, tavernai: TavernAIGuide
}
