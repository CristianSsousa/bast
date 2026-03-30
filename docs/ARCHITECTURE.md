# Arquitetura do go-bast-cli

Documentação da organização do projeto **bast**: entrada, adaptador CLI e pacotes **por feature** em `internal/`.

## Visão geral

| Parte | Papel |
|-------|--------|
| [`main.go`](../main.go) | Chama apenas `cmd.Execute()`. |
| [`cmd/`](../cmd/) | Pacote Cobra: um arquivo por comando, flags e delegação para `internal/<feature>`. |
| [`internal/config/`](../internal/config/) | Viper, struct `Config`, estado global `Cfg` (compartilhado). |
| [`internal/logger/`](../internal/logger/) | Logrus (compartilhado). |
| [`internal/constants/`](../internal/constants/) | Constantes da aplicação. |
| [`internal/clienv/`](../internal/clienv/) | `Env` com `*config.Config` e `logrus.FieldLogger`, preenchido no `PersistentPreRun` do root. |
| [`internal/serve/`](../internal/serve/) | Feature **serve**: servidor HTTP com mux local, timeouts. |
| [`internal/update/`](../internal/update/) | Feature **update**: API GitHub, comparação de versões, `go install`. |
| [`internal/install/`](../internal/install/) | Feature **install**: instalação do Git por SO (exec). |
| [`pkg/utils/`](../pkg/utils/) | Helpers reutilizáveis (paths, arquivos) exportados pelo módulo. |

## Fluxo de bootstrap

1. `root` `PersistentPreRun`: `config.Init`, `logger.Init`, `clienv.Set(config.Get(), logger.GetLogger())`.
2. Comandos usam `configForFeatures()` e `logForFeatures()` em [`cmd/root.go`](../cmd/root.go) (leem `clienv.Current` com fallback seguro).
3. Não há logger global nomeado no pacote `cmd` além do padrão em `internal/logger`; mensagens de erro em `init()` usam `logger.GetLogger()` antes do bootstrap completo.

## Convenção por feature

- Cada comando com lógica relevante ganha um pacote `internal/<nome-da-feature>` (alinhado ao subcomando: `serve`, `update`, `install`).
- `internal/*` **não** importa `cmd` (evita ciclos).
- Novas ferramentas em `install` podem virar subpacotes (`internal/install/<tool>`) mantendo o dispatch no `cmd`.

## Testes

- Pacotes de feature testam IO com `httptest` / condicionais ao PATH (`git`, `go`) quando aplicável.
- Testes de integração do CLI permanecem em `cmd/*_test.go`.
