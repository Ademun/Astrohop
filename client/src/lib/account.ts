import { apiClient } from '@/api/client'

const STORAGE_KEY = 'account_key'

let pendingCreate: Promise<string> | null = null

export async function getOrCreateAccountKey(): Promise<string> {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored) {
    apiClient.setToken(stored)
    return stored
  }

  if (!pendingCreate) {
    pendingCreate = createAndStoreAccount().finally(() => {
      pendingCreate = null
    })
  }
  return pendingCreate
}

async function createAndStoreAccount(): Promise<string> {
  const { token } = await apiClient.createAccount()
  localStorage.setItem(STORAGE_KEY, token)
  return token
}
