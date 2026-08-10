#!/bin/bash

# Script de comandos Git para registrar as etapas do projeto PioKe

# 1. Adicionar e commitar Fase 6 (Interface Gráfica Ebitengine)
git add pkg/ui/gui/
git commit -m "feat: adiciona interface gráfica 2D usando Ebitengine"

# 2. Adicionar e commitar o CHANGELOG e scripts de atualização
git add CHANGELOG.md tasks.md
git commit -m "docs: atualiza documentação do projeto e historico de tarefas"

echo "Commits concluídos com sucesso!"
