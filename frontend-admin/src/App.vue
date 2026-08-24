<template>
  <div class="min-h-full w-full p-6 md:p-10">
    <header class="flex flex-wrap items-end justify-between gap-3 mb-8">
      <div>
        <p class="text-xs tracking-[0.3em] uppercase text-muted">GoCamping Admin</p>
        <h1 class="font-display text-3xl mt-1">营地管理台</h1>
      </div>
      <form v-if="!token" class="flex gap-2" @submit.prevent="login">
        <input v-model="user" class="bg-panel border border-line rounded-lg px-3 py-2" placeholder="管理员账号" />
        <input v-model="pass" type="password" class="bg-panel border border-line rounded-lg px-3 py-2" placeholder="密码" />
        <button class="bg-clay text-bg rounded-lg px-4">登录</button>
      </form>
      <button v-else class="border border-line rounded-lg px-3 py-2" @click="token=''">退出</button>
    </header>
    <p v-if="err" class="text-red-400 mb-4">{{ err }}</p>
    <div v-if="token" class="grid md:grid-cols-2 gap-6">
      <section class="bg-panel border border-line rounded-2xl p-5">
        <h2 class="font-display text-xl mb-3">用户</h2>
        <div v-if="!users.length" class="text-muted text-sm">暂无数据</div>
        <ul class="space-y-2 text-sm">
          <li v-for="u in users" :key="u.id" class="flex justify-between border-b border-line pb-2">
            <span>{{ u.nickname }} · {{ u.username }}</span>
            <span class="text-muted">{{ u.role }}</span>
          </li>
        </ul>
      </section>
      <section class="bg-panel border border-line rounded-2xl p-5">
        <h2 class="font-display text-xl mb-3">路书审核</h2>
        <div v-if="!books.length" class="text-muted text-sm">暂无路书</div>
        <ul class="space-y-2 text-sm">
          <li v-for="b in books" :key="b.id">{{ b.title }} · {{ b.visibility }} · {{ Math.round(b.distance_m) }} m</li>
        </ul>
      </section>
      <section class="md:col-span-2 bg-panel border border-line rounded-2xl p-5">
        <h2 class="font-display text-xl mb-3">运行指标</h2>
        <pre class="text-xs text-muted whitespace-pre-wrap">{{ JSON.stringify(metrics, null, 2) }}</pre>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
const API = (import.meta.env.VITE_API_BASE as string) || '/api'
const user = ref('admin')
const pass = ref('admin123')
const token = ref('')
const err = ref('')
const users = ref<any[]>([])
const books = ref<any[]>([])
const metrics = ref<any>({})

async function req(path: string, init: RequestInit = {}) {
  const res = await fetch(`${API}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', Authorization: token.value ? `Bearer ${token.value}` : '', ...(init.headers as any) },
  })
  const j = await res.json()
  if (j.code !== 0) throw new Error(j.message)
  return j.data
}
async function login() {
  err.value = ''
  try {
    const d = await req('/v1/auth/login', { method: 'POST', body: JSON.stringify({ username: user.value, password: pass.value }) })
    if (d.user.role !== 'admin') throw new Error('需要管理员账号')
    token.value = d.token
  } catch (e: any) { err.value = e.message }
}
watch(token, async (t) => {
  if (!t) return
  try {
    users.value = await req('/v1/admin/users')
    books.value = await req('/v1/admin/route-books')
    metrics.value = await req('/v1/admin/metrics')
  } catch (e: any) { err.value = e.message }
})
</script>
