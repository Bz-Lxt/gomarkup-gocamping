const DB = 'gocamping'
const STORE = 'gps'

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB, 1)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(STORE)) db.createObjectStore(STORE, { keyPath: 'id', autoIncrement: true })
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

export async function cachePoint(p: { trip_id: number; lat: number; lon: number; recorded_at: string }) {
  const db = await openDB()
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE, 'readwrite')
    tx.objectStore(STORE).add(p)
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

export async function drainPoints(tripId: number) {
  const db = await openDB()
  const all: any[] = await new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, 'readonly')
    const req = tx.objectStore(STORE).getAll()
    req.onsuccess = () => resolve(req.result || [])
    req.onerror = () => reject(req.error)
  })
  const mine = all.filter((x) => x.trip_id === tripId)
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE, 'readwrite')
    const store = tx.objectStore(STORE)
    mine.forEach((x) => store.delete(x.id))
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
  return mine.map(({ lat, lon, recorded_at }) => ({ lat, lon, recorded_at }))
}
