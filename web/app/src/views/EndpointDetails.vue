<template>
  <div class="dashboard-container bg-background">
    <div class="container mx-auto px-4 py-8 max-w-7xl">
      <div class="mb-6">
        <Button variant="ghost" class="mb-4" @click="goBack">
          <ArrowLeft class="h-4 w-4 mr-2" />
          Back to Dashboard
        </Button>

        <div v-if="endpointStatus && endpointStatus.name" class="space-y-6">

          <!-- Header -->
          <div class="flex items-start justify-between">
            <div>
              <h1 class="text-4xl font-bold tracking-tight">{{ endpointStatus.name }}</h1>
              <div class="flex items-center gap-3 text-muted-foreground mt-2">
                <span v-if="endpointStatus.group">Group: {{ endpointStatus.group }}</span>
                <span v-if="endpointStatus.group && hostname">•</span>
                <span v-if="hostname">{{ hostname }}</span>
              </div>
            </div>
            <StatusBadge :status="currentHealthStatus" />
          </div>

          <!-- Summary cards -->
          <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Current Status</CardTitle>
              </CardHeader>
              <CardContent>
                <div class="text-2xl font-bold">{{ currentHealthStatus === 'healthy' ? 'Operational' : 'Issues Detected' }}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Avg Response Time</CardTitle>
              </CardHeader>
              <CardContent>
                <div class="text-2xl font-bold">{{ pageAverageResponseTime }}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Response Time Range</CardTitle>
              </CardHeader>
              <CardContent>
                <div class="text-2xl font-bold">{{ pageResponseTimeRange }}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardTitle class="text-sm font-medium text-muted-foreground">Last Check</CardTitle>
              </CardHeader>
              <CardContent>
                <div class="text-2xl font-bold">{{ lastCheckTime }}</div>
              </CardContent>
            </Card>
          </div>

          <!-- ── UPTIME OVERVIEW ─────────────────────────────────────────── -->
          <Card>
            <CardHeader>
              <CardTitle>Uptime Overview</CardTitle>
            </CardHeader>
            <CardContent class="space-y-6">

              <!-- Uptime % numbers -->
              <div class="grid grid-cols-3 gap-4">
                <div v-for="period in ['24h', '7d', '30d']" :key="period"
                     class="flex flex-col items-center justify-center rounded-lg border p-4 gap-1">
                  <span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    {{ period === '24h' ? 'Last 24h' : period === '7d' ? 'Last 7 days' : 'Last 30 days' }}
                  </span>
                  <span v-if="uptimeStats[period] !== null"
                        :class="['text-3xl font-bold tabular-nums', uptimeColorClass(uptimeStats[period])]">
                    {{ uptimeStats[period].toFixed(2) }}%
                  </span>
                  <span v-else class="text-3xl font-bold text-muted-foreground">—</span>
                </div>
              </div>

              <!-- 30-day daily status bar -->
              <div class="space-y-2">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-sm font-medium">Daily Status — Last 30 Days</span>
                  <span class="text-xs text-muted-foreground">{{ dailyBarCoverage }}</span>
                </div>
                <div class="flex items-end gap-[3px] h-10">
                  <div
                    v-for="seg in dailyStatusSegments"
                    :key="seg.date"
                    :class="['flex-1 rounded-sm transition-colors cursor-default', segmentColorClass(seg)]"
                    :title="segmentTooltip(seg)"
                  />
                </div>
                <div class="flex justify-between text-xs text-muted-foreground">
                  <span>30 days ago</span>
                  <span>Today</span>
                </div>
                <!-- Legend -->
                <div class="flex items-center gap-4 pt-1 text-xs text-muted-foreground">
                  <span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded-sm bg-green-500"></span> Operational ≥99%</span>
                  <span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded-sm bg-yellow-400"></span> Degraded 80–98%</span>
                  <span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded-sm bg-red-500"></span> Outage &lt;80%</span>
                  <span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded-sm bg-gray-200 dark:bg-gray-700"></span> No data</span>
                </div>
              </div>
            </CardContent>
          </Card>

          <!-- ── RESPONSE TIME ───────────────────────────────────────────── -->
          <Card v-if="showResponseTimeChartAndBadges">
            <CardHeader>
              <div class="flex items-center justify-between">
                <CardTitle>Response Time</CardTitle>
                <select
                  v-model="selectedChartDuration"
                  class="text-sm bg-background border rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="24h">Last 24 hours</option>
                  <option value="7d">Last 7 days</option>
                  <option value="30d">Last 30 days</option>
                </select>
              </div>
            </CardHeader>
            <CardContent>
              <ResponseTimeChart
                v-if="endpointStatus && endpointStatus.key"
                :endpointKey="endpointStatus.key"
                :duration="selectedChartDuration"
                :serverUrl="serverUrl"
                :events="endpointStatus.events || []"
              />
            </CardContent>
          </Card>

          <!-- ── RECENT CHECKS ───────────────────────────────────────────── -->
          <Card>
            <CardHeader>
              <div class="flex items-center justify-between">
                <CardTitle>Recent Checks</CardTitle>
                <div class="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    @click="toggleShowAverageResponseTime"
                    :title="showAverageResponseTime ? 'Show min-max response time' : 'Show average response time'"
                  >
                    <Activity v-if="showAverageResponseTime" class="h-5 w-5" />
                    <Timer v-else class="h-5 w-5" />
                  </Button>
                  <Button variant="ghost" size="icon" @click="fetchData" title="Refresh" :disabled="isRefreshing">
                    <RefreshCw :class="['h-4 w-4', isRefreshing && 'animate-spin']" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <EndpointCard
                v-if="endpointStatus"
                :endpoint="endpointStatus"
                :maxResults="resultPageSize"
                :showAverageResponseTime="showAverageResponseTime"
                @showTooltip="showTooltip"
                class="border-0 shadow-none bg-transparent p-0"
              />
              <div v-if="endpointStatus && endpointStatus.key" class="pt-4 border-t">
                <Pagination @page="changePage" :numberOfResultsPerPage="resultPageSize" :currentPageProp="currentPage" />
              </div>
            </CardContent>
          </Card>

          <!-- ── EVENTS ──────────────────────────────────────────────────── -->
          <Card v-if="events && events.length > 0">
            <CardHeader>
              <CardTitle>Events</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="space-y-4">
                <div v-for="event in events" :key="event.timestamp" class="flex items-start gap-4 pb-4 border-b last:border-0">
                  <div class="mt-1">
                    <ArrowUpCircle v-if="event.type === 'HEALTHY'" class="h-5 w-5 text-green-500" />
                    <ArrowDownCircle v-else-if="event.type === 'UNHEALTHY'" class="h-5 w-5 text-red-500" />
                    <PlayCircle v-else class="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div class="flex-1">
                    <p class="font-medium">{{ event.fancyText }}</p>
                    <p class="text-sm text-muted-foreground">{{ prettifyTimestamp(event.timestamp) }} • {{ event.fancyTimeAgo }}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

        </div>

        <div v-else class="flex items-center justify-center py-20">
          <Loading size="lg" />
        </div>
      </div>
    </div>

    <Settings @refreshData="fetchData" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft, RefreshCw, ArrowUpCircle, ArrowDownCircle, PlayCircle, Activity, Timer } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import StatusBadge from '@/components/StatusBadge.vue'
import EndpointCard from '@/components/EndpointCard.vue'
import Settings from '@/components/Settings.vue'
import Pagination from '@/components/Pagination.vue'
import Loading from '@/components/Loading.vue'
import ResponseTimeChart from '@/components/ResponseTimeChart.vue'
import { generatePrettyTimeAgo, generatePrettyTimeDifference } from '@/utils/time'

const router = useRouter()
const route = useRoute()
const emit = defineEmits(['showTooltip'])

const endpointStatus = ref(null)
const currentStatus = ref(null)
const events = ref([])
const currentPage = ref(1)
const resultPageSize = 50
const showResponseTimeChartAndBadges = ref(false)
const showAverageResponseTime = ref(localStorage.getItem('gatus:show-average-response-time') !== 'false')
const selectedChartDuration = ref('30d')
const isRefreshing = ref(false)
const serverUrl = ref('..')

// Uptime stats (24h / 7d / 30d as percentages 0–100)
const uptimeStats = ref({ '24h': null, '7d': null, '30d': null })

// All results for the 30-day daily bar (fetched once with large pageSize)
const allResults = ref([])

// ── Computed ─────────────────────────────────────────────────────────────────

const latestResult = computed(() => {
  if (!currentStatus.value?.results?.length) return null
  return currentStatus.value.results[currentStatus.value.results.length - 1]
})

const currentHealthStatus = computed(() => {
  if (!latestResult.value) return 'unknown'
  return latestResult.value.success ? 'healthy' : 'unhealthy'
})

const hostname = computed(() => latestResult.value?.hostname || null)

const pageAverageResponseTime = computed(() => {
  if (!endpointStatus.value?.results?.length) return 'N/A'
  let total = 0, count = 0
  for (const r of endpointStatus.value.results) {
    if (r.duration) { total += r.duration; count++ }
  }
  return count ? `${Math.round(total / count / 1000000)}ms` : 'N/A'
})

const pageResponseTimeRange = computed(() => {
  if (!endpointStatus.value?.results?.length) return 'N/A'
  let min = Infinity, max = 0, hasData = false
  for (const r of endpointStatus.value.results) {
    if (r.duration) { min = Math.min(min, r.duration); max = Math.max(max, r.duration); hasData = true }
  }
  if (!hasData) return 'N/A'
  const minMs = Math.trunc(min / 1000000)
  const maxMs = Math.trunc(max / 1000000)
  return minMs === maxMs ? `${minMs}ms` : `${minMs}-${maxMs}ms`
})

const lastCheckTime = computed(() => {
  if (!currentStatus.value?.results?.length) return 'Never'
  return generatePrettyTimeAgo(currentStatus.value.results[currentStatus.value.results.length - 1].timestamp)
})

// 30 segments (oldest → newest), each representing one day
const dailyStatusSegments = computed(() => {
  const byDay = {}
  for (const r of allResults.value) {
    const d = new Date(r.timestamp)
    const key = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}-${String(d.getUTCDate()).padStart(2, '0')}`
    if (!byDay[key]) byDay[key] = { total: 0, success: 0 }
    byDay[key].total++
    if (r.success) byDay[key].success++
  }

  const now = new Date()
  return Array.from({ length: 30 }, (_, i) => {
    const d = new Date(now)
    d.setUTCDate(d.getUTCDate() - (29 - i))
    const key = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}-${String(d.getUTCDate()).padStart(2, '0')}`
    const label = d.toLocaleDateString('it-IT', { month: 'short', day: 'numeric', timeZone: 'UTC' })
    const day = byDay[key]
    if (!day) return { date: label, uptime: null, checks: 0, hasData: false }
    return { date: label, uptime: day.success / day.total, checks: day.total, hasData: true }
  })
})

const dailyBarCoverage = computed(() => {
  const daysWithData = dailyStatusSegments.value.filter(s => s.hasData).length
  if (daysWithData === 0) return 'No historical data yet'
  if (daysWithData === 30) return '30 / 30 days covered'
  return `${daysWithData} / 30 days covered`
})

// ── Helpers ───────────────────────────────────────────────────────────────────

const uptimeColorClass = (pct) => {
  if (pct >= 99) return 'text-green-600'
  if (pct >= 95) return 'text-yellow-500'
  return 'text-red-500'
}

const segmentColorClass = (seg) => {
  if (!seg.hasData) return 'bg-gray-200 dark:bg-gray-700 h-6'
  if (seg.uptime >= 0.99) return 'bg-green-500 h-10'
  if (seg.uptime >= 0.80) return 'bg-yellow-400 h-10'
  return 'bg-red-500 h-10'
}

const segmentTooltip = (seg) => {
  if (!seg.hasData) return `${seg.date}: No data`
  const pct = (seg.uptime * 100).toFixed(1)
  return `${seg.date}: ${pct}% uptime (${seg.checks} checks)`
}

// ── Data fetching ─────────────────────────────────────────────────────────────

const fetchData = async () => {
  isRefreshing.value = true
  try {
    const response = await fetch(
      `/api/v1/endpoints/${route.params.key}/statuses?page=${currentPage.value}&pageSize=${resultPageSize}`,
      { credentials: 'include' }
    )
    if (response.status === 200) {
      const data = await response.json()
      endpointStatus.value = data
      if (currentPage.value === 1) currentStatus.value = data

      let processedEvents = []
      if (data.events?.length) {
        for (let i = data.events.length - 1; i >= 0; i--) {
          const event = data.events[i]
          const next = data.events[i + 1]
          if (i === data.events.length - 1) {
            event.fancyText = event.type === 'UNHEALTHY' ? 'Endpoint is unhealthy'
              : event.type === 'HEALTHY' ? 'Endpoint is healthy' : 'Monitoring started'
          } else {
            if (event.type === 'HEALTHY') event.fancyText = 'Endpoint became healthy'
            else if (event.type === 'UNHEALTHY') {
              event.fancyText = next
                ? 'Endpoint was unhealthy for ' + generatePrettyTimeDifference(next.timestamp, event.timestamp)
                : 'Endpoint became unhealthy'
            } else event.fancyText = 'Monitoring started'
          }
          event.fancyTimeAgo = generatePrettyTimeAgo(event.timestamp)
          processedEvents.push(event)
        }
      }
      events.value = processedEvents

      if (data.results?.some(r => r.duration > 0)) showResponseTimeChartAndBadges.value = true
    }
  } catch (err) {
    console.error('[Details][fetchData]', err)
  } finally {
    isRefreshing.value = false
  }
}

// Fetch raw uptime % for 24h / 7d / 30d
const fetchUptimeStats = async () => {
  await Promise.all(['24h', '7d', '30d'].map(async (period) => {
    try {
      const r = await fetch(`/api/v1/endpoints/${route.params.key}/uptimes/${period}`, { credentials: 'include' })
      if (r.ok) uptimeStats.value[period] = parseFloat(await r.text()) * 100
    } catch (e) { console.warn('uptime fetch failed', e) }
  }))
}

// Fetch all stored results (up to maximum-number-of-results) for the daily bar
const fetchAllResults = async () => {
  try {
    const r = await fetch(
      `/api/v1/endpoints/${route.params.key}/statuses?page=1&pageSize=43200`,
      { credentials: 'include' }
    )
    if (r.ok) {
      const data = await r.json()
      allResults.value = data.results || []
    }
  } catch (e) { console.warn('results fetch failed', e) }
}

// ── Actions ───────────────────────────────────────────────────────────────────

const goBack = () => router.push('/')

const changePage = (page) => {
  currentPage.value = page
  fetchData()
}

const showTooltip = (result, event, action = 'hover') => emit('showTooltip', result, event, action)

const toggleShowAverageResponseTime = () => {
  showAverageResponseTime.value = !showAverageResponseTime.value
  localStorage.setItem('gatus:show-average-response-time', showAverageResponseTime.value ? 'true' : 'false')
}

const prettifyTimestamp = (ts) => new Date(ts).toLocaleString()

onMounted(() => {
  fetchData()
  fetchUptimeStats()
  fetchAllResults()
})
</script>
