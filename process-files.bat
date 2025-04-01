@echo off
setlocal

echo Enviando requisição para processar arquivos da pasta data/raw...

curl -X POST http://localhost:8080/process-files ^
  -H "Copilot-Integration-Id: local-dev" ^
  -H "X-GitHub-Token: local-fake-token"

echo.
echo Requisição enviada. Verifique o terminal do servidor para os logs.
pause
endlocal