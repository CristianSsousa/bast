package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/CristianSsousa/go-bast-cli/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	// protectedEnvVars lista de variáveis sensíveis que não devem ser modificadas ou deletadas
	// sem confirmação explícita usando --force
	protectedEnvVars = []string{
		// Variáveis críticas do sistema
		"PATH", "HOME", "USER", "USERNAME", "TEMP", "TMP",
		// Windows
		"WINDIR", "SYSTEMROOT", "COMSPEC", "PATHEXT",
		"PROCESSOR_ARCHITECTURE", "PROCESSOR_IDENTIFIER",
		"OS", "COMPUTERNAME", "HOMEDRIVE", "HOMEPATH",
		"APPDATA", "LOCALAPPDATA", "PROGRAMFILES",
		"PROGRAMFILES(X86)", "COMMONPROGRAMFILES",
		// Linux/macOS
		"SHELL", "PWD", "LANG", "LC_ALL", "DISPLAY",
		"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME",
		// Variáveis de desenvolvimento comuns
		"GOROOT", "GOPATH", "GOBIN",
		"JAVA_HOME", "ANDROID_HOME",
		"NODE_PATH", "NVM_DIR",
	}

	// appendOnlyEnvVars lista de variáveis protegidas que permitem apenas append (não substituição)
	// Essas variáveis só podem ser modificadas usando --append, nunca com --force
	appendOnlyEnvVars = []string{
		"PATH", "LD_LIBRARY_PATH", "DYLD_LIBRARY_PATH",
		"PYTHONPATH", "NODE_PATH",
	}
)

var environmentCmd = &cobra.Command{
	Use:   "env",
	Short: "Gerencia variáveis de ambiente do usuário",
	Long: `Gerencia variáveis de ambiente do usuário atual.

Permite listar, definir e deletar variáveis de ambiente do usuário atual.
Todas as operações são feitas apenas no escopo do usuário, não afetam variáveis do sistema.

Exemplos:
  bast env --list                    # Lista todas as variáveis de ambiente
  bast env --get --key PATH          # Busca uma variável específica pela chave
  bast env --set --key MINHA_VAR --value "meu valor"  # Define uma nova variável do usuário
  bast env --set --key PATH --value "C:\\bin" --append  # Adiciona ao PATH do usuário
  bast env --set --key PATH --value "C:\\bin" --force    # Substitui valor existente
  bast env --delete --key MINHA_VAR  # Deleta uma variável de ambiente do usuário
  bast env --help                    # Mostra ajuda deste comando`,
	RunE: func(cmd *cobra.Command, args []string) error {
		listFlag, err := cmd.Flags().GetBool("list")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'list': %w", err)
		}
		getFlag, err := cmd.Flags().GetBool("get")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'get': %w", err)
		}
		setFlag, err := cmd.Flags().GetBool("set")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'set': %w", err)
		}
		deleteFlag, err := cmd.Flags().GetBool("delete")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'delete': %w", err)
		}
		keyFlag, err := cmd.Flags().GetString("key")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'key': %w", err)
		}
		valueFlag, err := cmd.Flags().GetString("value")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'value': %w", err)
		}
		appendFlag, err := cmd.Flags().GetBool("append")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'append': %w", err)
		}
		forceFlag, err := cmd.Flags().GetBool("force")
		if err != nil {
			return fmt.Errorf("erro ao obter flag 'force': %w", err)
		}

		// Validação: pelo menos uma ação deve ser especificada
		if !listFlag && !getFlag && !setFlag && !deleteFlag {
			if err := cmd.Usage(); err != nil {
				verbosePrint(cmd, "Erro ao exibir uso: %v", err)
			}
			return fmt.Errorf("erro: nenhuma ação especificada\n\nUse 'bast env --list' para listar " +
				"variáveis\nUse 'bast env --get --key NOME' para buscar uma variável\n" +
				"Use 'bast env --set --key NOME --value VALOR' para definir uma variável\n" +
				"Use 'bast env --delete --key NOME' para deletar uma variável")
		}

		verbosePrint(cmd, "Iniciando gerenciamento de variáveis de ambiente...")
		osType := utils.GetOS()
		verbosePrint(cmd, "Sistema operacional: %s", osType)

		if listFlag {
			return showListOfEnvironments(cmd)
		}

		if getFlag {
			if keyFlag == "" {
				return fmt.Errorf("erro: é necessário especificar --key para buscar\n\nExemplo: bast env --get --key PATH")
			}

			return getEnvironmentVariable(cmd, keyFlag)
		}

		if setFlag {
			if keyFlag == "" || valueFlag == "" {
				return fmt.Errorf("erro: é necessário especificar --key e --value\n\nExemplo: bast env --set --key MINHA_VAR --value 'meu valor'")
			}

			return setEnvironmentVariable(cmd, osType, keyFlag, valueFlag, appendFlag, forceFlag)
		}

		if deleteFlag {
			if keyFlag == "" {
				return fmt.Errorf("erro: é necessário especificar --key para deletar\n\nExemplo: bast env --delete --key MINHA_VAR")
			}

			return deleteEnvironmentVariable(cmd, osType, keyFlag)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(environmentCmd)

	environmentCmd.Flags().BoolP("list", "l", false, "Lista todas as variáveis de ambiente do sistema")
	environmentCmd.Flags().BoolP("get", "g", false, "Busca uma variável de ambiente específica pela chave")
	environmentCmd.Flags().BoolP("set", "s", false, "Define uma nova variável de ambiente")
	environmentCmd.Flags().BoolP("delete", "d", false, "Deleta uma variável de ambiente")
	environmentCmd.Flags().StringP("key", "k", "", "Nome da variável de ambiente (obrigatório com --get, --set ou --delete)")
	environmentCmd.Flags().String("value", "", "Valor da variável de ambiente (obrigatório com --set)")
	environmentCmd.Flags().BoolP("append", "a", false, "Adiciona ao valor existente (útil para PATH)")
	environmentCmd.Flags().BoolP("force", "f", false, "Força substituição mesmo se a variável já existir")
}

// showListOfEnvironments lista todas as variáveis de ambiente
func showListOfEnvironments(cmd *cobra.Command) error {
	fmt.Println("Lista de Variáveis de Ambiente:")
	fmt.Println(strings.Repeat("-", 80))

	// Usar Go nativo para melhor performance e compatibilidade
	envVars := os.Environ()

	// Filtrar variáveis internas do sistema que não são relevantes
	filteredVars := make([]string, 0, len(envVars))
	for _, env := range envVars {
		// Ignorar variáveis que começam com _ (variáveis internas do shell)
		if strings.HasPrefix(env, "_=") {
			continue
		}
		filteredVars = append(filteredVars, env)
	}

	for _, env := range filteredVars {
		fmt.Println(env)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d variáveis\n", len(filteredVars))
	verbosePrint(cmd, "Listagem concluída com sucesso")

	return nil
}

// getEnvironmentVariable busca uma variável de ambiente específica pela chave
func getEnvironmentVariable(cmd *cobra.Command, key string) error {
	verbosePrint(cmd, "Buscando variável de ambiente: %s", key)

	value := os.Getenv(key)
	if value == "" {
		return fmt.Errorf("variável '%s' não encontrada ou não está definida", key)
	}

	fmt.Printf("%s=%s\n", key, value)
	verbosePrint(cmd, "Variável encontrada: %s = %s", key, value)

	return nil
}

// isProtectedVariable verifica se uma variável está na lista de protegidas
func isProtectedVariable(key string) bool {
	for _, protected := range protectedEnvVars {
		if strings.EqualFold(protected, key) {
			return true
		}
	}
	return false
}

// isAppendOnlyVariable verifica se uma variável permite apenas append
func isAppendOnlyVariable(key string) bool {
	for _, appendOnly := range appendOnlyEnvVars {
		if strings.EqualFold(appendOnly, key) {
			return true
		}
	}
	return false
}

// getPathSeparator retorna o separador correto baseado no sistema operacional
func getPathSeparator(osType string) string {
	if osType == "windows" {
		return ";"
	}
	return ":"
}

// setEnvironmentVariable define uma variável de ambiente
func setEnvironmentVariable(cmd *cobra.Command, osType, key, value string, shouldAppend, force bool) error {
	verbosePrint(cmd, "Definindo variável: %s = %s", key, value)

	// Verificar se é uma variável que permite apenas append
	if isAppendOnlyVariable(key) {
		if !shouldAppend {
			return fmt.Errorf("erro: variável '%s' é protegida e só pode ser modificada usando --append\n\n"+
				"Esta variável é crítica e não pode ser substituída completamente.\n"+
				"Use 'bast env --set --key %s --value VALOR --append' para adicionar ao valor existente.", key, key)
		}

		// Validar separador no valor fornecido
		separator := getPathSeparator(osType)
		if strings.Contains(value, separator) {
			return fmt.Errorf("erro: o valor não deve conter o separador '%s'\n\n"+
				"Para variáveis como %s, forneça apenas um caminho por vez.\n"+
				"O separador será adicionado automaticamente.", separator, key)
		}

		verbosePrint(cmd, "Variável %s permite apenas append, validando separador", key)
	}

	// Verificar se é uma variável protegida (mas não append-only)
	if isProtectedVariable(key) && !isAppendOnlyVariable(key) {
		if !force {
			return fmt.Errorf("erro: variável '%s' é protegida e não pode ser modificada\n\n"+
				"Esta variável é crítica para o funcionamento do sistema.\n"+
				"Se você realmente deseja modificá-la, use --force (não recomendado)", key)
		}

		// Aviso para variáveis protegidas mesmo com --force
		fmt.Printf("AVISO: Você está prestes a modificar a variável protegida '%s'\n", key)
		fmt.Println("   Esta variável é crítica para o funcionamento do sistema.")
		fmt.Println("   Modificá-la incorretamente pode causar problemas!")
		fmt.Print("\n   Tem certeza que deseja continuar? Esta ação pode ser perigosa.\n\n")
	}

	// Verificar se a variável já existe
	existingValue := os.Getenv(key)
	if existingValue != "" && !force && !shouldAppend {
		return fmt.Errorf("erro: variável '%s' já existe com valor '%s'\n\nUse --force para substituir "+
			"ou --append para adicionar ao valor existente", key, existingValue)
	}

	// Preparar valor final
	finalValue := value
	if shouldAppend && existingValue != "" {
		// Adicionar ao valor existente
		if osType == "windows" {
			finalValue = existingValue + ";" + value
		} else {
			finalValue = existingValue + ":" + value
		}
		verbosePrint(cmd, "Adicionando ao valor existente: %s", finalValue)
	}

	// Definir variável baseado no sistema operacional
	switch osType {
	case "windows":
		return setWindowsEnvironmentVariable(cmd, key, finalValue)
	case "linux", "darwin":
		return setUnixEnvironmentVariable(cmd, key, finalValue)
	default:
		return fmt.Errorf("erro: sistema operacional '%s' não suportado", osType)
	}
}

// setWindowsEnvironmentVariable define variável de ambiente no Windows
func setWindowsEnvironmentVariable(cmd *cobra.Command, key, value string) error {
	verbosePrint(cmd, "Usando PowerShell para definir variável no Windows")

	// Escapar aspas no valor para PowerShell
	escapedValue := strings.ReplaceAll(value, "'", "''")
	formattedCommand := fmt.Sprintf(
		"[System.Environment]::SetEnvironmentVariable('%s', '%s', 'User')",
		key, escapedValue,
	)

	psCmd := exec.Command("powershell", "-Command", formattedCommand)
	psCmd.Stdout = os.Stdout
	psCmd.Stderr = os.Stderr

	if err := psCmd.Run(); err != nil {
		verbosePrint(cmd, "Erro ao executar PowerShell: %v", err)
		return fmt.Errorf("erro ao definir variável de ambiente: %w\n\nCertifique-se de que o PowerShell está disponível", err)
	}

	fmt.Printf("Variável '%s' definida com sucesso!\n", key)
	fmt.Println("\nIMPORTANTE: É necessário reiniciar o terminal para que a variável seja reconhecida.")
	fmt.Println("   A variável foi definida no escopo do usuário, mas só estará disponível")
	fmt.Println("   em novas sessões do terminal.")
	verbosePrint(cmd, "Variável definida com sucesso via PowerShell")

	return nil
}

// setUnixEnvironmentVariable define variável de ambiente no Linux/macOS
func setUnixEnvironmentVariable(cmd *cobra.Command, key, value string) error {
	verbosePrint(cmd, "Definindo variável no sistema Unix")

	// Obter shell do usuário
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	// Determinar arquivo de configuração baseado no shell
	var configFile string
	if strings.Contains(shell, "zsh") {
		configFile = "~/.zshrc"
	} else {
		configFile = "~/.bashrc"
	}

	fmt.Printf("Para definir variáveis de ambiente permanentemente no %s, adicione ao seu arquivo de configuração:\n", runtime.GOOS)
	fmt.Printf("   export %s=\"%s\"\n", key, value)
	fmt.Printf("\nOu execute manualmente:\n")
	fmt.Printf("   echo 'export %s=\"%s\"' >> %s\n", key, value, configFile)
	fmt.Println("\nPara aplicar imediatamente:")
	fmt.Printf("   export %s=\"%s\"\n", key, value)
	fmt.Println("\nOu recarregue o arquivo de configuração:")
	if strings.Contains(shell, "zsh") {
		fmt.Println("   source ~/.zshrc")
	} else {
		fmt.Println("   source ~/.bashrc")
	}

	verbosePrint(cmd, "Instruções fornecidas para definir variável no Unix")

	// Tentar definir para a sessão atual usando os.Setenv
	if err := os.Setenv(key, value); err != nil {
		verbosePrint(cmd, "Aviso: não foi possível definir para a sessão atual: %v", err)
	} else {
		fmt.Printf("\nVariável '%s' definida para a sessão atual\n", key)
	}

	fmt.Println("\nIMPORTANTE: Para tornar a variável permanente, adicione ao arquivo de configuração")
	fmt.Println("   e reinicie o terminal ou recarregue o arquivo com 'source'.")
	fmt.Println("   A variável definida acima só está disponível nesta sessão atual.")

	return nil
}

// deleteEnvironmentVariable deleta uma variável de ambiente do usuário
func deleteEnvironmentVariable(cmd *cobra.Command, osType, key string) error {
	verbosePrint(cmd, "Verificando variável para deleção: %s", key)

	// Verificar se é uma variável protegida
	if isProtectedVariable(key) {
		return fmt.Errorf("erro: variável '%s' é protegida e não pode ser deletada\n\n"+
			"Esta variável é crítica para o funcionamento do sistema.\n"+
			"Deletá-la pode causar problemas graves no sistema.\n"+
			"Esta operação não é permitida por segurança.", key)
	}

	// Verificar se a variável existe (apenas para informar ao usuário)
	existingValue := os.Getenv(key)
	if existingValue == "" {
		verbosePrint(cmd, "Variável não encontrada na sessão atual, mas pode existir no escopo do usuário")
		fmt.Printf("Variável '%s' não encontrada na sessão atual.\n", key)
		fmt.Println("   Tentando deletar do escopo do usuário...")
	} else {
		verbosePrint(cmd, "Variável encontrada na sessão atual: %s = %s", key, existingValue)
		fmt.Printf("Variável encontrada: %s = %s\n", key, existingValue)
	}

	// Deletar variável baseado no sistema operacional (sempre no escopo do usuário)
	switch osType {
	case "windows":
		return deleteWindowsEnvironmentVariable(cmd, key)
	case "linux", "darwin":
		return deleteUnixEnvironmentVariable(cmd, key)
	default:
		return fmt.Errorf("erro: sistema operacional '%s' não suportado", osType)
	}
}

// deleteWindowsEnvironmentVariable deleta variável de ambiente do usuário no Windows
func deleteWindowsEnvironmentVariable(cmd *cobra.Command, key string) error {
	verbosePrint(cmd, "Usando PowerShell para deletar variável do usuário no Windows")

	// PowerShell: SetEnvironmentVariable com $null deleta a variável
	formattedCommand := fmt.Sprintf(
		"[System.Environment]::SetEnvironmentVariable('%s', $null, 'User')",
		key,
	)

	psCmd := exec.Command("powershell", "-Command", formattedCommand)
	psCmd.Stdout = os.Stdout
	psCmd.Stderr = os.Stderr

	if err := psCmd.Run(); err != nil {
		verbosePrint(cmd, "Erro ao executar PowerShell: %v", err)
		return fmt.Errorf("erro ao deletar variável de ambiente: %w\n\nCertifique-se de que o PowerShell está disponível", err)
	}

	fmt.Printf("Variável '%s' deletada com sucesso!\n", key)
	fmt.Println("\nIMPORTANTE: É necessário reiniciar o terminal para que a mudança seja reconhecida.")
	fmt.Println("   A variável foi deletada do escopo do usuário, mas ainda pode estar")
	fmt.Println("   disponível na sessão atual do terminal até que seja reiniciado.")
	verbosePrint(cmd, "Variável deletada com sucesso via PowerShell")

	return nil
}

// deleteUnixEnvironmentVariable deleta variável de ambiente no Linux/macOS
func deleteUnixEnvironmentVariable(cmd *cobra.Command, key string) error {
	verbosePrint(cmd, "Fornecendo instruções para deletar variável no sistema Unix")

	// Obter shell do usuário
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	// Determinar arquivo de configuração baseado no shell
	var configFile string
	if strings.Contains(shell, "zsh") {
		configFile = "~/.zshrc"
	} else {
		configFile = "~/.bashrc"
	}

	fmt.Printf("Para deletar variáveis de ambiente permanentemente no %s:\n", runtime.GOOS)
	fmt.Println("\n1. Remova a linha do arquivo de configuração:")
	fmt.Printf("   %s\n", configFile)
	fmt.Printf("   Procure por: export %s=...\n", key)
	fmt.Println("\n2. Ou use sed para remover automaticamente:")
	if runtime.GOOS == "darwin" {
		// macOS usa BSD sed: backup vazio obrigatório com -i
		fmt.Printf("   sed -i '' '/^export %s=/d' %s\n", key, configFile)
	} else {
		fmt.Printf("   sed -i '/^export %s=/d' %s\n", key, configFile)
	}
	fmt.Println("\n3. Para remover da sessão atual:")
	fmt.Printf("   unset %s\n", key)
	fmt.Println("\n4. Recarregue o arquivo de configuração:")
	if strings.Contains(shell, "zsh") {
		fmt.Println("   source ~/.zshrc")
	} else {
		fmt.Println("   source ~/.bashrc")
	}

	verbosePrint(cmd, "Instruções fornecidas para deletar variável no Unix")

	// Tentar remover da sessão atual usando os.Unsetenv (Go 1.19+)
	// Para versões anteriores, apenas avisar
	if err := os.Unsetenv(key); err != nil {
		verbosePrint(cmd, "Aviso: não foi possível remover da sessão atual: %v", err)
		fmt.Printf("\nNão foi possível remover '%s' da sessão atual.\n", key)
		fmt.Println("   Execute manualmente: unset " + key)
	} else {
		fmt.Printf("\nVariável '%s' removida da sessão atual\n", key)
		fmt.Println("   Para remover permanentemente, siga as instruções acima.")
	}

	return nil
}
