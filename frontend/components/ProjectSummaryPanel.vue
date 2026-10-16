<template>
  <div class="project-shell">
    <header class="project-header">Final Project</header>

    <h2 class="project-title">
      โปรแกรมสำหรับ Backup รูปภาพจาก flash drive (หรือ โทรศัพท์มือถือ)
    </h2>

    <ol class="project-list">
      <li>ใช้ Wails สร้าง โดย Frontend เป็น Nuxt</li>
      <li>ใช้ SQLite เป็น database และ ใช้ GORM ในการเชื่อมต่อ</li>
      <li>ความสามารถของการ Backup เทียบกับข้อบกพร่อง</li>
      <li>มีการวิเคราะห์ภาพรายละเอียดและอธิบายภาพได้ เช่น กำหนดแปลงภาพในภาษาภาพ</li>
      <li>
        ค้นหาภาพและแสดงภาพ (หลายภาพ) ได้จากคำอธิบายภาพ หรือ Tag ของภาพ
        (อย่างใดอย่างหนึ่ง หรือ ทั้ง 2 อย่างก็ได้)
        <ol class="nested-list">
          <li>หากค้นหาจาก Tag ของภาพที่ตั้ง Tag Cloud ด้วย</li>
          <li>หากค้นหาจากคำอธิบายภาพ ต้องมีข้อมูลภายในภาพและแสดงผลได้ตามคำอธิบาย</li>
        </ol>
      </li>
    </ol>

    <section class="project-section">
      <h3>เงื่อนไข</h3>
      <ul>
        <li>จำนวนสมาชิก 4 คน (จำนวนกลุ่มทั้งหมด 4 คน)</li>
        <li>ทำงานได้บนระบบ Windows และ macOS</li>
      </ul>
    </section>

    <section class="project-section">
      <h3>สางาน</h3>
      <ul>
        <li>คะแนน Progress Demo อธิบายความคืบหน้าของงาน และ เทคนิครภาคไคที่ใช้ได้ (1 ครั้ง ไม่เกิน 15 นาที ก่อนช่วงสัปดาห์ Final)</li>
        <li>คะแนน Technical Demo เน้นอธิบายเทคนิคการพัฒนาหรือสิ่งที่น่าสนใจ (1 ครั้ง ช่วงสัปดาห์ Final)</li>
        <li>เวลาการ Demo ไม่เกิน 5 ชม.</li>
        <li>เทคนิคการใช้ AI เป็นตัวช่วยในการสื่อสารและการพัฒนางาน</li>
      </ul>
    </section>
  </div>
</template>

<style scoped>
.project-shell {
  min-height: 100%;
  background: #1f2733;
  color: #edf3f8;
  padding: 28px 34px 40px;
  border-radius: 18px;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.04);
}

.project-header {
  font-size: clamp(2.2rem, 3vw, 4rem);
  font-weight: 700;
  letter-spacing: -0.05em;
  margin: 0 0 20px;
}

.project-title {
  margin: 0 0 18px;
  font-size: clamp(1.9rem, 2.5vw, 3.6rem);
  line-height: 1.3;
  letter-spacing: -0.04em;
  font-weight: 600;
  max-width: 1200px;
}

.project-list,
.project-section ul {
  margin: 0;
  padding-left: 1.8rem;
}

.project-list {
  font-size: clamp(1.1rem, 1.7vw, 2rem);
  line-height: 1.7;
  max-width: 1240px;
}

.project-list > li {
  margin-bottom: 0.45rem;
}

.nested-list {
  margin-top: 0.5rem;
  padding-left: 1.8rem;
  font-size: 0.95em;
}

.project-section {
  margin-top: 42px;
}

.project-section h3 {
  margin: 0 0 14px;
  font-size: clamp(2rem, 2.4vw, 3rem);
  font-weight: 700;
  letter-spacing: -0.04em;
}

.project-section ul {
  font-size: clamp(1.1rem, 1.5vw, 2rem);
  line-height: 1.7;
  list-style: disc;
}

.project-section li {
  margin-bottom: 0.45rem;
}
</style>
