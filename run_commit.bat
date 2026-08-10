@echo off
REM Script em Batch para executar os commits Git no Windows

echo Executando commits do projeto PioKe...

git add pkg/ cmd/ songs/ go.mod go.sum README.md tasks.md CHANGELOG.md .gitignore run_commit.bat run_commit.sh
git commit -m "chore: salva alteracoes pendentes de configuracao e documentacao"

echo Commits concluidos com sucesso!
pause
