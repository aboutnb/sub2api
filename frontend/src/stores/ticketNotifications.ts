import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminTicketsAPI } from '@/api/admin/tickets'

export const useTicketNotificationsStore = defineStore('ticketNotifications', () => {
 const pendingCount = ref(0)
 const totalCount = ref(0)
 let owner: number | null = null
 let initialized = false
 let controller: AbortController | null = null
 function setOwner(id: number | null) {
  if (id === owner) return
  controller?.abort(); controller = null
  owner = id; initialized = false; pendingCount.value = 0; totalCount.value = 0
 }
 async function refresh(): Promise<{ initial: boolean; added: number } | null> {
  if (owner === null || controller) return null
  const request = new AbortController(); controller = request
  try {
   const result = await adminTicketsAPI.summary({ signal: request.signal })
   if (request.signal.aborted) return null
   const initial = !initialized
   const added = initialized ? Math.max(0, result.total_count - totalCount.value) : 0
   pendingCount.value = result.pending_count; totalCount.value = result.total_count; initialized = true
   return { initial, added }
  } catch { return null } finally { if (controller === request) controller = null }
 }
 function cancel() { controller?.abort(); controller = null }
 return { pendingCount, totalCount, setOwner, refresh, cancel }
})
