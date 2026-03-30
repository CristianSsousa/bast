package serve

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Options configura o servidor HTTP do comando serve.
type Options struct {
	Host     string
	Port     string
	Endpoint string // usado apenas para logs (rotas fixas / e /health)

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultTimeouts replica os valores usados anteriormente em cmd/serve.
func DefaultTimeouts() (read, write, idle time.Duration) {
	return 15 * time.Second, 15 * time.Second, 60 * time.Second
}

// Run inicia o servidor HTTP com mux local até erro fatal em ListenAndServe.
func Run(log logrus.FieldLogger, opts Options) error {
	if opts.ReadTimeout == 0 || opts.WriteTimeout == 0 || opts.IdleTimeout == 0 {
		r, w, i := DefaultTimeouts()
		if opts.ReadTimeout == 0 {
			opts.ReadTimeout = r
		}
		if opts.WriteTimeout == 0 {
			opts.WriteTimeout = w
		}
		if opts.IdleTimeout == 0 {
			opts.IdleTimeout = i
		}
	}

	addr := fmt.Sprintf("%s:%s", opts.Host, opts.Port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
	}

	log.Infof("Servidor iniciando em http://%s", addr)
	log.Infof("Endpoint principal: %s", opts.Endpoint)

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("servidor HTTP: %w", err)
	}
	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World! Este é um servidor CLI construído com Go e Cobra.")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
