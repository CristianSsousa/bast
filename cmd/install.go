package cmd

import (
	"fmt"
	"os"

	"github.com/CristianSsousa/go-bast-cli/internal/install"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala ferramentas e dependências",
	Long: `Instala ferramentas e dependências necessárias.
Atualmente suporta instalação do Git em diferentes sistemas operacionais.

Exemplos:
  bast install git              # Instala o Git
  bast install --help           # Mostra ajuda deste comando`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Erro: especifique o que deseja instalar.")
			fmt.Println("Uso: bast install <ferramenta>")
			fmt.Println("\nFerramentas disponíveis:")
			fmt.Println("  git - Instala o Git")
			os.Exit(1)
		}

		tool := args[0]
		switch tool {
		case "git":
			if err := install.InstallGit(logForFeatures()); err != nil {
				fmt.Printf("\nErro ao executar instalação: %v\n", err)
				printGitManualInstallHelp()
				os.Exit(1)
			}
		default:
			fmt.Printf("Erro: ferramenta '%s' não é suportada.\n", tool)
			fmt.Println("\nFerramentas disponíveis:")
			fmt.Println("  git - Instala o Git")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func printGitManualInstallHelp() {
	fmt.Println("\nTente instalar manualmente:")
	fmt.Println("  Windows: https://git-scm.com/download/win")
	fmt.Println("  Linux: Use o gerenciador de pacotes da sua distribuição")
	fmt.Println("  macOS: brew install git")
}
