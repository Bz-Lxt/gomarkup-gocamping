<template>
  <div class="min-h-full w-full grid place-items-center relative overflow-hidden">
    <div class="absolute inset-0 opacity-40" style="background:radial-gradient(ellipse at 20% 10%,#3F6B4A 0%,transparent 50%),radial-gradient(ellipse at 80% 90%,#C47A3A 0%,transparent 46%);" />
    <form class="relative w-[min(420px,92vw)] bg-panel/90 border border-line rounded-3xl p-8 shadow-2xl" @submit.prevent="submit">
      <p class="font-display text-moss tracking-[0.3em] text-xs uppercase">Mini 驴友足迹</p>
      <h1 class="font-display text-3xl mt-2 mb-1">暮色林线</h1>
      <p class="text-muted text-sm mb-6">出发前编路书，行进中共享位置，断网也能把脚印带回来。</p>
      <label class="block text-sm mb-1">账号</label>
      <input v-model="username" class="w-full bg-bg border border-line rounded-xl px-3 py-2 mb-3 outline-none focus:border-moss" />
      <p v-if="errors.username" class="text-danger text-xs -mt-2 mb-2">{{ errors.username }}</p>
      <label class="block text-sm mb-1">密码</label>
      <input v-model="password" type="password" class="w-full bg-bg border border-line rounded-xl px-3 py-2 mb-3 outline-none focus:border-moss" />
      <p v-if="errors.password" class="text-danger text-xs -mt-2 mb-2">{{ errors.password }}</p>
      <button class="w-full bg-clay text-bg font-semibold rounded-xl py-2.5 hover:brightness-110 disabled:opacity-50" :disabled="loading">
        {{ loading ? '正在入山…' : '进入营地' }}
      </button>
      <p v-if="err" class="text-danger text-sm mt-3">{{ err }}</p>
      <p class="text-muted text-xs mt-5 leading-6">
        测试账号 leader / leader123 · member / member123<br />
        管理员请走后台 28312
      </p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, setToken } from '../api'

const router = useRouter()
const username = ref('leader')
const password = ref('leader123')
const loading = ref(false)
const err = ref('')
const errors = reactive({ username: '', password: '' })

async function submit() {
  errors.username = username.value.trim() ? '' : '请填写账号'
  errors.password = password.value.length >= 6 ? '' : '密码至少 6 位'
  if (errors.username || errors.password) {
    err.value = '请先修正表单'
    return
  }
  loading.value = true
  err.value = ''
  try {
    const data = await api.login(username.value.trim(), password.value)
    setToken(data.token)
    localStorage.setItem('gc_user', JSON.stringify(data.user))
    router.replace('/')
  } catch (e: any) {
    err.value = e.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>
