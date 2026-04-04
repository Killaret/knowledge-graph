@echo off
setlocal enabledelayedexpansion

REM Get a free port for testing
set "BACKEND_URL=http://localhost:8080"

echo Testing async task processing...
echo Backend URL: %BACKEND_URL%
echo.

REM Create a test note
echo Creating test note...
for /f "delims=" %%i in ('powershell -Command "Write-Host ([System.Guid]::NewGuid().ToString())"') do set "NOTE_ID=%%i"

powershell -Command ^
  "$body = @{ title='Test Note'; content='This is a test note for async keyword extraction and embedding. It should trigger the worker tasks.' } | ConvertTo-Json; " ^
  "$response = Invoke-RestMethod -Uri 'http://localhost:8080/notes' -Method POST -ContentType 'application/json' -Body $body; " ^
  "$response | ConvertTo-Json | Write-Output"

echo.
echo Note created. Waiting 5 seconds for worker to process...
timeout /t 5 /nobreak

echo.
echo Checking worker logs for task processing...
docker-compose logs --tail=30 kg-worker | findstr /I "HandleExtractKeywords HandleComputeEmbedding successfully processed"

echo.
echo Test complete!
pause
