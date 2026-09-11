# Aivoza visual coverage ledger

This ledger is the release gate for the second-pass UI refresh. `MASTER.md` is
the visual source of truth; this file records every route-level view and every
view-owned workflow that must inherit it. A file is not considered covered by
appearance alone: its loading, empty, error, success, warning, disabled, long
content, overlay, light, dark, 375px, 768px, 1024px, and 1440px states must all
remain usable.

## Verification matrix

| Dimension | Required variants |
| --- | --- |
| Theme | Light, dark, initial system preference, persisted explicit preference |
| Viewport | 375px, 768px, 1024px, 1440px |
| Identity | Guest, user, administrator |
| Product mode | Standard, simple, backend mode, feature disabled |
| Data state | Loading, populated, empty, recoverable error, long values, wide data |
| Interaction | Keyboard, pointer, dialog, dropdown/Teleport, nested flow, reduced motion |

## Wave 1 — public and setup

- `frontend/src/views/HomeView.vue`
- `frontend/src/views/KeyUsageView.vue`
- `frontend/src/views/ModelPlazaView.vue`
- `frontend/src/views/NotFoundView.vue`
- `frontend/src/views/public/LegalDocumentView.vue`
- `frontend/src/views/setup/SetupWizardView.vue`

## Wave 2 — authentication and callbacks

- `frontend/src/views/auth/DingTalkCallbackView.vue`
- `frontend/src/views/auth/DingTalkEmailCompletionView.vue`
- `frontend/src/views/auth/EmailVerifyView.vue`
- `frontend/src/views/auth/ForgotPasswordView.vue`
- `frontend/src/views/auth/LinuxDoCallbackView.vue`
- `frontend/src/views/auth/LoginView.vue`
- `frontend/src/views/auth/OAuthCallbackView.vue`
- `frontend/src/views/auth/OidcCallbackView.vue`
- `frontend/src/views/auth/RegisterView.vue`
- `frontend/src/views/auth/ResetPasswordView.vue`
- `frontend/src/views/auth/WechatCallbackView.vue`
- `frontend/src/views/auth/WechatPaymentCallbackView.vue`

## Wave 3 — user core

- `frontend/src/views/user/AvailableChannelsView.vue`
- `frontend/src/views/user/BatchImageGuideView.vue`
- `frontend/src/views/user/ChannelStatusV1View.vue`
- `frontend/src/views/user/ChannelStatusV2View.vue`
- `frontend/src/views/user/ChannelStatusV3View.vue`
- `frontend/src/views/user/ChannelStatusView.vue`
- `frontend/src/views/user/DashboardView.vue`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/views/user/UsageView.vue`

## Wave 4 — user commerce and account

- `frontend/src/views/user/AffiliateView.vue`
- `frontend/src/views/user/AirwallexPaymentView.vue`
- `frontend/src/views/user/CheckinView.vue`
- `frontend/src/views/user/CustomPageView.vue`
- `frontend/src/views/user/PaymentQRCodeView.vue`
- `frontend/src/views/user/PaymentResultView.vue`
- `frontend/src/views/user/PaymentView.vue`
- `frontend/src/views/user/ProfileView.vue`
- `frontend/src/views/user/RedeemView.vue`
- `frontend/src/views/user/StripePaymentView.vue`
- `frontend/src/views/user/StripePopupView.vue`
- `frontend/src/views/user/SubscriptionsView.vue`
- `frontend/src/views/user/UserOrdersView.vue`

## Wave 5 — administrative resources

- `frontend/src/views/admin/AccountsView.vue`
- `frontend/src/views/admin/ChannelMonitorView.vue`
- `frontend/src/views/admin/ChannelsView.vue`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/ProxiesView.vue`
- `frontend/src/views/admin/SubscriptionsView.vue`
- `frontend/src/views/admin/UsersView.vue`

## Wave 6 — operations, billing, and engagement

- `frontend/src/views/admin/AnnouncementsView.vue`
- `frontend/src/views/admin/CheckinView.vue`
- `frontend/src/views/admin/DashboardView.vue`
- `frontend/src/views/admin/EmailBroadcastsView.vue`
- `frontend/src/views/admin/PromoCodesView.vue`
- `frontend/src/views/admin/RedeemView.vue`
- `frontend/src/views/admin/UsageView.vue`
- `frontend/src/views/admin/affiliates/AdminAffiliateInvitesView.vue`
- `frontend/src/views/admin/affiliates/AdminAffiliateRebatesView.vue`
- `frontend/src/views/admin/affiliates/AdminAffiliateRecordsTable.vue`
- `frontend/src/views/admin/affiliates/AdminAffiliateTransfersView.vue`
- `frontend/src/views/admin/orders/AdminOrdersView.vue`
- `frontend/src/views/admin/orders/AdminPaymentDashboardView.vue`
- `frontend/src/views/admin/orders/AdminPaymentPlansView.vue`
- `frontend/src/views/admin/orders/PlanEditDialog.vue`

## Wave 7 — security, settings, and recovery

- `frontend/src/views/admin/AuditLogView.vue`
- `frontend/src/views/admin/AuthIPBanView.vue`
- `frontend/src/views/admin/RegistrationProtectionView.vue`
- `frontend/src/views/admin/BackupView.vue`
- `frontend/src/views/admin/PluginsView.vue`
- `frontend/src/views/admin/RiskControlView.vue`
- `frontend/src/views/admin/SettingsView.vue`
- `frontend/src/views/admin/settings/EmailTemplateEditor.vue`
- `frontend/src/views/admin/settings/OpenAIFastPolicyUserSelector.vue`
- `frontend/src/features/prompt-audit/PromptAuditView.vue`

## Operations workspace owned by the route-level views

- `frontend/src/views/admin/ops/OpsDashboard.vue`
- `frontend/src/views/admin/ops/components/OpsAlertEventsCard.vue`
- `frontend/src/views/admin/ops/components/OpsAlertRulesCard.vue`
- `frontend/src/views/admin/ops/components/OpsConcurrencyCard.vue`
- `frontend/src/views/admin/ops/components/OpsDashboardHeader.vue`
- `frontend/src/views/admin/ops/components/OpsDashboardSkeleton.vue`
- `frontend/src/views/admin/ops/components/OpsEmailNotificationCard.vue`
- `frontend/src/views/admin/ops/components/OpsErrorDetailModal.vue`
- `frontend/src/views/admin/ops/components/OpsErrorDetailsModal.vue`
- `frontend/src/views/admin/ops/components/OpsErrorDistributionChart.vue`
- `frontend/src/views/admin/ops/components/OpsErrorLogTable.vue`
- `frontend/src/views/admin/ops/components/OpsErrorTrendChart.vue`
- `frontend/src/views/admin/ops/components/OpsLatencyChart.vue`
- `frontend/src/views/admin/ops/components/OpsOpenAITokenStatsCard.vue`
- `frontend/src/views/admin/ops/components/OpsRequestDetailsModal.vue`
- `frontend/src/views/admin/ops/components/OpsRuntimeSettingsCard.vue`
- `frontend/src/views/admin/ops/components/OpsSettingsDialog.vue`
- `frontend/src/views/admin/ops/components/OpsSwitchRateTrendChart.vue`
- `frontend/src/views/admin/ops/components/OpsSystemLogTable.vue`
- `frontend/src/views/admin/ops/components/OpsThroughputTrendChart.vue`

## Hidden feature surfaces

These feature-owned views and panels are reachable through feature flags,
dynamic routing, or parent views and are part of the same release gate.

- `frontend/src/features/channel-monitor-v2/FilterMultiSelect.vue`
- `frontend/src/features/channel-monitor-v2/MetricCell.vue`
- `frontend/src/features/channel-monitor-v2/MonitorRankBadge.vue`
- `frontend/src/features/channel-monitor-v2/MonitorSettingsPanel.vue`
- `frontend/src/features/channel-monitor-v2/MonitorTrendChart.vue`
- `frontend/src/features/channel-monitor-v2/RelayPulseMatrix.vue`
- `frontend/src/features/prompt-audit/PromptAuditView.vue`
- `frontend/src/features/prompt-audit/components/EndpointPool.vue`
- `frontend/src/features/prompt-audit/components/EventDetailDialog.vue`
- `frontend/src/features/prompt-audit/components/EventWorkspace.vue`
- `frontend/src/features/prompt-audit/components/FilterDeleteDialog.vue`
- `frontend/src/features/prompt-audit/components/PolicyPanel.vue`
- `frontend/src/features/prompt-audit/components/RuntimeOverview.vue`

## Shared overlay and Portal archive

The following shared components contain a Portal/Teleport, dialog semantics, or
compose the shared dialog contracts. Every item must pass focus containment,
focus return, Escape/top-stack behavior, scroll lock, nested overlay, busy
state, light/dark, and narrow viewport review. This exact inventory is enforced
by the coverage contract test so new overlays cannot ship unrecorded.

- `frontend/src/components/account/AccountGroupsCell.vue`
- `frontend/src/components/account/AccountStatsModal.vue`
- `frontend/src/components/account/AccountTestModal.vue`
- `frontend/src/components/account/BulkEditAccountModal.vue`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/OllamaCloudUsageSettings.vue`
- `frontend/src/components/account/OpenAIQuotaResetCell.vue`
- `frontend/src/components/account/ReAuthAccountModal.vue`
- `frontend/src/components/account/SyncFromCrsModal.vue`
- `frontend/src/components/account/TempUnschedStatusModal.vue`
- `frontend/src/components/admin/AdminComplianceDialog.vue`
- `frontend/src/components/admin/ErrorPassthroughRulesModal.vue`
- `frontend/src/components/admin/TLSFingerprintProfilesModal.vue`
- `frontend/src/components/admin/account/AccountActionMenu.vue`
- `frontend/src/components/admin/account/AccountStatsModal.vue`
- `frontend/src/components/admin/account/AccountTestModal.vue`
- `frontend/src/components/admin/account/BatchAccountTestModal.vue`
- `frontend/src/components/admin/account/ImportDataModal.vue`
- `frontend/src/components/admin/account/ReAuthAccountModal.vue`
- `frontend/src/components/admin/account/ScheduledTestsPanel.vue`
- `frontend/src/components/admin/announcements/AnnouncementReadStatusDialog.vue`
- `frontend/src/components/admin/group/GroupRPMOverridesModal.vue`
- `frontend/src/components/admin/group/GroupRateMultipliersModal.vue`
- `frontend/src/components/admin/monitor/MonitorFormDialog.vue`
- `frontend/src/components/admin/monitor/MonitorKeyPickerDialog.vue`
- `frontend/src/components/admin/monitor/MonitorRunResultDialog.vue`
- `frontend/src/components/admin/monitor/MonitorTemplateApplyPickerDialog.vue`
- `frontend/src/components/admin/monitor/MonitorTemplateManagerDialog.vue`
- `frontend/src/components/admin/payment/AdminOrderDetail.vue`
- `frontend/src/components/admin/payment/AdminRefundDialog.vue`
- `frontend/src/components/admin/proxy/ImportDataModal.vue`
- `frontend/src/components/admin/usage/UsageCleanupDialog.vue`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/user/BulkEditUserModal.vue`
- `frontend/src/components/admin/user/GroupReplaceModal.vue`
- `frontend/src/components/admin/user/UserAllowedGroupsModal.vue`
- `frontend/src/components/admin/user/UserApiKeysModal.vue`
- `frontend/src/components/admin/user/UserBalanceHistoryModal.vue`
- `frontend/src/components/admin/user/UserBalanceModal.vue`
- `frontend/src/components/admin/user/UserCreateModal.vue`
- `frontend/src/components/admin/user/UserEditModal.vue`
- `frontend/src/components/admin/user/UserPlatformQuotaModal.vue`
- `frontend/src/components/auth/LoginAgreementPrompt.vue`
- `frontend/src/components/channels/SupportedModelChip.vue`
- `frontend/src/components/checkin/CheckinShortcut.vue`
- `frontend/src/components/checkin/LuckyCheckinConfirmDialog.vue`
- `frontend/src/components/common/AnnouncementBell.vue`
- `frontend/src/components/common/AnnouncementPopup.vue`
- `frontend/src/components/common/BaseDialog.vue`
- `frontend/src/components/common/ConfirmDialog.vue`
- `frontend/src/components/common/ExportProgressDialog.vue`
- `frontend/src/components/common/HelpTooltip.vue`
- `frontend/src/components/common/ProxySelector.vue`
- `frontend/src/components/common/Select.vue`
- `frontend/src/components/common/DateRangePicker.vue`
- `frontend/src/components/common/Toast.vue`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/components/layout/AppSidebar.vue`
- `frontend/src/components/layout/SupportContact.vue`
- `frontend/src/components/admin/SupportContactsEditor.vue`
- `frontend/src/components/payment/PaymentProviderDialog.vue`
- `frontend/src/components/payment/PaymentQRDialog.vue`
- `frontend/src/components/user/MonitorDetailDialog.vue`
- `frontend/src/components/user/UserAttributesConfigModal.vue`
- `frontend/src/components/user/UserErrorDetailModal.vue`
- `frontend/src/components/user/monitor/ChannelMonitorV3Timeline.vue`
- `frontend/src/features/channel-monitor-v2/FilterMultiSelect.vue`
- `frontend/src/features/channel-monitor-v2/RelayPulseMatrix.vue`
- `frontend/src/features/prompt-audit/PromptAuditView.vue`
- `frontend/src/features/prompt-audit/components/EndpointPool.vue`
- `frontend/src/features/prompt-audit/components/EventDetailDialog.vue`
- `frontend/src/features/prompt-audit/components/FilterDeleteDialog.vue`

## Release evidence

- 2026-09-05 inventory gate: the coverage contract archived all 90 current
  `frontend/src/views/**/*.vue` files and all 70 shared overlay/Portal owners.
- Automated verification: 287 Vitest files / 2020 tests, `vue-tsc`, ESLint,
  production Vite build, `go test ./...`, and both root/deploy Docker image
  builds passed. The final images are tagged
  `sub2api-local:aivoza-system-20260905` and
  `sub2api-local:aivoza-system-deploy-20260905`.
- Live browser gate at `http://127.0.0.1:18081`: 70 routes x 2 themes x 4
  breakpoints (560 combinations) completed with zero navigation, runtime,
  theme, whole-page overflow, header clipping, sidebar-brand clipping, or
  poster-heading findings. The final sidebar geometry preserves a 44px brand
  mark and interaction target around the visually compact 28px version chip.
  A read-only mocked ordinary-user identity pass also
  covered light/dark at 375px and 1440px without database writes.
- Visual archive: 20 final screenshots cover the fixed homepage and
  authenticated administrator shell in light/dark at 375px, 768px, 1024px,
  and 1440px, plus the read-only ordinary-user shell at 375px and 1440px,
  under `output/playwright/aivoza-system/`.
- Runtime contract: the public settings response neutralizes both legacy home
  fields, configured branding remains active, nonce-bearing HTML is served
  `no-store` without an ETag, and each response receives a matching fresh CSP
  nonce. Existing PostgreSQL and Redis container IDs and named volumes remained
  unchanged throughout candidate validation and cutover.
