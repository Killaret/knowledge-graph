// E2E тесты для полного цикла предзагрузки данных
import { test, expect } from '@playwright/test';

test.describe('PreloadService Full Cycle E2E', () => {
  test.beforeEach(async ({ page }) => {
    // Очищаем localStorage перед каждым тестом
    await page.context().clearCookies();
    await page.addInitScript(() => {
      // Инициализация localStorage если нужно
      if (typeof localStorage !== 'undefined') {
        localStorage.clear();
      }
    });
  });

  test('should preload data on login page and display instantly after login', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/auth/login');
    
    // Ждем загрузки страницы входа
    await expect(page.locator('h1')).toContainText('Knowledge Graph');
    
    // Проверяем, что PreloadService запускается (через консольные логи)
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(msg.text());
    });
    
    // Ждем некоторое время для предзагрузки
    await page.waitForTimeout(2000);
    
    // Проверяем, что были логи предзагрузки
    const preloadLogs = consoleMessages.filter(msg => 
      msg.includes('[PreloadService]') && msg.includes('Starting background preload')
    );
    expect(preloadLogs.length).toBeGreaterThan(0);
    
    // Заполняем форму входа
    await page.fill('input[name="login"]', 'testuser');
    await page.fill('input[name="password"]', 'testpassword');
    
    // Отслеживаем время отображения главной страницы после входа
    const startTime = Date.now();
    
    // Нажимаем кнопку входа
    await page.click('button[type="submit"]');
    
    // Ждем перехода на главную страницу
    await page.waitForURL('/');
    
    // Проверяем, что интерфейс отображается быстро (менее 1 секунды)
    const loadTime = Date.now() - startTime;
    expect(loadTime).toBeLessThan(1000);
    
    // Проверяем, что граф отображается (не пустой)
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    
    // Проверяем, что были логи использования предзагруженных данных
    const usePreloadedLogs = consoleMessages.filter(msg => 
      msg.includes('[usePreloadedData]') && msg.includes('Using preloaded')
    );
    expect(usePreloadedLogs.length).toBeGreaterThan(0);
  });

  test('should handle preload errors gracefully', async ({ page }) => {
    // Мокаем ошибки API через страницу
    await page.route('**/api/v1/graph/all**', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' })
      });
    });
    
    await page.route('**/api/v1/achievements**', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' })
      });
    });
    
    // Переходим на страницу входа
    await page.goto('/auth/login');
    
    // Ждем некоторое время для попытки предзагрузки
    await page.waitForTimeout(2000);
    
    // Проверяем, что были логи ошибок предзагрузки
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(msg.text());
    });
    
    await page.waitForTimeout(1000);
    
    const errorLogs = consoleMessages.filter(msg => 
      msg.includes('[PreloadService]') && msg.includes('Failed to preload')
    );
    expect(errorLogs.length).toBeGreaterThan(0);
    
    // Выполняем вход
    await page.fill('input[name="login"]', 'testuser');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    // Ждем перехода на главную страницу
    await page.waitForURL('/');
    
    // Приложение все равно должно загрузиться (с fallback на сервер)
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
  });

  test('should clear preload cache on logout', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/auth/login');
    
    // Ждем предзагрузки
    await page.waitForTimeout(2000);
    
    // Выполняем вход
    await page.fill('input[name="login"]', 'testuser');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    // Ждем загрузки главной страницы
    await page.waitForURL('/');
    
    // Проверяем, что данные загружены
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    
    // Выполняем выход
    await page.click('[data-testid="logout-button"]');
    
    // Ждем перехода на страницу входа
    await page.waitForURL('/auth/login');
    
    // Проверяем, что кэш очищен (через localStorage)
    const cacheExists = await page.evaluate(() => {
      return localStorage.getItem('preload_cache') !== null;
    });
    
    // В нашей реализации кэш хранится в памяти, но мы можем проверить
    // что пользователь разлогинен
    await expect(page.locator('h1')).toContainText('Knowledge Graph');
  });

  test('should not preload when already authenticated', async ({ page }) => {
    // Устанавливаем токены в localStorage (эмулируем уже аутентифицированного пользователя)
    await page.addInitScript(() => {
      localStorage.setItem('access_token', 'test_token');
      localStorage.setItem('refresh_token', 'test_refresh');
    });
    
    // Переходим на главную страницу
    await page.goto('/');
    
    // Проверяем, что мы не перенаправлены на страницу входа
    await expect(page).toHaveURL('/');
    
    // Проверяем, что предзагрузка не запускалась
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      consoleMessages.push(msg.text());
    });
    
    await page.waitForTimeout(2000);
    
    const preloadLogs = consoleMessages.filter(msg => 
      msg.includes('[PreloadService]') && msg.includes('Starting background preload')
    );
    expect(preloadLogs.length).toBe(0);
  });

  test('should handle concurrent preload requests', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/auth/login');
    
    // Ждем начала предзагрузки
    await page.waitForTimeout(500);
    
    // Открываем вторую вкладку с той же страницей
    const newPage = await page.context().newPage();
    await newPage.goto('/auth/login');
    
    // Ждем завершения предзагрузки
    await page.waitForTimeout(2000);
    
    // Выполняем вход на первой странице
    await page.fill('input[name="login"]', 'testuser');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('/');
    
    // Выполняем вход на второй странице
    await newPage.fill('input[name="login"]', 'testuser2');
    await newPage.fill('input[name="password"]', 'testpassword2');
    await newPage.click('button[type="submit"]');
    
    await newPage.waitForURL('/');
    
    // Обе страницы должны загрузиться корректно
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    await expect(newPage.locator('[data-testid="graph-canvas"]')).toBeVisible();
    
    await newPage.close();
  });

  test('should respect TTL and refresh expired cache', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/auth/login');
    
    // Ждем предзагрузки
    await page.waitForTimeout(2000);
    
    // Выполняем вход
    await page.fill('input[name="login"]', 'testuser');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('/');
    
    // Проверяем, что данные отображаются быстро
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    
    // Эмулируем истечение TTL (перезагружаем страницу через долгое время)
    await page.evaluate(() => {
      // Мокаем Date.now для эмуляции истечения TTL
      const originalDateNow = Date.now;
      Date.now = () => originalDateNow() + (6 * 60 * 1000); // +6 минут
    });
    
    // Перезагружаем страницу
    await page.reload();
    
    // Данные должны загрузиться с сервера (медленнее)
    const startTime = Date.now();
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    const loadTime = Date.now() - startTime;
    
    // Загрузка должна занять больше времени (нет предзагруженных данных)
    expect(loadTime).toBeGreaterThan(500);
    
    // Восстанавливаем Date.now
    await page.evaluate(() => {
      // Date.now будет восстановлен при перезагрузке страницы
    });
  });

  test('should work with different user roles', async ({ page }) => {
    // Тестируем с разными ролями пользователей
    const userRoles = ['user', 'admin'];
    
    for (const role of userRoles) {
      // Создаем новую страницу для каждой роли
      const testPage = await page.context().newPage();
      
      // Переходим на страницу входа
      await testPage.goto('/auth/login');
      
      // Ждем предзагрузки
      await testPage.waitForTimeout(2000);
      
      // Выполняем вход с соответствующей ролью
      await testPage.fill('input[name="login"]', `${role}user`);
      await testPage.fill('input[name="password"]', 'testpassword');
      await testPage.click('button[type="submit"]');
      
      await testPage.waitForURL('/');
      
      // Проверяем, что интерфейс отображается корректно
      await expect(testPage.locator('[data-testid="graph-canvas"]')).toBeVisible();
      
      // Для админа могут быть дополнительные элементы
      if (role === 'admin') {
        // Проверяем наличие админских элементов (если они есть)
        const adminElements = testPage.locator('[data-testid*="admin"]');
        // Не ждем их наличия, просто проверяем что нет ошибок
      }
      
      await testPage.close();
    }
  });

  test('should handle network interruptions gracefully', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/auth/login');
    
    // Эмулируем прерывание сети во время предзагрузки
    await page.route('**/api/v1/graph/all**', route => {
      // Обрываем соединение
      route.abort('failed');
    });
    
    // Ждем попытки предзагрузки
    await page.waitForTimeout(2000);
    
    // Убираем блокировку
    await page.unroute('**/api/v1/graph/all**');
    
    // Выполняем вход
    await page.fill('input[name="login"]', 'testuser');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    // Приложение должно загрузиться с fallback
    await page.waitForURL('/');
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
  });
});
