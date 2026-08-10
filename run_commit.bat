@echo off
REM Script em Batch para executar os commits Git no Windows

echo Executando commits do projeto PioKe...

git add pkg/ui/gui/
git commit -m "feat: adiciona interface gráfica 2D usando Ebitengine"

git add CHANGELOG.md tasks.md
git commit -m "docs: atualiza documentacao do projeto e historico de tarefas"

echo Commits concluidos com sucesso!
pause
