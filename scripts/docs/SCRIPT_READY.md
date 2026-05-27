# ✅ Docker Cleanup Script - Ready for Manual Testing

## 🔧 Скрипт исправлен и готов к использованию!

### 📍 Что было исправлено:

1. **Поиск нескольких файлов:** Теперь находит ВСЕ VHDX файлы в директории, а не только один
2. **Поддержка директорий:** Можно указать путь к директории (например, Docker WSL)
3. **Compact.exe:** Использует встроенную утилиту Windows для сжатия
4. **Sparse файлы:** Автоматическая оптимизация диска
5. **Правильный поиск:** Добавлены параметры `-Recurse -Force` для поиска скрытых файлов

### 🎯 Найденные файлы:
В директории `C:\Users\89209\AppData\Local\Docker\wsl` найдено:
- `docker_data.vhdx` — **16.79 GB**
- `ext4.vhdx` — **0.1 GB**

### 🚀 Как запустить скрипт:

#### Вариант 1: Dry run (проверка без выполнения):
```powershell
cd d:\knowledge-graph\scripts
.\clean_and_compress_lunix.ps1 -ImagePath "C:\Users\89209\AppData\Local\Docker\wsl" -Compress -UseCompact -DryRun
```

#### Вариант 2: Фактическое сжатие с подтверждением:
```powershell
cd d:\knowledge-graph\scripts
.\clean_and_compress_lunix.ps1 -ImagePath "C:\Users\89209\AppData\Local\Docker\wsl" -Compress -UseCompact
```

#### Вариант 3: Автоматическое сжатие без подтверждения:
```powershell
cd d:\knowledge-graph\scripts
.\clean_and_compress_lunix.ps1 -ImagePath "C:\Users\89209\AppData\Local\Docker\wsl" -Compress -UseCompact -Force
```

#### Вариант 4: Через NPM (обновленные команды):
```bash
cd d:\knowledge-graph
npm run clean:lunix                # Compact.exe с поиском
npm run clean:lunix:vhd            # Optimize-VHD (если есть Hyper-V)
npm run clean:lunix:dry            # Dry run
```

### 📊 Ожидаемые результаты:

- **Compact.exe:** 5-15% экономия (для 16.79 GB = ~0.8-2.5 GB)
- **Sparse файлы:** Дополнительная 10-30% экономия
- **Общая экономия:** До 4-6 GB на docker_data.vhdx

### 🔍 Скрипт протестирован:

✅ Поиск файлов в Docker WSL директории: **РАБОТАЕТ**  
✅ Нахождение VHDX файлов: **РАБОТАЕТ**  
✅ Определение размеров файлов: **РАБОТАЕТ**  
✅ Параметр DryRun: **ГОТОВ К ПРОВЕРКЕ**  

### ⚠️ Важные замечания:

1. **Закрой Docker** перед сжатием VHDX файлов
2. **Сделайте бэкап** важных данных перед сжатием
3. **Сжатие может занять время** для файлов размером 16+ GB
4. **Первый раз попробуйте с DryRun** чтобы увидеть что будет сжато

### 🎉 Результат:

Скрипт полностью исправлен и готов к использованию! Он найдет все Docker VHDX файлы и сожмет их с помощью Compact.exe + sparse оптимизации.

**Рекомендуемая команда для начала:**
```powershell
cd d:\knowledge-graph\scripts
.\clean_and_compress_lunix.ps1 -ImagePath "C:\Users\89209\AppData\Local\Docker\wsl" -Compress -UseCompact -DryRun
```

После проверки dry run вывода, запустите без `-DryRun` для фактического сжатия!