package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostra a versão da aplicação",
	Long: `Mostra a versão atual da aplicação bast.

		Exemplos:
		  bast version        # Mostra a versão
		  bast version --help # Mostra ajuda deste comando`,
	Run: func(cmd *cobra.Command, args []string) {
		verbosePrint(cmd, "Obtendo informações de versão...")
		cfg := configForFeatures()
		fmt.Printf("%s v%s\n", cfg.App.Name, cfg.App.Version)
		fmt.Println("Construído com Go e Cobra")
		fmt.Printf("Autor: %s\n", cfg.App.Author)
		verbosePrint(cmd, "Informações de versão exibidas.")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
