@echo off
REM Script em Batch para executar os commits Git no Windows

echo Executando commits do projeto PioKe...

git add -A
git commit -m "chore: salva alteracoes pendentes de configuracao e documentacao"

echo Commits concluidos com sucesso!
pause
