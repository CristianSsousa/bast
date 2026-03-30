package install

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/sirupsen/logrus"
)

// InstallGit instala o Git no sistema quando ausente; se já instalado, apenas informa.
func InstallGit(log logrus.FieldLogger) error {
	log.Debug("Iniciando processo de instalação do Git...")
	fmt.Println("Verificando se o Git já está instalado...")

	if isGitInstalled() {
		log.Debug("Git encontrado no sistema.")
		fmt.Println("Git já está instalado!")
		version, err := getGitVersion()
		if version != "" {
			fmt.Printf("  Versão: %s\n", version)
		}
		if err != nil {
			log.Debugf("Erro ao obter versão: %v", err)
		}
		return nil
	}

	fmt.Println("Git não encontrado. Iniciando instalação...")
	fmt.Printf("Sistema operacional detectado: %s\n", runtime.GOOS)
	log.Debugf("Arquitetura: %s", runtime.GOARCH)

	instCmd, installMethod, err := gitInstallCommand(log)
	if err != nil {
		return err
	}

	fmt.Printf("Método de instalação: %s\n", installMethod)
	log.Debugf("Comando completo: %s", instCmd.String())
	fmt.Println("Executando comando de instalação...")
	fmt.Println("Nota: Você pode precisar inserir sua senha de administrador.")

	if installMethod == "apt-get" && runtime.GOOS == "linux" {
		fmt.Println("Atualizando lista de pacotes...")
		updateCmd := exec.Command("sudo", "apt-get", "update")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		log.Debug("Executando: sudo apt-get update")
		if err := updateCmd.Run(); err != nil {
			fmt.Printf("Aviso: falha ao atualizar lista de pacotes: %v\n", err)
			log.Debugf("Erro detalhado: %v", err)
			fmt.Println("Continuando com a instalação...")
		} else {
			log.Debug("Lista de pacotes atualizada com sucesso.")
		}
		instCmd = exec.Command("sudo", "apt-get", "install", "-y", "git")
		log.Debug("Comando de instalação: sudo apt-get install -y git")
	}

	instCmd.Stdout = os.Stdout
	instCmd.Stderr = os.Stderr
	instCmd.Stdin = os.Stdin

	log.Debug("Iniciando execução do comando de instalação...")
	if err := instCmd.Run(); err != nil {
		log.Debugf("Erro durante execução: %v", err)
		return fmt.Errorf("executar instalação: %w", err)
	}

	fmt.Println("\nInstalação concluída!")
	fmt.Println("Verificando instalação...")
	log.Debug("Verificando se Git está acessível no PATH...")

	if isGitInstalled() {
		version, err := getGitVersion()
		if version != "" {
			fmt.Printf("Git instalado com sucesso! Versão: %s\n", version)
		} else {
			fmt.Println("Git instalado com sucesso!")
		}
		if err != nil {
			log.Debugf("Aviso ao obter versão: %v", err)
		}
		log.Debug("Instalação verificada e funcionando corretamente.")
	} else {
		fmt.Println("Git pode ter sido instalado, mas não foi encontrado no PATH.")
		fmt.Println("   Tente fechar e reabrir o terminal.")
		log.Debugf("PATH atual: %s", os.Getenv("PATH"))
	}
	return nil
}

func gitInstallCommand(log logrus.FieldLogger) (*exec.Cmd, string, error) {
	switch runtime.GOOS {
	case "windows":
		return windowsGitInstall(log)
	case "linux":
		return linuxGitInstall(log)
	case "darwin":
		return darwinGitInstall(log)
	default:
		return nil, "", fmt.Errorf("sistema operacional '%s' não suportado para instalação automática", runtime.GOOS)
	}
}

func windowsGitInstall(log logrus.FieldLogger) (*exec.Cmd, string, error) {
	if isCommandAvailable("winget") {
		log.Debug("Usando winget como gerenciador de pacotes.")
		return exec.Command("winget", "install", "--id", "Git.Git", "-e", "--source", "winget"), "winget", nil
	}
	if isCommandAvailable("choco") {
		log.Debug("Usando Chocolatey como gerenciador de pacotes.")
		return exec.Command("choco", "install", "git", "-y"), "chocolatey", nil
	}
	log.Debug("Nenhum gerenciador de pacotes encontrado (winget ou chocolatey).")
	return nil, "", fmt.Errorf("não foi possível determinar o método de instalação para Windows")
}

func linuxGitInstall(log logrus.FieldLogger) (*exec.Cmd, string, error) {
	if isCommandAvailable("apt-get") {
		log.Debug("Usando apt-get como gerenciador de pacotes.")
		return exec.Command("sudo", "apt-get", "install", "-y", "git"), "apt-get", nil
	}
	if isCommandAvailable("yum") {
		log.Debug("Usando yum como gerenciador de pacotes.")
		return exec.Command("sudo", "yum", "install", "-y", "git"), "yum", nil
	}
	if isCommandAvailable("dnf") {
		log.Debug("Usando dnf como gerenciador de pacotes.")
		return exec.Command("sudo", "dnf", "install", "-y", "git"), "dnf", nil
	}
	if isCommandAvailable("pacman") {
		log.Debug("Usando pacman como gerenciador de pacotes.")
		return exec.Command("sudo", "pacman", "-S", "--noconfirm", "git"), "pacman", nil
	}
	if isCommandAvailable("zypper") {
		log.Debug("Usando zypper como gerenciador de pacotes.")
		return exec.Command("sudo", "zypper", "install", "-y", "git"), "zypper", nil
	}
	log.Debug("Nenhum gerenciador de pacotes Linux encontrado.")
	return nil, "", fmt.Errorf("nenhum gerenciador de pacotes Linux encontrado")
}

func darwinGitInstall(log logrus.FieldLogger) (*exec.Cmd, string, error) {
	if isCommandAvailable("brew") {
		log.Debug("Usando Homebrew como gerenciador de pacotes.")
		return exec.Command("brew", "install", "git"), "homebrew", nil
	}
	log.Debug("Homebrew não encontrado. Instale em: https://brew.sh")
	return nil, "", fmt.Errorf("homebrew não encontrado")
}

func isGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

func getGitVersion() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	version := string(output)
	if version != "" && version[len(version)-1] == '\n' {
		version = version[:len(version)-1]
	}
	return version, nil
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
