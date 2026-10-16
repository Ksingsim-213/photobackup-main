<script setup lang="ts">
import { GetThumbnail } from '~/wailsjs/go/main/App'

const props = defineProps<{ img: any }>()
defineEmits<{ (e: 'open', img: any): void }>()

const el = ref<HTMLElement>()
const src = ref('')
const failed = ref(false)
let io: IntersectionObserver | undefined

onMounted(() => {
  io = new IntersectionObserver(async (entries) => {
    if (!entries[0].isIntersecting) return
    io?.disconnect()
    try { src.value = await GetThumbnail(props.img.id) } catch { failed.value = true }
  }, { rootMargin: '200px' })
  if (el.value) io.observe(el.value)
})
onBeforeUnmount(() => io?.disconnect())
</script>

<template>
  <button ref="el" class="card" @click="$emit('open', img)">
    <div class="ph">
      <img v-if="src" :src="src" :alt="img.caption || img.fileName" />
      <span v-else-if="failed">ไม่มีภาพตัวอย่าง</span>
    </div>
    <div class="meta">
      <div class="name">{{ img.fileName }}</div>
      <div class="cap">{{ img.caption || 'ยังไม่มีคำอธิบาย' }}</div>
    </div>
  </button>
</template>
