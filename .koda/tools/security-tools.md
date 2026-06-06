# Инструменты Security Агента

## 🎯 Основные задачи

1. Security scanning
2. Аудит зависимостей
3. Проверка конфигураций
4. Auth/AuthZ hardening
5. Vulnerability management

---

## 🛠️ Инструменты

### 1. Dependency security
- `trivy` — сканирование контейнеров и файловой системы
- `gosec` — статический анализ Go кода
- `checkov` — проверка IaC и YAML
- `dependency-check` — анализ CVE для зависимостей

### 2. Static analysis
- `gosec` — Go security linting
- `staticcheck` — дополнительные проверки кода
- `ody` / `tfsec` — проверка Terraform, Docker и Kubernetes

### 3. Runtime security
- `Open Policy Agent` (OPA) — политика доступа
- `Falco` — runtime monitoring
- `Auditd` / logging security events

---

## 🧩 Практики

- Не выводить секреты в логах
- Использовать least privilege
- Шифровать данные в покое и в пути
- Проверять все входные данные
- Ручной review для критичных изменений

---

## ✅ Шаблоны

### Пример сканирования контейнера
```bash
trivy image --severity HIGH,CRITICAL --exit-code 1 --no-progress myapp:latest
```

### Пример запуска gosec
```bash
gosec ./...
```

### Пример проверки Terraform
```bash
checkov -d .
```
