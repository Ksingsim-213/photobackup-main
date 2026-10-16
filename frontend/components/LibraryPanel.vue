<script setup lang="ts">
import { SearchImages, GetTagCloud, GetPreview, AnalyzeImage } from '~/wailsjs/go/main/App'

const dest = useDest()
const refreshKey = useRefreshKey()

const caption = ref('')
const selected = ref<string[]>([])
const matchAll = ref(false)
const cloud = ref<{ name: string; count: number }[]>([])
const results = ref<any[]>([])
const loading = ref(false)
const error = ref('')
const open = ref<any>(null)
const openSrc = ref('')
const busy = ref(false)

const range = computed(() => {
  const c = cloud.value.map((t) => t.count)
  return { min: Math.min(...c), max: Math.max(...c) }
})
function tagSize(n: number) {
  const { min, max } = range.value
  const t = max === min ? 0.4 : (n - min) / (max - min)
  return Math.round(14 + t * 26)
}

async function loadCloud() {
  if (!dest.value) { cloud.value = []; return }
  try { cloud.value = (await GetTagCloud(dest.value)) ?? [] } catch (e: any) { error.value = String(e) }
}

async function search() {
  if (!dest.value) { results.value = []; return }
  loading.value = true
  error.value = ''
  try {
    results.value = (await SearchImages(dest.value, caption.value, selected.value, matchAll.value)) ?? []
  } catch (e: any) { error.value = String(e) } finally { loading.value = false }
}

function toggle(name: string) {
  const i = selected.value.indexOf(name)
  if (i >= 0) selected.value.splice(i, 1)
  else selected.value.push(name)
}

let timer: ReturnType<typeof setTimeout>
watch([caption, () => selected.value.length, matchAll], () => {
  clearTimeout(timer)
  timer = setTimeout(search, 250)
})
watch([dest, refreshKey], () => { loadCloud(); search() }, { immediate: true })

async function show(img: any) {
  open.value = img
  openSrc.value = ''
  try { openSrc.value = await GetPreview(img.id) } catch { /* ใช้ placeholder */ }
}

async function reanalyze() {
  busy.value = true
  try {
    const updated = await AnalyzeImage(open.value.id)
    open.value = updated
    const i = results.value.findIndex((r) => r.id === updated.id)
    if (i >= 0) results.value[i] = updated
    await loadCloud()
  } catch (e: any) { error.value = String(e) } finally { busy.value = false }
}
</script>

<template>
  <h2>ค้นหาภาพ</h2>
  <p class="lead">ค้นจากคำอธิบายภาพ หรือคลิกแท็กใน Tag Cloud เพื่อกรองภาพ ใช้ได้ทั้งสองอย่างพร้อมกัน</p>

  <div v-if="!dest" class="panel empty">เลือกโฟลเดอร์ปลายทางที่แถบด้านซ้ายก่อน</div>
  <template v-else>
    <div class="panel">
      <div class="row">
        <input v-model="caption" class="grow" type="search" placeholder="พิมพ์คำที่อยู่ในคำอธิบายภาพ เช่น ทะเล, อาหาร, เด็ก" />
        <label class="check"><input v-model="matchAll" type="checkbox" /> ต้องมีครบทุกแท็กที่เลือก</label>
        <button v-if="caption || selected.length" class="btn" @click="caption = ''; selected = []">ล้างตัวกรอง</button>
      </div>
    </div>

    <div class="panel">
      <h3>Tag Cloud</h3>
      <div v-if="!cloud.length" class="small">ยังไม่มีแท็ก วิเคราะห์ภาพที่หน้า "สำรองรูปภาพ" ก่อน</div>
      <div v-else class="cloud">
        <button v-for="t in cloud" :key="t.name" :class="{ on: selected.includes(t.name) }"
          :style="{ fontSize: tagSize(t.count) + 'px' }" @click="toggle(t.name)">
          {{ t.name }}<sup>{{ t.count }}</sup>
        </button>
      </div>
    </div>

    <div v-if="error" class="errbox">{{ error }}</div>
    <div class="small" style="margin-bottom:10px">{{ loading ? 'กำลังค้นหา...' : `พบ ${results.length} ภาพ` }}</div>
    <div v-if="results.length" class="grid">
      <ImageCard v-for="im in results" :key="im.id" :img="im" @open="show" />
    </div>
    <div v-else-if="!loading" class="empty">ไม่พบภาพที่ตรงกับเงื่อนไข</div>
  </template>

  <div v-if="open" class="modal" @click.self="open = null">
    <div class="box">
      <img v-if="openSrc" :src="openSrc" :alt="open.caption || open.fileName" />
      <div v-else class="empty">กำลังโหลดภาพ...</div>
      <div class="info">
        <div><b>{{ open.fileName }}</b><div class="small">{{ fmtSize(open.size) }} · สำรองเมื่อ {{ fmtDate(open.backedUpAt) }}</div></div>
        <div>{{ open.caption || 'ยังไม่มีคำอธิบาย' }}</div>
        <div class="chips"><span v-for="t in open.tags" :key="t.id" class="chip">{{ t.name }}</span></div>
        <div class="small" style="word-break:break-all">{{ open.path }}</div>
        <div class="row">
          <button class="btn" :disabled="busy" @click="reanalyze">{{ busy ? 'กำลังวิเคราะห์...' : 'วิเคราะห์ใหม่' }}</button>
          <button class="btn" @click="open = null">ปิด</button>
        </div>
      </div>
    </div>
  </div>
</template>
