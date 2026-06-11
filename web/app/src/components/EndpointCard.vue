<template>
  <Card class="endpoint h-full flex flex-col transition hover:shadow-lg hover:scale-[1.01] dark:hover:border-gray-700">
    <CardHeader class="endpoint-header px-3 sm:px-6 pt-3 sm:pt-6 pb-2 space-y-0">
      <div class="flex items-start justify-between gap-2 sm:gap-3">
        <div class="flex-1 min-w-0 overflow-hidden">
          <CardTitle class="text-base sm:text-lg truncate">
            <span 
              class="hover:text-primary cursor-pointer hover:underline text-sm sm:text-base block truncate" 
              @click="navigateToDetails" 
              @keydown.enter="navigateToDetails"
              :title="endpoint.name"
              role="link"
              tabindex="0"
              :aria-label="`View details for ${endpoint.name}`">
              {{ endpoint.name }}
            </span>
          </CardTitle>
          <div class="flex items-center gap-2 text-xs sm:text-sm text-muted-foreground min-h-[1.25rem]">
            <span v-if="endpoint.group" class="truncate" :title="endpoint.group">{{ endpoint.group }}</span>
            <span v-if="endpoint.group && hostname">•</span>
            <span v-if="hostname" class="truncate" :title="hostname">{{ hostname }}</span>
          </div>
        </div>
        <div class="flex-shrink-0 ml-2 flex flex-col items-end gap-1">
          <StatusBadge :status="currentStatus" />
          <span
            v-if="maintenanceBadgeLabel"
            :class="[
              'inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full border',
              activeMaintenance
                ? 'bg-orange-100 text-orange-700 border-orange-300 dark:bg-orange-900/30 dark:text-orange-400 dark:border-orange-700'
                : 'bg-blue-100 text-blue-700 border-blue-300 dark:bg-blue-900/30 dark:text-blue-400 dark:border-blue-700'
            ]"
            :title="activeMaintenance ? activeMaintenance.description : upcomingMaintenance.description"
          >
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
            {{ maintenanceBadgeLabel }}
          </span>
        </div>
      </div>
    </CardHeader>
    <CardContent class="endpoint-content flex-1 pb-3 sm:pb-4 px-3 sm:px-6 pt-2">
      <div class="space-y-2">
        <div>
          <div class="flex items-center justify-between mb-1">
            <div class="flex-1"></div>
            <p class="text-xs text-muted-foreground" :title="showAverageResponseTime ? 'Average response time' : 'Minimum and maximum response time'">{{ formattedResponseTime }}</p>
          </div>
          <div class="flex gap-0.5">
            <div
              v-for="(result, index) in displayResults"
              :key="index"
              :class="[
                'flex-1 h-6 sm:h-8 rounded-sm transition-all',
                result ? 'cursor-pointer' : '',
                result ? (
                  result.success
                    ? (selectedResultIndex === index ? 'bg-green-700' : 'bg-green-500 hover:bg-green-700')
                    : (selectedResultIndex === index ? 'bg-red-700' : 'bg-red-500 hover:bg-red-700')
                ) : 'bg-gray-200 dark:bg-gray-700'
              ]"
              @mouseenter="result && handleMouseEnter(result, $event)"
              @mouseleave="result && handleMouseLeave(result, $event)"
              @click.stop="result && handleClick(result, $event, index)"
            />
          </div>
          <div class="flex items-center justify-between text-xs text-muted-foreground mt-1">
            <span>{{ oldestResultTime }}</span>
            <span>{{ newestResultTime }}</span>
          </div>
        </div>

        <!-- Uptime badges + SSL indicator -->
        <div class="flex items-center justify-between pt-1 border-t border-border/50">
          <div class="flex items-center gap-1.5">
            <img
              :src="`/api/v1/endpoints/${endpoint.key}/uptimes/7d/badge.svg`"
              alt="Uptime 7d"
              class="h-5"
              loading="lazy"
            />
            <img
              :src="`/api/v1/endpoints/${endpoint.key}/uptimes/30d/badge.svg`"
              alt="Uptime 30d"
              class="h-5"
              loading="lazy"
            />
          </div>
          <span
            v-if="sslStatus !== null"
            :class="['text-xs font-medium flex items-center gap-1', sslStatus ? 'text-green-600' : 'text-red-500']"
            title="SSL certificate status"
          >
            <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
            </svg>
            SSL
          </span>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { computed, ref, inject, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import StatusBadge from '@/components/StatusBadge.vue'
import { generatePrettyTimeAgo } from '@/utils/time'

const router = useRouter()
const maintenanceEvents = inject('maintenanceEvents', ref({}))

const props = defineProps({
  endpoint: {
    type: Object,
    required: true
  },
  maxResults: {
    type: Number,
    default: 50
  },
  showAverageResponseTime: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['showTooltip'])

// Track selected data point
const selectedResultIndex = ref(null)

const latestResult = computed(() => {
  if (!props.endpoint.results || props.endpoint.results.length === 0) {
    return null
  }
  return props.endpoint.results[props.endpoint.results.length - 1]
})

const currentStatus = computed(() => {
  if (!latestResult.value) return 'unknown'
  return latestResult.value.success ? 'healthy' : 'unhealthy'
})

const hostname = computed(() => {
  return latestResult.value?.hostname || null
})

// Returns true if SSL condition passed, false if failed, null if not evaluated
const sslStatus = computed(() => {
  if (!latestResult.value?.conditionResults) return null
  const sslCondition = latestResult.value.conditionResults.find(
    c => c.condition && c.condition.includes('CERTIFICATE_EXPIRATION')
  )
  if (!sslCondition) return null
  return sslCondition.success
})

const displayResults = computed(() => {
  const results = [...(props.endpoint.results || [])]
  while (results.length < props.maxResults) {
    results.unshift(null)
  }
  return results.slice(-props.maxResults)
})

const formattedResponseTime = computed(() => {
  if (!props.endpoint.results || props.endpoint.results.length === 0) {
    return 'N/A'
  }
  
  let total = 0
  let count = 0
  let min = Infinity
  let max = 0
  
  for (const result of props.endpoint.results) {
    if (result.duration) {
      const durationMs = result.duration / 1000000
      total += durationMs
      count++
      min = Math.min(min, durationMs)
      max = Math.max(max, durationMs)
    }
  }
  
  if (count === 0) return 'N/A'
  
  if (props.showAverageResponseTime) {
    const avgMs = Math.round(total / count)
    return `~${avgMs}ms`
  } else {
    // Show min-max range
    const minMs = Math.trunc(min)
    const maxMs = Math.trunc(max)
    // If min and max are the same, show single value
    if (minMs === maxMs) {
      return `${minMs}ms`
    }
    return `${minMs}-${maxMs}ms`
  }
})

const oldestResultTime = computed(() => {
  if (!props.endpoint.results || props.endpoint.results.length === 0) return ''
  const oldestResultIndex = Math.max(0, props.endpoint.results.length - props.maxResults)
  return generatePrettyTimeAgo(props.endpoint.results[oldestResultIndex].timestamp)
})

const newestResultTime = computed(() => {
  if (!props.endpoint.results || props.endpoint.results.length === 0) return ''
  return generatePrettyTimeAgo(props.endpoint.results[props.endpoint.results.length - 1].timestamp)
})

// Returns the active maintenance event for this endpoint, or null.
const activeMaintenance = computed(() => {
  const events = maintenanceEvents.value[props.endpoint.key] || []
  const now = Date.now()
  return events.find(e => new Date(e.start).getTime() <= now && new Date(e.end).getTime() > now) || null
})

// Returns the next upcoming maintenance event within 24 h, or null.
const upcomingMaintenance = computed(() => {
  if (activeMaintenance.value) return null
  const events = maintenanceEvents.value[props.endpoint.key] || []
  const now = Date.now()
  const in24h = now + 24 * 60 * 60 * 1000
  return events.find(e => new Date(e.start).getTime() > now && new Date(e.start).getTime() <= in24h) || null
})

const maintenanceBadgeLabel = computed(() => {
  if (activeMaintenance.value) {
    const end = new Date(activeMaintenance.value.end)
    return `Manutenzione fino alle ${end.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
  }
  if (upcomingMaintenance.value) {
    const start = new Date(upcomingMaintenance.value.start)
    return `Manutenzione alle ${start.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
  }
  return null
})

const navigateToDetails = () => {
  router.push(`/endpoints/${props.endpoint.key}`)
}

const handleMouseEnter = (result, event) => {
  emit('showTooltip', result, event, 'hover')
}

const handleMouseLeave = (result, event) => {
  emit('showTooltip', null, event, 'hover')
}

const handleClick = (result, event, index) => {
  // Clear selections in other cards first
  window.dispatchEvent(new CustomEvent('clear-data-point-selection'))
  // Then toggle this card's selection
  if (selectedResultIndex.value === index) {
    selectedResultIndex.value = null
    emit('showTooltip', null, event, 'click')
  } else {
    selectedResultIndex.value = index
    emit('showTooltip', result, event, 'click')
  }
}

// Listen for clear selection event
const handleClearSelection = () => {
  selectedResultIndex.value = null
}

onMounted(() => {
  window.addEventListener('clear-data-point-selection', handleClearSelection)
})

onUnmounted(() => {
  window.removeEventListener('clear-data-point-selection', handleClearSelection)
})
</script>