import { SetDestination } from '~/wailsjs/go/main/App'

export const useDest = () => useState<string>('dest', () => '')
export const useCheck = () => useState<any>('check', () => null)
// เพิ่มค่านี้เมื่อข้อมูลใน DB เปลี่ยน เพื่อให้หน้าคลังภาพโหลดใหม่
export const useRefreshKey = () => useState<number>('refreshKey', () => 0)

export async function applyDest(path: string) {
  const dest = useDest()
  const check = useCheck()
  const r = await SetDestination(path)
  dest.value = path
  check.value = r
  return r
}

export function fmtSize(n: number) {
  if (n < 1024) return n + ' B'
  if (n < 1024 * 1024) return (n / 1024).toFixed(0) + ' KB'
  return (n / 1024 / 1024).toFixed(1) + ' MB'
}

export function fmtDate(s: string) {
  return new Date(s).toLocaleString('th-TH', { dateStyle: 'medium', timeStyle: 'short' })
}
