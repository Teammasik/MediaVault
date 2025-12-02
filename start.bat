@echo off
echo Starting Go-server...
start "Go Server" cmd /c "cd /d %~dp0 && go run main.go"

timeout /t 3 >nul

echo Starting Python UI...
start "Python UI" cmd /c "cd /d %~dp0UI && python app.py"

echo Press any button to stop the service...
pause

taskkill /f /im go.exe >nul 2>&1
taskkill /f /im python.exe >nul 2>&1

echo Application is closed.