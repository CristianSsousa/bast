package clienv

import (
	"github.com/CristianSsousa/go-bast-cli/internal/config"
	"github.com/sirupsen/logrus"
)

// Env agrupa dependências compartilhadas entre comandos após o bootstrap em PersistentPreRun.
type Env struct {
	Config *config.Config
	Log    logrus.FieldLogger
}

// Current é preenchido no root PersistentPreRun após config.Init e logger.Init.
var Current *Env

// Set atualiza o ambiente atual (config + logger) para consumo das features.
func Set(cfg *config.Config, log logrus.FieldLogger) {
	Current = &Env{Config: cfg, Log: log}
}
