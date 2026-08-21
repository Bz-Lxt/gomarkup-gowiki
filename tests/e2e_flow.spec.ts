import { test, expect } from '@playwright/test'

const base = process.env.GOWIKI_WEB || 'http://web'

test('critical wiki path', async ({ page }) => {
  await page.goto(base + '/#/login')
  await page.getByPlaceholder('admin@gowiki.dev').fill('admin@gowiki.dev')
  await page.getByPlaceholder('至少 6 位').fill('admin123')
  await page.getByRole('button', { name: '进入知识库' }).click()
  await expect(page.getByText('最近浏览')).toBeVisible({ timeout: 15000 })

  await page.getByRole('button', { name: '新文档' }).click()
  await expect(page.locator('.title')).toBeVisible({ timeout: 15000 })
  await page.locator('.title').fill('E2E 文档')
  await page.locator('.md').fill('# 协同\n全文检索与版本回滚')
  await page.getByRole('button', { name: '保存版本' }).click()
  await page.getByPlaceholder('例如：发布前终稿').fill('E2E 快照')
  await page.getByRole('button', { name: '保存' }).click()
  await expect(page.getByText('E2E 快照')).toBeVisible({ timeout: 10000 })
})
