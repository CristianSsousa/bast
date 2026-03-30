package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// GitHubAPIURL é o endpoint da API para a última release do repositório.
	GitHubAPIURL = "https://api.github.com/repos/CristianSsousa/go-bast-cli/releases/latest"
	// DefaultHTTPTimeout timeout padrão para requisições à API.
	DefaultHTTPTimeout = 10 * time.Second
)

// ModulePath é o caminho do módulo Go usado em go install.
const ModulePath = "github.com/CristianSsousa/go-bast-cli"

// Release representa uma release retornada pela API do GitHub.
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

// FetchLatest obtém a release mais recente do GitHub usando o client informado.
func FetchLatest(client *http.Client, apiURL string) (*Release, error) {
	if client == nil {
		client = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	if apiURL == "" {
		apiURL = GitHubAPIURL
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("criar requisição: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API GitHub: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ler resposta: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("decodificar JSON: %w", err)
	}

	return &release, nil
}

// TrimVersionPrefix remove o prefixo "v" de tags tipo v1.2.3.
func TrimVersionPrefix(tag string) string {
	return strings.TrimPrefix(tag, "v")
}

// IsUpToDate retorna true se currentVersion >= latestVersion (comparação semver simples).
func IsUpToDate(currentVersion, latestVersion string) bool {
	return CompareVersions(currentVersion, latestVersion) >= 0
}

// CompareVersions compara duas versões no formato semver (X.Y.Z).
// Retorna: -1 se v1 < v2, 0 se v1 == v2, 1 se v1 > v2.
func CompareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int

		if i < len(parts1) {
			if _, err := fmt.Sscanf(parts1[i], "%d", &num1); err != nil {
				num1 = 0
			}
		}
		if i < len(parts2) {
			if _, err := fmt.Sscanf(parts2[i], "%d", &num2); err != nil {
				num2 = 0
			}
		}

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	return 0
}

// ValidTag retorna true se a tag pode ser usada com go install @tag.
func ValidTag(tag string) bool {
	if tag == "" {
		return false
	}
	return !strings.Contains(tag, " ")
}
