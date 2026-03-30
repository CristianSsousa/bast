package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/CristianSsousa/go-bast-cli/internal/update"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Atualiza o bast CLI para a versão mais recente",
	Long: `Verifica e atualiza o bast CLI para a versão mais recente disponível no GitHub.

O comando verifica a versão atual instalada e compara com a versão mais recente
disponível no GitHub. Se houver uma atualização disponível, o comando oferece
para atualizar usando 'go install'.

Exemplos:
  bast update              # Verifica e atualiza para a versão mais recente
  bast update --check      # Apenas verifica se há atualização disponível
  bast update --help       # Mostra ajuda deste comando`,
	Run: func(cmd *cobra.Command, args []string) {
		runUpdate(cmd)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("check", "c", false, "Apenas verifica se há atualização disponível, sem atualizar")
}

func runUpdate(cmd *cobra.Command) {
	checkOnly, err := cmd.Flags().GetBool("check")
	if err != nil {
		verbosePrint(cmd, "Erro ao obter flag 'check': %v", err)
		checkOnly = false
	}

	verbosePrint(cmd, "Iniciando verificação de atualização...")
	cfg := configForFeatures()

	fmt.Printf("Versão atual: %s\n", cfg.App.Version)
	fmt.Println("Verificando atualizações disponíveis...")

	client := &http.Client{Timeout: update.DefaultHTTPTimeout}
	latestRelease, err := update.FetchLatest(client, update.GitHubAPIURL)
	if err != nil {
		verbosePrint(cmd, "Erro ao obter release: %v", err)
		fmt.Printf("Erro ao verificar atualizações: %v\n", err)
		fmt.Println("\nDica: Verifique sua conexão com a internet.")
		os.Exit(1)
	}

	latestVersion := update.TrimVersionPrefix(latestRelease.TagName)
	currentVersion := cfg.App.Version

	verbosePrint(cmd, "Versão mais recente encontrada: %s", latestVersion)
	verbosePrint(cmd, "Versão atual: %s", currentVersion)

	if update.IsUpToDate(currentVersion, latestVersion) {
		fmt.Printf("Você já está usando a versão mais recente (%s)!\n", currentVersion)
		return
	}

	fmt.Printf("Nova versão disponível: %s\n", latestVersion)
	if latestRelease.Name != "" {
		fmt.Printf("Release: %s\n", latestRelease.Name)
	}

	if checkOnly {
		fmt.Println("\nPara atualizar, execute: bast update")
		return
	}

	fmt.Print("\nDeseja atualizar agora? [y/n]: ")
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil {
		verbosePrint(cmd, "Erro ao ler entrada do usuário: %v", err)
		return
	}

	responseLower := strings.ToLower(response)
	if !strings.EqualFold(responseLower, "s") && !strings.EqualFold(responseLower, "sim") &&
		!strings.EqualFold(responseLower, "y") && !strings.EqualFold(responseLower, "yes") {
		fmt.Println("Atualização cancelada.")
		return
	}

	fmt.Println("\nAtualizando bast CLI...")
	verbosePrint(cmd, "Executando: go install %s@%s", update.ModulePath, latestRelease.TagName)

	tag := latestRelease.TagName
	if !update.ValidTag(tag) {
		fmt.Fprintf(os.Stderr, "Erro: tag de versão inválida: %s\n", tag)
		os.Exit(1)
	}

	if err := update.RunGoInstall(tag, os.Stdout, os.Stderr); err != nil {
		verbosePrint(cmd, "Erro ao executar go install: %v", err)
		fmt.Printf("\nErro ao atualizar: %v\n", err)
		fmt.Println("\nTente atualizar manualmente:")
		fmt.Printf("  go install %s@%s\n", update.ModulePath, latestRelease.TagName)
		os.Exit(1)
	}

	fmt.Println("\nAtualização concluída com sucesso!")
	fmt.Printf("Versão instalada: %s\n", latestVersion)
	fmt.Println("\nNota: Se o comando 'bast' não refletir a nova versão, certifique-se de que")
	fmt.Println("o diretório $GOPATH/bin ou $GOBIN está no seu PATH.")
}
