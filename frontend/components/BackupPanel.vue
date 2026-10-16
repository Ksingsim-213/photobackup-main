<script setup lang="ts">
import { PickFolder, MoveImages, AnalyzeImages } from '~/wailsjs/go/main/App'
import { EventsOn } from '~/wailsjs/runtime/runtime'

const dest = useDest()
const refreshKey = useRefreshKey()

const source = ref('')
const deleteSource = ref(true)
const autoAnalyze = ref(true)
const running = ref(false)
const phase = ref<'' | 'move' | 'analyze'>('')
const mprog = ref({ done: 0, total: 0, file: '' })
const aprog = ref({ done: 0, total: 0, file: '' })
const result = ref<any>(null)
const aresult = ref<any>(null)
const error = ref('')
const notice = ref('')

const off: Array<() => void> = []
onMounted(() => {
  off.push(EventsOn('move:progress', (p: any) => (mprog.value = p)))
  off.push(EventsOn('analyze:progress', (p: any) => (aprog.value = p)))
})
onBeforeUnmount(() => off.forEach((f) => f()))

async function pickSource() {
  const p = await PickFolder('เลือกโฟลเดอร์ต้นทาง (flash drive หรือโทรศัพท์)')
  if (p) source.value = p
}

const pct = (p: { done: number; total: number }) => (p.total ? Math.round((p.done / p.total) * 100) : 0)

async function run() {
  error.value = notice.value = ''
  result.value = aresult.value = null
  mprog.value = { done: 0, total: 0, file: '' }
  aprog.value = { done: 0, total: 0, file: '' }
  running.value = true
  phase.value = 'move'
  try {
    result.value = await MoveImages(source.value, dest.value, deleteSource.value)
    await applyDest(dest.value) // ตรวจความถูกต้องอีกครั้ง
    refreshKey.value++
    if (autoAnalyze.value && result.value.moved > 0) {
      phase.value = 'analyze'
      try {
        aresult.value = await AnalyzeImages(dest.value)
        refreshKey.value++
      } catch (e: any) {
        notice.value = 'สำรองไฟล์เสร็จแล้ว แต่ยังวิเคราะห์ภาพไม่ได้: ' + String(e)
      }
    }
  } catch (e: any) {
    error.value = String(e)
  } finally {
    running.value = false
    phase.value = ''
  }
}

async function analyzeOnly() {
  error.value = notice.value = ''
  running.value = true
  phase.value = 'analyze'
  aprog.value = { done: 0, total: 0, file: '' }
  try {
    aresult.value = await AnalyzeImages(dest.value)
    refreshKey.value++
  } catch (e: any) { error.value = String(e) } finally { running.value = false; phase.value = '' }
}
</script>

<template>
  <h2>สำรองรูปภาพ</h2>
  <p class="lead">เลือกโฟลเดอร์ต้นทางและปลายทาง โปรแกรมจะข้ามไฟล์ที่ชื่อซ้ำ ย้ายไฟล์ แล้วบันทึกประวัติลงฐานข้อมูล</p>

  <div class="panel">
    <h3>ต้นทาง</h3>
    <div class="row">
      <div class="field-path grow">{{ source || 'ยังไม่ได้เลือกโฟลเดอร์ต้นทาง' }}</div>
      <button class="btn" :disabled="running" @click="pickSource">เลือกโฟลเดอร์</button>
    </div>
    <div class="row" style="margin-top:14px; gap:22px">
      <label class="check"><input v-model="deleteSource" type="checkbox" /> ย้าย (ลบไฟล์ต้นทางหลังสำรองสำเร็จ)</label>
      <label class="check"><input v-model="autoAnalyze" type="checkbox" /> วิเคราะห์ภาพด้วย Gemini หลังสำรองเสร็จ</label>
    </div>
    <div class="row" style="margin-top:16px">
      <button class="btn primary" :disabled="running || !source || !dest" @click="run">
        {{ running && phase === 'move' ? 'กำลังสำรอง...' : 'เริ่มสำรองรูปภาพ' }}
      </button>
      <button class="btn" :disabled="running || !dest" @click="analyzeOnly">วิเคราะห์ภาพที่ยังไม่มีคำอธิบาย</button>
      <span v-if="!dest" class="small">เลือกโฟลเดอร์ปลายทางที่แถบด้านซ้ายก่อน</span>
    </div>
  </div>

  <div v-if="running" class="panel">
    <template v-if="phase === 'move'">
      <h3>กำลังสำรองไฟล์ {{ mprog.done }} / {{ mprog.total }}</h3>
      <div class="bar"><i :style="{ width: pct(mprog) + '%' }" /></div>
      <div class="small" style="margin-top:6px">{{ mprog.file }}</div>
    </template>
    <template v-else>
      <h3>กำลังวิเคราะห์ภาพ {{ aprog.done }} / {{ aprog.total || '...' }}</h3>
      <div class="bar"><i :style="{ width: pct(aprog) + '%' }" /></div>
      <div class="small" style="margin-top:6px">{{ aprog.file }}</div>
    </template>
  </div>

  <div v-if="error" class="errbox">{{ error }}</div>
  <div v-if="notice" class="warnbox">{{ notice }}</div>

  <div v-if="result" class="panel">
    <h3>สรุปผลการสำรอง</h3>
    <div class="stats">
      <div><b>{{ result.found }}</b><span>พบในต้นทาง</span></div>
      <div><b>{{ result.moved }}</b><span>สำรองสำเร็จ</span></div>
      <div><b>{{ result.skipped }}</b><span>ข้าม (ชื่อซ้ำ)</span></div>
      <div><b>{{ result.failed }}</b><span>ผิดพลาด</span></div>
      <div><b>{{ result.fallbacks }}</b><span>ใช้ Copy + Delete</span></div>
    </div>
    <div>เวลาที่ใช้ทั้งหมด: <b>{{ result.elapsed }}</b></div>
    <div v-if="result.warnings?.length" class="warnbox">
      แจ้งเตือน {{ result.warnings.length }} รายการ
      <ul><li v-for="w in result.warnings" :key="w">{{ w }}</li></ul>
    </div>
    <div v-if="result.errors?.length" class="errbox">
      ผิดพลาด {{ result.errors.length }} รายการ
      <ul><li v-for="w in result.errors" :key="w">{{ w }}</li></ul>
    </div>
  </div>

  <div v-if="aresult" class="panel">
    <h3>ผลการวิเคราะห์ภาพ</h3>
    <div>วิเคราะห์สำเร็จ {{ aresult.done }} จาก {{ aresult.total }} ภาพ ใช้เวลา {{ (aresult.elapsedMs / 1000).toFixed(1) }} วินาที</div>
    <div v-if="aresult.errors?.length" class="errbox">
      ผิดพลาด {{ aresult.failed }} ภาพ
      <ul><li v-for="w in aresult.errors" :key="w">{{ w }}</li></ul>
    </div>
  </div>
</template>
