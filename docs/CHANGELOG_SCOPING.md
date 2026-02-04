# CHANGELOG - Scaffolding criado

**Data:** 04/02/2026
**Autor:** GitHub Copilot

## Arquivos/Diretórios adicionados

- `cmd/server/main.go` (placeholder)
- `config/env.go` (carrega .env)
- `internal/domain/admin/service/auth_service.go` (SecurityMiddleware + Oauth)
- `internal/domain/pgrs/models/*` (PGRS, VersaoPGRS, LogPGRS, Gerador, Unidade, Inventario, Vistoria, Auditoria, Intelligence)
- `internal/domain/comercial/models/proposta.go` (modelo)
- `internal/domain/financeiro/models/lancamento.go` (modelo)
- `internal/shared/models/base.go` (Base)
- `internal/shared/websocket/*` (client/message)
- `migrations/001_create_usuarios.sql`, `002_create_pgrs.sql` (esqueleto)
- `templates/README.md` (placeholder)
- `docs/PROJETO_EXECUTIVO.md` (conversão inicial)

## Observações

- Nenhuma lógica de negócio crítica foi alterada — apenas organizamos/encapsulamos modelos e criamos scaffolding. Alguns arquivos existentes continuam no lugar (por exemplo `main.go` é o entrypoint atual).
- Próximo passo: alinhar mais arquivos para refletir o projeto executivo (mover handlers, services, e testes por domínio).
