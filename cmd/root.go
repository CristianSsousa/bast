package cmd

import (
	"fmt"
	"os"

	"github.com/CristianSsousa/go-bast-cli/internal/clienv"
	"github.com/CristianSsousa/go-bast-cli/internal/config"
	"github.com/CristianSsousa/go-bast-cli/internal/constants"
	"github.com/CristianSsousa/go-bast-cli/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// isVerbose verifica se o modo verbose está ativado
func isVerbose(cmd *cobra.Command) bool {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return false
	}
	return verbose
}

// verbosePrint imprime mensagens apenas se o modo verbose estiver ativo usando logger estruturado
func verbosePrint(cmd *cobra.Command, format string, args ...interface{}) {
	if isVerbose(cmd) {
		logForFeatures().Debugf(format, args...)
	}
}

// logForFeatures retorna o logger após clienv.Set (fallback para o logger global).
func logForFeatures() logrus.FieldLogger {
	if clienv.Current != nil && clienv.Current.Log != nil {
		return clienv.Current.Log
	}
	return logger.GetLogger()
}

// configForFeatures retorna a config após clienv.Set (fallback para config.Get).
func configForFeatures() *config.Config {
	if clienv.Current != nil && clienv.Current.Config != nil {
		return clienv.Current.Config
	}
	return config.Get()
}

var rootCmd = &cobra.Command{
	Use:   "bast",
	Short: "Uma CLI moderna construída com Go e Cobra",
	Long: `bast é uma aplicação CLI moderna construída com Go e Cobra.
Ela fornece uma interface de linha de comando poderosa e extensível.

Exemplos:
  bast version                    # Mostra a versão
  bast greet --name "João"        # Cumprimenta alguém
  bast serve --port 3000          # Inicia servidor na porta 3000
  bast install git                # Instala o Git
  bast info                       # Mostra informações do sistema
  bast port 8080                  # Verifica se porta está em uso
  bast config list                # Lista configurações
  bast update                     # Atualiza o CLI para a versão mais recente
  bast --help                     # Mostra esta mensagem de ajuda`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := config.Init(cfgFile); err != nil {
			logger.GetLogger().Warnf("Erro ao carregar configuração: %v", err)
		}

		cfg := config.Get()
		logger.Init(cfg.Logging.Level, cfg.Logging.Format)

		log := logger.GetLogger()
		clienv.Set(config.Get(), log)

		if isVerbose(cmd) {
			log.SetLevel(logrus.DebugLevel)
			log.Debug("Modo verbose ativado")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		verbosePrint(cmd, "Executando comando raiz...")
		cfg := configForFeatures()
		fmt.Printf(constants.WelcomeMessage+"\n", cfg.App.Name)
		fmt.Printf(constants.HelpMessage+"\n", cfg.App.Name)
		verbosePrint(cmd, "Modo verbose está ativo.")
	},
}

// Execute adiciona todos os comandos filhos ao comando raiz e define flags apropriadas.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logForFeatures().Errorf("Erro ao executar comando: %v", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Flags globais
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Modo verboso")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Arquivo de configuração (padrão: ~/.bast/config.yaml)")

	// Bind flags ao Viper
	if err := viper.BindPFlag("features.verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		logger.GetLogger().Warnf("Erro ao vincular flag verbose: %v", err)
	}
}
