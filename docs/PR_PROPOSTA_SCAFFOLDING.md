# PR Proposta: Scaffolding conforme PROJETO_EXECUTIVO

**Versão:** 1.0
**Autor:** GitHub Copilot
**Data:** 04/02/2026

---

## 🔎 Objetivo

Criar a estrutura de diretórios e arquivos _scaffold_ exatamente conforme o `PROJETO_EXECUTIVO` (docs/projeto executivo.txt) para que a base do repositório fique congruente com o projeto executivo antes de iniciar o trabalho de refatoração e implementação.

> Este PR é propositalmente _apenas scaffolding_: arquivos conterão comentários explicativos, TODOs e testes mínimos (quando aplicável). Nenhuma lógica de domínio robusta será implementada neste PR.

---

## ✅ Itens que serão adicionados (lista resumida)

(Arquivos serão criados apenas após sua aprovação)

- Criar diretórios:
  - `cmd/server/`
  - `internal/domain/admin/{models,repository,service,handlers,dto}`
  - `internal/domain/pgrs/{models,repository,service,handlers,dto,agents}`
  - `internal/domain/comercial/{...}`
  - `internal/domain/financeiro/{...}`
  - `internal/shared/{database,websocket,middleware,validator,utils,errors}`
  - `migrations/`
  - `templates/`

- Criar arquivos placeholder (com comentários/TODOs):
  - `cmd/server/main.go` (entrypoint limpo)
  - `config/config.go`, `config/env.go`
  - `internal/shared/database/connection.go`, `internal/shared/database/migrations.go`
  - `internal/shared/websocket/hub.go`, `internal/shared/websocket/client.go`, `internal/shared/websocket/message.go`
  - `internal/shared/validator/cnpj.go`, `internal/shared/validator/email.go`, `internal/shared/validator/validator.go` (com testes mínimos)
  - `internal/domain/admin/models/usuario.go`, `internal/domain/admin/repository/interfaces.go`, `internal/domain/admin/service/auth_service.go`, `internal/domain/admin/handlers/http_handler.go`, `internal/domain/admin/dto/login_request.go`
  - `internal/domain/pgrs/models/pgrs.go`, `internal/domain/pgrs/repository/interfaces.go`, `internal/domain/pgrs/service/pgrs_service.go`, `internal/domain/pgrs/handlers/http_handler.go`, `internal/domain/pgrs/dto/requests.go`
  - `migrations/001_create_usuarios.sql`, `migrations/002_create_pgrs.sql` (esqueleto)
  - `templates/pgrs_template.docx` (placeholder README explicando template esperado)
  - `docs/PROJETO_EXECUTIVO.md` (versão Markdown do txt existente)

- Adicionar `docs/CHANGELOG_SCOPING.md` descrevendo os arquivos criados e motivação.

---

## 🧪 Testes e CI

- Incluir testes mínimos (unitários) para validação (`validator`), e testes de smoke para `internal/shared/websocket` hub.
- Criar `.github/workflows/ci-scaffold.yml` que executa `go vet`, `go test ./... -short` e `golangci-lint run` (opcional, se preferir adiciono lint config depois).

---

## 📋 Checklist do PR (para revisão antes de merge)

- [ ] Estrutura de diretórios e arquivos criada conforme lista proposta
- [ ] Arquivos contém comentários/TODOs com next steps claros
- [ ] Tests mínimos inclusos e passando (`go test ./... -short`)
- [ ] Documentação (`docs/PROJETO_EXECUTIVO.md`, `docs/CHANGELOG_SCOPING.md`) atualizada
- [ ] Nenhuma lógica de negócio crítica implementada (apenas placeholders)

---

## ✳️ Próximos passos após sua aprovação

1. Criar branch `refactor/ddd-setup` e commitar scaffolding.
2. Abrir PR com descrição, checklist e solicitar sua revisão.
3. Após merge, iniciar Fase 1 de migração (mover models, extrair DB, inicializar Admin domain como piloto).

---

## ❓ Perguntas para você

- Deseja que eu inclua `golangci-lint` e um job de lint no PR-scaffold ou prefere adicionar isso depois (em PR separado)?
- Prefere que eu gere os arquivos SQL de migrations completos (com tipos e constraints) desde já ou apenas os esboços para começarmos?

---

Se concorda com o plano, responda com **"Aprovado"** e eu criarei a branch `refactor/ddd-setup` e o PR-proposta com os arquivos scaffold. Se quiser ajustar a lista, indique as mudanças e eu atualizo o documento antes de criar qualquer arquivo.
