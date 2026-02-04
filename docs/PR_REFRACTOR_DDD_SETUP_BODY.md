# PR: refactor/ddd-setup

Resumo: scaffolding inicial conforme `docs/projeto executivo.txt`.

### Alterações principais

- Criado scaffolding de domínios: `internal/domain/pgrs`, `internal/domain/comercial`, `internal/domain/financeiro`, `internal/domain/admin/service/auth_service.go`
- Movidos/aliasados modelos para `internal/domain/*/models`
- Adicionado `templates/README.md`, `migrations/*` e docs de apoio
- Atualizado `main.go` para usar `internal/domain/admin/service` middleware

### Checklist

- [ ] Estrutura conforma `PROJETO_EXECUTIVO`
- [ ] Tests mínimos passam (`go test ./... -short`)
- [ ] Documentos `docs/PROJETO_EXECUTIVO.md` e `docs/CHANGELOG_SCOPING.md` incluídos

### Observações

PR inicial contém apenas scaffolding e pequenas mudanças para manter o sistema compilando. A migração de lógica Será feita em PRs subsequentes com alteração por domínio.

Link para criar PR: https://github.com/SerafimNunes/kpop-backend/pull/new/refactor/ddd-setup
