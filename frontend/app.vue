<script setup lang="ts">
import { GetSettings, SaveAPIKey, PickFolder } from '~/wailsjs/go/main/App'

const tab = ref<'project' | 'backup' | 'library' | 'history'>('project')
const dest = useDest()
const check = useCheck()
const msg = ref('')
const hasKey = ref(false)
const keyInput = ref('')

async function chooseDest() {
  msg.value = ''
  try {
    const p = await PickFolder('เลือกโฟลเดอร์ปลายทาง')
    if (p) await applyDest(p)
  } catch (e: any) { msg.value = String(e) }
}

async function saveKey() {
  try {
    await SaveAPIKey(keyInput.value)
    hasKey.value = true
    keyInput.value = ''
  } catch (e: any) { msg.value = String(e) }
}

onMounted(async () => {
  const s = await GetSettings()
  hasKey.value = s.hasApiKey
  if (s.lastDest) {
    try { await applyDest(s.lastDest) } catch { /* โฟลเดอร์เดิมอาจหายไป */ }
  }
})
</script>

<template>
  <div class="shell">
    <aside class="side">
      <h1>Photo Backup</h1>
      <nav>
        <button :class="{ on: tab === 'project' }" @click="tab = 'project'">Final Project</button>
        <button :class="{ on: tab === 'backup' }" @click="tab = 'backup'">สำรองรูปภาพ</button>
        <button :class="{ on: tab === 'library' }" @click="tab = 'library'">ค้นหาภาพ</button>
        <button :class="{ on: tab === 'history' }" @click="tab = 'history'">ประวัติและการตรวจสอบ</button>
      </nav>

      <section class="dest">
        <div class="label">โฟลเดอร์ปลายทาง</div>
        <div class="path">{{ dest || 'ยังไม่ได้เลือก' }}</div>
        <button class="btn" @click="chooseDest">เลือกโฟลเดอร์ปลายทาง</button>
        <div v-if="check" class="status" :class="{ warn: check.missing?.length }">
          <template v-if="check.missing?.length">ไฟล์สูญหาย {{ check.missing.length }} จาก {{ check.total }} ไฟล์</template>
          <template v-else>ตรวจแล้ว: ไฟล์ครบ {{ check.total }} ไฟล์</template>
        </div>
        <div v-if="msg" class="status warn">{{ msg }}</div>
      </section>

      <details>
        <summary>Anthropic API key {{ hasKey ? '(ตั้งค่าแล้ว)' : '(ยังไม่ได้ตั้งค่า)' }}</summary>
        <div class="row" style="margin-top:10px">
          <input v-model="keyInput" type="password" placeholder="sk-ant-..." autocomplete="off" />
          <button class="btn" @click="saveKey">บันทึก key</button>
        </div>
      </details>
    </aside>

    <main>
      <ProjectSummaryPanel v-if="tab === 'project'" />
      <BackupPanel v-show="tab === 'backup'" />
      <LibraryPanel v-if="tab === 'library'" />
      <HistoryPanel v-if="tab === 'history'" />
    </main>
  </div>
</template>
