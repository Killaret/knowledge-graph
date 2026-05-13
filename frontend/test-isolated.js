// Изолированный запуск тестов Playwright без конфликта с Vitest
const { spawn } = require('child_process');
const path = require('path');

console.log('🧪 Запуск изолированных тестов Playwright...');

// Устанавливаем переменные окружения
const env = {
  ...process.env,
  SKIP_AUTH: 'true',
  FRONTEND_URL: 'http://localhost:5173',
  NODE_OPTIONS: '--no-experimental-fetch'
};

// Запускаем тесты в отдельном процессе
const testProcess = spawn('npx', ['playwright', 'test', '--config=playwright.config.docker.ts'], {
  cwd: path.join(__dirname),
  stdio: 'inherit',
  env: env,
  shell: true
});

testProcess.on('close', (code) => {
  console.log(`\n✅ Тесты завершены с кодом: ${code}`);
  process.exit(code);
});

testProcess.on('error', (error) => {
  console.error('❌ Ошибка запуска тестов:', error);
  process.exit(1);
});
