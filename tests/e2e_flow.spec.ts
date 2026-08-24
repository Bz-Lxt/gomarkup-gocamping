import { test, expect } from '@playwright/test'

const USER = process.env.USER_BASE || 'http://127.0.0.1:28311'

test('leader login and see map studio', async ({ page }) => {
  await page.goto(USER + '/login')
  await page.fill('input', 'leader')
  await page.locator('input[type="password"]').fill('leader123')
  await page.getByRole('button', { name: /进入营地/ }).click()
  await expect(page.getByText('路书编辑器')).toBeVisible({ timeout: 15000 })
  await expect(page.locator('#map')).toBeVisible()
})
