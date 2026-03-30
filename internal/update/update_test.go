package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"2.0.0", "2.0.0", 0},
	}
	for _, tt := range tests {
		if got := CompareVersions(tt.v1, tt.v2); got != tt.want {
			t.Errorf("CompareVersions(%q,%q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestTrimVersionPrefix(t *testing.T) {
	t.Parallel()
	if got := TrimVersionPrefix("v1.2.3"); got != "1.2.3" {
		t.Fatalf("got %q", got)
	}
}

func TestValidTag(t *testing.T) {
	t.Parallel()
	if ValidTag("") || ValidTag("bad tag") {
		t.Fatal("esperado inválido")
	}
	if !ValidTag("v1.0.0") {
		t.Fatal("esperado válido")
	}
}

func TestFetchLatest(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0","name":"Release"}`))
	}))
	defer ts.Close()

	client := ts.Client()
	rel, err := FetchLatest(client, ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if rel.TagName != "v1.0.0" || rel.Name != "Release" {
		t.Fatalf("release: %+v", rel)
	}
}

func TestIsUpToDate(t *testing.T) {
	t.Parallel()
	if !IsUpToDate("2.0.0", "1.0.0") {
		t.Fatal("2.0.0 deve estar atualizado vs 1.0.0")
	}
	if IsUpToDate("1.0.0", "2.0.0") {
		t.Fatal("1.0.0 não deve estar atualizado vs 2.0.0")
	}
}
