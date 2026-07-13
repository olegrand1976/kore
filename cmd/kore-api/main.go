package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kore/kore/internal/app"
	"github.com/kore/kore/internal/platform/config"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrate(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		if len(os.Args) > 2 && os.Args[2] == "reset" {
			if err := runSeedReset(); err != nil {
				log.Fatal(err)
			}
			return
		}
		if err := runSeed(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := runServer(); err != nil {
		log.Fatal(err)
	}
}

func runMigrate() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer application.Close()
	if err := application.Migrate(ctx); err != nil {
		return err
	}
	log.Println("migrations applied successfully")
	return nil
}

func runSeed() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer application.Close()
	if err := application.Migrate(ctx); err != nil {
		return err
	}
	return application.Seed(ctx)
}

func runSeedReset() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer application.Close()
	if err := application.Migrate(ctx); err != nil {
		return err
	}
	return application.ResetSeed(ctx)
}

func runServer() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}

	if cfg.MigrateOnBoot {
		if err := application.Migrate(ctx); err != nil {
			application.Close()
			return fmt.Errorf("migrate on boot: %w", err)
		}
	}
	if cfg.DevSeedEnabled {
		if err := application.Seed(ctx); err != nil {
			application.Close()
			return fmt.Errorf("seed: %w", err)
		}
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           application.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("kore-api listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		application.Close()
		return err
	}
	application.Close()
	return nil
}
