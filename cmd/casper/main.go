package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Port                    uint16
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	GracefulShutdownTimeout time.Duration
	EnableTLS               bool
}

func CastIntToUint16(n int) (uint16, error) {
	if n < 0 {
		return 0, fmt.Errorf("%d is negative", n)
	}
	if n > math.MaxUint16 {
		return 0, fmt.Errorf("%d is bigger than the max uint16 %d", n, math.MaxUint16)
	}
	return uint16(n), nil
}

type Application struct {
	Config Config
	Store  *Store
}

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lshortfile)

	cfg := Config{}

	var port int
	flag.IntVar(&port, "port", 3030, "port of the server")
	flag.DurationVar(&cfg.ReadTimeout, "read-timeout", time.Second, "request read timeout")
	flag.DurationVar(&cfg.WriteTimeout, "write-timeout", time.Second, "request write timeout")
	flag.DurationVar(&cfg.GracefulShutdownTimeout, "graceful-shutdown-timeout", 10*time.Second, "graceful shutdown timeout")
	flag.BoolVar(&cfg.EnableTLS, "enable-tls", true, "enable TLS (transport layer security)")
	flag.Parse()

	var err error
	cfg.Port, err = CastIntToUint16(port)
	if err != nil {
		log.Fatal(err)
	}

	app := Application{
		Config: cfg,
		Store:  NewStore(),
	}

	mux := &http.ServeMux{}
	mux.HandleFunc("PUT /v1/key/{key}", app.putHandler)
	mux.HandleFunc("GET /v1/key/{key}", app.getHandler)
	mux.HandleFunc("DELETE /v1/key/{key}", app.deleteHandler)

	srv := http.Server{
		Addr:         fmt.Sprintf("localhost:%d", port),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      mux,
	}

	if cfg.EnableTLS {
		srv.TLSConfig = &tls.Config{
			MinVersion:       tls.VersionTLS12,
			MaxVersion:       tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	shutdownSignalCh := make(chan os.Signal, 1)
	go func() {
		signal.Notify(shutdownSignalCh, syscall.SIGINT, syscall.SIGQUIT)
		<-shutdownSignalCh
		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Starting server on port %d\n", cfg.Port)

	if cfg.EnableTLS {
		err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
	log.Println("Server was shutdown gracefully")
}
