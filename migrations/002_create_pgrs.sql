-- 002_create_pgrs.sql
-- Esqueleto da tabela PGRS (ajustar colunas conforme modelo final)

CREATE TABLE IF NOT EXISTS pgrs (
    id SERIAL PRIMARY KEY,
    numero VARCHAR(100) NOT NULL UNIQUE,
    unidade_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    dados_formulario JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
