<script setup lang="ts">
import { ListDB, CheckIntegrity } from '~/wailsjs/go/main/App'

const dest = useDest()
const check = useCheck()
const refreshKey = useRefreshKey()
const rows = ref<any[]>([])
const error = ref('')
const loading = ref(false)

const missingIds = computed(() => new Set((check.value?.missing ?? []).map((m: any) => m.id)))

async function load() {
  error.value = ''
  if (!dest.value) { rows.value = []; return }
  loading.value = true
  try { rows.value = (await ListDB(dest.value)) ?? [] } catch (e: any) { error.value = String(e) } finally { loading.value = false }
}

async function recheck() {
  error.value = ''
  try { check.value = await CheckIntegrity(dest.value) } catch (e: any) { error.value = String(e) }
}

watch([dest, refreshKey], load, { immediate: true })
</script>

<template>
  <h2>ประวัติและการตรวจสอบ</h2>
  <p class="lead">แสดงเฉพาะไฟล์ที่ถูกสำรองไว้ในโฟลเดอร์ปลายทางที่เลือก และเทียบกับไฟล์จริงในเครื่อง</p>

  <div v-if="!dest" class="warnbox">กรุณาระบุโฟลเดอร์ปลายทางก่อน (เลือกที่แถบด้านซ้าย)</div>
  <template v-else>
    <div class="panel">
      <div class="row">
        <div class="grow">
          <b>ตรวจสอบความถูกต้อง</b>
          <div class="small">
            <template v-if="check">ในฐานข้อมูล {{ check.total }} ไฟล์ · พบไฟล์จริง {{ check.present }} ไฟล์ · สูญหาย {{ check.missing.length }} ไฟล์ ({{ check.elapsedMs }} ms)</template>
          </div>
        </div>
        <button class="btn primary" @click="recheck">ตรวจสอบอีกครั้ง (/check)</button>
      </div>
      <div v-if="check?.missing?.length" class="warnbox">
        ไฟล์ที่มีประวัติในฐานข้อมูลแต่ไม่พบในโฟลเดอร์:
        <ul><li v-for="m in check.missing" :key="m.id">{{ m.fileName }}</li></ul>
      </div>
    </div>

    <div v-if="error" class="errbox">{{ error }}</div>
    <div v-if="rows.length" class="tablewrap">
      <table>
        <thead><tr><th>ชื่อไฟล์</th><th>ขนาด</th><th>สำรองเมื่อ</th><th>คำอธิบาย</th><th>แท็ก</th><th>สถานะ</th></tr></thead>
        <tbody>
          <tr v-for="r in rows" :key="r.id" :class="{ missing: missingIds.has(r.id) }">
            <td>{{ r.fileName }}</td>
            <td>{{ fmtSize(r.size) }}</td>
            <td>{{ fmtDate(r.backedUpAt) }}</td>
            <td>{{ r.caption }}</td>
            <td><span v-for="t in r.tags" :key="t.id" class="chip" style="margin-right:4px">{{ t.name }}</span></td>
            <td><span v-if="missingIds.has(r.id)" class="tag-missing">สูญหาย</span><span v-else>ปกติ</span></td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else-if="!loading" class="empty">ยังไม่มีประวัติการสำรองในโฟลเดอร์นี้</div>
  </template>
</template>
