// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/APS-Conecta/Modding/internal/catalog"
	"github.com/APS-Conecta/Modding/internal/state"
	"github.com/APS-Conecta/Modding/internal/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Modding:", err)
		os.Exit(1)
	}
}

func run() error {
	listen := flag.String("listen", "127.0.0.1:8787", "IP y puerto de escucha; use TLS fuera de loopback")
	directory := flag.String("state-dir", ".modding", "Directorio privado para la configuración")
	cert := flag.String("tls-cert", "", "Certificado TLS para acceso por red")
	key := flag.String("tls-key", "", "Clave privada TLS")
	catalogPath := flag.String("catalog", "", "CSV DEIS alternativo con el esquema del catálogo de gestión")
	catalogVersion := flag.String("catalog-version", "", "Versión del CSV alternativo")
	flag.Parse()
	host, _, err := net.SplitHostPort(*listen)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.IsUnspecified() {
		return fmt.Errorf("indique una IP concreta de escucha, no un nombre ni una dirección comodín")
	}
	secure := *cert != "" && *key != ""
	if (*cert == "") != (*key == "") {
		return fmt.Errorf("indique tanto tls-cert como tls-key")
	}
	if !ip.IsLoopback() && !secure {
		return fmt.Errorf("el acceso por red requiere tls-cert y tls-key")
	}
	var c *catalog.Catalog
	if *catalogPath != "" {
		if *catalogVersion == "" {
			return fmt.Errorf("catalog-version es obligatorio con catalog")
		}
		data, err := os.ReadFile(*catalogPath)
		if err != nil {
			return err
		}
		c, err = catalog.Load(data, *catalogVersion)
		if err != nil {
			return err
		}
	} else {
		if *catalogVersion != "" {
			return fmt.Errorf("catalog-version requiere catalog")
		}
		c, err = catalog.Default()
		if err != nil {
			return err
		}
	}
	store, err := state.Open(*directory)
	if err != nil {
		return err
	}
	defer store.Close()
	app, token, err := web.New(c, store, web.Options{Host: *listen, Secure: secure})
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		return err
	}
	defer listener.Close()
	scheme := "http"
	if secure {
		scheme = "https"
	}
	fmt.Fprintf(os.Stderr, "Modding · configuración y vista previa\nAbra %s://%s\nCódigo de acceso de un solo uso: %s\nEste asistente guarda un plan; todavía no instala ni modifica Nextcloud.\n", scheme, *listen, token)
	server := &http.Server{Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	done := make(chan error, 1)
	go func() {
		if secure {
			done <- server.ServeTLS(listener, *cert, *key)
		} else {
			done <- server.Serve(listener)
		}
	}()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
