<template>
  <div class="h-full w-full flex flex-col bg-bg text-ink">
    <header class="flex flex-wrap items-center gap-3 px-4 py-3 border-b border-line bg-panel/80 backdrop-blur z-20">
      <div>
        <p class="font-display text-lg leading-none">Mini 驴友足迹</p>
        <p class="text-muted text-xs mt-1">{{ user?.nickname || '未登录' }} · GMT+8</p>
      </div>
      <span class="px-2 py-1 rounded-full text-xs border border-line" :class="trip?.status==='active' ? 'bg-pine/30 text-moss' : 'text-muted'">
        {{ trip?.status || '无行程' }}
      </span>
      <label class="ml-auto flex items-center gap-2 text-sm cursor-pointer">
        <span class="text-muted">模拟断网</span>
        <input type="checkbox" v-model="offline" class="accent-clay" />
      </label>
      <button class="px-3 py-1.5 rounded-lg bg-danger text-white text-sm blink-ready" @click="sendSOS">一键呼救</button>
      <button class="px-3 py-1.5 rounded-lg border border-line text-sm" @click="logout">离开</button>
    </header>

    <div v-if="toast.show" class="fixed top-4 right-4 z-50 bg-panel border border-line px-4 py-3 rounded-xl shadow-xl max-w-sm">
      <div class="flex justify-between gap-4">
        <p>{{ toast.msg }}</p>
        <button @click="toast.show=false">×</button>
      </div>
    </div>

    <div v-if="sos" class="fixed inset-0 z-40 blink bg-danger/80 flex items-center justify-center p-6">
      <div class="bg-bg border border-danger rounded-2xl p-6 w-full max-w-md">
        <p class="font-display text-2xl">掉队呼救</p>
        <p class="mt-2 text-sm">{{ sos.nickname || sos.member_id }} · {{ sos.reason || '手动呼救' }}</p>
        <p class="text-muted text-sm mt-1">{{ sos.lat?.toFixed(5) }}, {{ sos.lon?.toFixed(5) }} · {{ sos.created_at }}</p>
        <div class="flex gap-2 mt-4">
          <button class="flex-1 bg-clay text-bg rounded-lg py-2" @click="flyTo(sos.lat, sos.lon); sos=null">定位队员</button>
          <button class="flex-1 border border-line rounded-lg py-2" @click="sos=null">关闭</button>
        </div>
      </div>
    </div>

    <div v-if="confirm.show" class="fixed inset-0 z-40 bg-black/50 flex items-center justify-center p-6">
      <div class="bg-panel border border-line rounded-2xl p-6 w-full max-w-sm">
        <p class="font-display text-xl">{{ confirm.title }}</p>
        <p class="text-muted text-sm mt-2">{{ confirm.body }}</p>
        <div class="flex gap-2 mt-4">
          <button class="flex-1 bg-danger text-white rounded-lg py-2" @click="confirm.ok(); confirm.show=false">确认</button>
          <button class="flex-1 border border-line rounded-lg py-2" @click="confirm.show=false">取消</button>
        </div>
      </div>
    </div>

    <main class="flex-1 min-h-0 grid grid-cols-1 md:grid-cols-[340px_1fr] lg:grid-cols-[360px_1fr_300px]">
      <aside class="border-r border-line overflow-y-auto p-4 space-y-4 bg-panel/40">
        <section>
          <h2 class="font-display text-xl">路书编辑器</h2>
          <input v-model="draft.title" placeholder="路书标题" class="w-full mt-2 bg-bg border border-line rounded-lg px-3 py-2" />
          <p v-if="formErr.title" class="text-danger text-xs mt-1">{{ formErr.title }}</p>
          <div class="flex gap-2 mt-2">
            <select v-model="pinType" class="flex-1 bg-bg border border-line rounded-lg px-3 py-2">
              <option value="waypoint">途径点</option>
              <option value="camp">露营点</option>
              <option value="water">水源点</option>
              <option value="danger">危险区</option>
            </select>
            <button class="px-3 rounded-lg bg-pine text-ink" @click="saveRoute">保存</button>
          </div>
          <p class="text-muted text-xs mt-2">在卫星底图上点击标注。当前 {{ draft.waypoints.length }} 个点。</p>
          <ul class="mt-2 space-y-1 text-sm">
            <li v-for="(w,i) in draft.waypoints" :key="i" class="flex justify-between gap-2 border-b border-line/60 py-1">
              <span>{{ i+1 }}. {{ labelType(w.type) }} {{ w.note || '' }}</span>
              <button class="text-danger" @click="draft.waypoints.splice(i,1); redraw()">删</button>
            </li>
          </ul>
          <div v-if="!draft.waypoints.length" class="text-muted text-sm py-6 text-center border border-dashed border-line rounded-xl mt-2">还没有点位，点击地图开始编排。</div>
        </section>

        <section>
          <h2 class="font-display text-xl">队伍</h2>
          <div class="flex gap-2 mt-2">
            <input v-model="teamName" placeholder="新队伍名" class="flex-1 bg-bg border border-line rounded-lg px-3 py-2" />
            <button class="px-3 rounded-lg bg-moss text-bg" @click="makeTeam">组建</button>
          </div>
          <div class="flex gap-2 mt-2">
            <input v-model="invite" placeholder="邀请码" class="flex-1 bg-bg border border-line rounded-lg px-3 py-2 uppercase" />
            <button class="px-3 rounded-lg border border-line" @click="doJoin">加入</button>
          </div>
          <select v-model.number="teamId" class="w-full mt-2 bg-bg border border-line rounded-lg px-3 py-2" @change="onTeam">
            <option :value="0">选择队伍</option>
            <option v-for="t in teams" :key="t.id" :value="t.id">{{ t.name }} · {{ t.invite_code }}</option>
          </select>
          <div class="flex gap-2 mt-2">
            <button class="flex-1 bg-clay text-bg rounded-lg py-2 text-sm" @click="ensureTrip">出发 / 继续</button>
            <button class="flex-1 border border-line rounded-lg py-2 text-sm" @click="askFinish">结束</button>
          </div>
          <button class="w-full mt-2 border border-moss/40 rounded-lg py-2 text-sm" @click="runSim">注入离线足迹（含噪）</button>
        </section>
      </aside>

      <section class="relative min-h-[52vh]">
        <div id="map" class="absolute inset-0"></div>
        <div class="absolute bottom-0 left-0 right-0 h-36 bg-panel/85 border-t border-line">
          <div id="profile" class="w-full h-full"></div>
        </div>
      </section>

      <aside class="border-l border-line overflow-y-auto p-4 space-y-4 bg-panel/40 hidden lg:block">
        <section>
          <h2 class="font-display text-xl">位置墙</h2>
          <ul class="mt-2 space-y-2 text-sm">
            <li v-for="p in live" :key="p.member_id" class="flex justify-between">
              <span>{{ p.nickname || p.member_id }}</span>
              <span class="text-muted">{{ p.lat?.toFixed(4) }}, {{ p.lon?.toFixed(4) }}</span>
            </li>
          </ul>
          <div v-if="!live.length" class="text-muted text-sm mt-3">等待队员上线…</div>
        </section>
        <section>
          <h2 class="font-display text-xl">风险 / ETA</h2>
          <p class="mt-2 text-sm">等级 <b class="text-clay">{{ risk.level || '-' }}</b> · 分数 {{ risk.score ?? '-' }}</p>
          <p class="text-muted text-xs mt-1">离散 {{ risk.dispersion ?? '-' }} · 水源 {{ risk.water_dist_m ?? '-' }} m · 日落 {{ risk.sunset_left_h ?? '-' }} h</p>
          <p class="text-sm mt-3">ETA {{ eta.eta || '-' }}</p>
          <p class="text-muted text-xs">剩余 {{ eta.remaining_m ? Math.round(eta.remaining_m) : '-' }} m · {{ eta.corrected_kmh || '-' }} km/h</p>
          <button class="mt-3 w-full border border-line rounded-lg py-2 text-sm" @click="doBacktrack">原路返回</button>
        </section>
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import L from 'leaflet'
import * as echarts from 'echarts'
import { api, getToken, setToken, TILE_URL, wsURL } from '../api'
import { cachePoint, drainPoints } from '../offline'

const router = useRouter()
if (!getToken()) router.replace('/login')

const user = ref<any>(JSON.parse(localStorage.getItem('gc_user') || 'null'))
const offline = ref(false)
const pinType = ref('waypoint')
const draft = reactive<any>({ title: '清凉峰西坡小众线', waypoints: [] as any[] })
const formErr = reactive({ title: '' })
const teams = ref<any[]>([])
const teamId = ref(0)
const teamName = ref('')
const invite = ref('')
const trip = ref<any>(null)
const live = ref<any[]>([])
const risk = ref<any>({})
const eta = ref<any>({})
const sos = ref<any>(null)
const toast = reactive({ show: false, msg: '' })
const confirm = reactive({ show: false, title: '', body: '', ok: () => {} })
let map: L.Map | null = null
let layer: L.LayerGroup | null = null
let liveLayer: L.LayerGroup | null = null
let chart: echarts.ECharts | null = null
let ws: WebSocket | null = null
let tick: number | null = null
let routeId = 0

function show(msg: string) {
  toast.msg = msg
  toast.show = true
  setTimeout(() => { toast.show = false }, 5000)
}
function labelType(t: string) {
  return ({ camp: '露营', water: '水源', danger: '危险', waypoint: '途径' } as any)[t] || t
}
function logout() {
  setToken('')
  router.replace('/login')
}

function initMap() {
  map = L.map('map', { zoomControl: true }).setView([30.14, 118.87], 13)
  L.tileLayer(TILE_URL, { maxZoom: 16, attribution: '本地地形瓦片' }).addTo(map)
  layer = L.layerGroup().addTo(map)
  liveLayer = L.layerGroup().addTo(map)
  map.on('click', (e: L.LeafletMouseEvent) => {
    draft.waypoints.push({
      type: pinType.value, lat: e.latlng.lat, lon: e.latlng.lng,
      note: '', risk_weight: pinType.value === 'danger' ? 4 : 1,
      radius_m: pinType.value === 'danger' ? 80 : null, polygon: [],
    })
    redraw()
  })
}

function colorOf(t: string) {
  return ({ camp: '#C47A3A', water: '#4E8FA8', danger: '#D4533A', waypoint: '#7C9A5A' } as any)[t] || '#E8E1D1'
}

function redraw() {
  layer?.clearLayers()
  const line: L.LatLngExpression[] = []
  draft.waypoints.forEach((w: any, i: number) => {
    const m = L.circleMarker([w.lat, w.lon], { radius: 8, color: colorOf(w.type), fillOpacity: 0.9, weight: 2 })
    m.bindTooltip(`${i + 1} ${labelType(w.type)}`)
    layer?.addLayer(m)
    if (w.type === 'danger' && w.radius_m) {
      layer?.addLayer(L.circle([w.lat, w.lon], { radius: w.radius_m, color: '#D4533A', fillOpacity: 0.12 }))
    }
    if (w.type !== 'danger') line.push([w.lat, w.lon])
  })
  if (line.length >= 2) layer?.addLayer(L.polyline(line, { color: '#C47A3A', weight: 3, opacity: 0.85 }))
}

function flyTo(lat: number, lon: number) {
  map?.flyTo([lat, lon], 15)
}

async function saveRoute() {
  formErr.title = draft.title.trim() ? '' : '标题必填'
  if (formErr.title) { show('请先修正表单'); return }
  try {
    const body = { ...draft, visibility: 'public' }
    const saved = routeId ? await api.updateRoute(routeId, body) : await api.saveRoute(body)
    routeId = saved.id
    draft.waypoints = saved.waypoints || draft.waypoints
    show('路书已保存')
    const prof = await api.elevation(routeId)
    drawProfile(prof)
  } catch (e: any) { show(e.message) }
}

function drawProfile(prof: any) {
  const el = document.getElementById('profile')
  if (!el) return
  if (!chart) chart = echarts.init(el)
  const xs = (prof.along_m || []).map((v: number) => Math.round(v))
  const ys = (prof.samples || []).map((s: any) => s.elevation)
  chart.setOption({
    backgroundColor: 'transparent',
    grid: { left: 44, right: 16, top: 18, bottom: 24 },
    xAxis: { type: 'category', data: xs, axisLabel: { color: '#9AA48A' }, axisLine: { lineStyle: { color: '#2A3324' } } },
    yAxis: { type: 'value', axisLabel: { color: '#9AA48A' }, splitLine: { lineStyle: { color: '#2A3324' } } },
    tooltip: { trigger: 'axis' },
    series: [{ type: 'line', data: ys, smooth: true, areaStyle: { color: 'rgba(196,122,58,0.25)' }, lineStyle: { color: '#C47A3A' }, showSymbol: false }],
  })
}

async function refreshTeams() {
  teams.value = await api.teams()
  if (!teamId.value && teams.value.length) {
    teamId.value = teams.value[0].id
    await onTeam()
  }
}

async function makeTeam() {
  if (!teamName.value.trim()) { show('请填写队伍名'); return }
  if (!routeId) await saveRoute()
  const t = await api.createTeam(teamName.value.trim(), routeId || undefined)
  teams.value.unshift(t)
  teamId.value = t.id
  show(`邀请码 ${t.invite_code}`)
}

async function doJoin() {
  try {
    const t = await api.joinTeam(invite.value)
    teams.value.unshift(t)
    teamId.value = t.id
    show('已加入 ' + t.name)
  } catch (e: any) { show(e.message) }
}

async function onTeam() {
  if (!teamId.value) return
  const list = await api.listTrips(teamId.value)
  trip.value = list[0] || null
  if (trip.value) connectWS()
}

async function ensureTrip() {
  if (!teamId.value) { show('先选择队伍'); return }
  try {
    if (!trip.value) trip.value = await api.createTrip(teamId.value)
    if (trip.value.status === 'draft' || trip.value.status === 'paused') {
      trip.value = await api.startTrip(trip.value.id)
    }
    connectWS()
    show('行程已激活')
  } catch (e: any) { show(e.message) }
}

function askFinish() {
  if (!trip.value) return
  confirm.title = '结束行程？'
  confirm.body = '结束后将停止实时广播，轨迹仍可溯源。'
  confirm.ok = async () => {
    trip.value = await api.finishTrip(trip.value.id)
    show('行程已结束')
  }
  confirm.show = true
}

function connectWS() {
  if (!trip.value || offline.value) return
  ws?.close()
  ws = new WebSocket(wsURL(trip.value.id))
  ws.onmessage = (ev) => {
    const msg = JSON.parse(ev.data)
    if (msg.type === 'pos_update') {
      const p = msg.payload
      const i = live.value.findIndex((x) => x.member_id === p.member_id)
      if (i >= 0) live.value[i] = { ...live.value[i], ...p }
      else live.value.push(p)
      paintLive()
    }
    if (msg.type === 'sos') sos.value = msg.payload
    if (msg.type === 'risk_report') risk.value = msg.payload
    if (msg.type === 'eta_update') eta.value = msg.payload
  }
}

function paintLive() {
  liveLayer?.clearLayers()
  live.value.forEach((p) => {
    liveLayer?.addLayer(L.circleMarker([p.lat, p.lon], { radius: 7, color: '#E8E1D1', fillColor: '#3F6B4A', fillOpacity: 1 }))
  })
}

async function sendSOS() {
  if (!trip.value) { show('先出发'); return }
  const last = live.value[0] || draft.waypoints[0] || { lat: 30.14, lon: 118.87 }
  try {
    sos.value = await api.sos(trip.value.id, last.lat, last.lon, '手动呼救')
  } catch (e: any) { show(e.message) }
}

async function runSim() {
  if (!trip.value) await ensureTrip()
  try {
    const res = await api.simulate(trip.value.id)
    show(`合并完成 接受 ${res.accepted} / 拒绝 ${res.rejected}`)
    eta.value = await api.eta(trip.value.id)
    risk.value = await api.risk(trip.value.id)
    const tr = await api.tracks(trip.value.id)
    if (tr.length) {
      const line = tr.map((p: any) => [p.lat, p.lon] as L.LatLngExpression)
      liveLayer?.addLayer(L.polyline(line, { color: '#7C9A5A', weight: 2, dashArray: '4 6' }))
    }
  } catch (e: any) { show(e.message) }
}

async function doBacktrack() {
  if (!trip.value) return
  try {
    const b = await api.backtrack(trip.value.id)
    const line = (b.path || []).map((p: any) => [p.lat, p.lon])
    liveLayer?.addLayer(L.polyline(line, { color: '#4E8FA8', weight: 3 }))
    show(`回溯 ${Math.round(b.distance_m)} 米至 ${b.to?.kind}`)
  } catch (e: any) { show(e.message) }
}

async function pulse() {
  if (!trip.value || trip.value.status !== 'active') return
  const last = draft.waypoints.find((w: any) => w.type !== 'danger') || { lat: 30.14, lon: 118.87 }
  const jitter = () => (Math.random() - 0.5) * 0.0004
  const lat = last.lat + jitter()
  const lon = last.lon + jitter()
  const point = { trip_id: trip.value.id, lat, lon, recorded_at: new Date().toISOString() }
  if (offline.value) {
    await cachePoint(point)
    return
  }
  try {
    const pending = await drainPoints(trip.value.id)
    if (pending.length) await api.batch(trip.value.id, pending)
    await api.position(trip.value.id, lat, lon)
  } catch (e: any) {
    await cachePoint(point)
  }
}

onMounted(async () => {
  initMap()
  try {
    const list = await api.routes()
    if (list.length) {
      const rb = await api.getRoute(list[0].id)
      routeId = rb.id
      draft.title = rb.title
      draft.waypoints = rb.waypoints || []
      redraw()
      if (draft.waypoints[0]) map?.setView([draft.waypoints[0].lat, draft.waypoints[0].lon], 13)
      try { drawProfile(await api.elevation(routeId)) } catch {}
    }
    await refreshTeams()
  } catch (e: any) {
    if (e.code === 40100) router.replace('/login')
  }
  tick = window.setInterval(pulse, 4000)
})

onUnmounted(() => {
  if (tick) clearInterval(tick)
  ws?.close()
  chart?.dispose()
  map?.remove()
})
</script>
