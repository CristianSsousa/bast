package cmd

import (
	"strconv"

	"github.com/CristianSsousa/go-bast-cli/internal/constants"
	"github.com/CristianSsousa/go-bast-cli/internal/serve"
	"github.com/spf13/cobra"
)

var (
	port     string
	host     string
	endpoint string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Inicia um servidor HTTP",
	Long: `Inicia um servidor HTTP na porta especificada.
Por padrão, o servidor roda na porta 8080.

Exemplos:
  bast serve                      # Inicia na porta 8080 (padrão)
  bast serve --port 3000         # Inicia na porta 3000
  bast serve -p 3000 -H localhost # Inicia na porta 3000 em localhost
  bast serve --help              # Mostra ajuda deste comando`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer(cmd)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&port, "port", "p", strconv.Itoa(constants.DefaultPort), "Porta do servidor")
	serveCmd.Flags().StringVarP(&host, "host", "H", constants.DefaultHost, "Host do servidor")
	serveCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "/", "Endpoint principal")
}

func startServer(cmd *cobra.Command) {
	verbosePrint(cmd, "Configurando servidor HTTP...\n")
	verbosePrint(cmd, "Host: %s\n", host)
	verbosePrint(cmd, "Porta: %s\n", port)
	verbosePrint(cmd, "Endpoint principal: %s\n", endpoint)

	r, w, i := serve.DefaultTimeouts()
	verbosePrint(cmd, "ReadTimeout: %v\n", r)
	verbosePrint(cmd, "WriteTimeout: %v\n", w)
	verbosePrint(cmd, "IdleTimeout: %v\n", i)
	verbosePrint(cmd, "Servidor pronto para receber conexões.\n")

	opts := serve.Options{
		Host:         host,
		Port:         port,
		Endpoint:     endpoint,
		ReadTimeout:  r,
		WriteTimeout: w,
		IdleTimeout:  i,
	}

	log := logForFeatures()
	if err := serve.Run(log, opts); err != nil {
		verbosePrint(cmd, "Erro ao iniciar servidor: %v\n", err)
		log.Fatal(err)
	}
}
