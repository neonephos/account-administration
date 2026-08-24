// Command enterprise-sync reconciles a GitHub enterprise's owners and members
// from a YAML file, in the style of Prow's peribolos (which is org-scoped only).
//
// It manages enterprise OWNERS and MEMBERS. Organizations are out of scope.
//
// Auth: GITHUB_TOKEN must be an enterprise-installed GitHub App installation
// token (or a classic PAT with admin:enterprise) able to call the enterprise
// GraphQL APIs.
//
// Usage:
//
//	enterprise-sync --config enterprise/neonephos.yaml            # dry-run
//	enterprise-sync --config enterprise/neonephos.yaml --confirm  # apply
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		configPath      = flag.String("config", "", "path to the enterprise config YAML (required)")
		confirm         = flag.Bool("confirm", false, "apply changes; without this flag it is a dry-run")
		minOwners       = flag.Int("min-owners", 2, "refuse configs declaring fewer than this many owners (lockout guard)")
		maxRemovalDelta = flag.Float64("max-removal-delta", 0.25, "refuse plans removing more than this fraction of enterprise people (typo guard)")
	)
	flag.Parse()

	if *configPath == "" {
		return fmt.Errorf("--config is required")
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is not set")
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}

	g := guards{minOwners: *minOwners, maxRemovalDelta: *maxRemovalDelta}
	// Config-only guard first, before any network call.
	if err := checkMinOwners(cfg, g); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	c := newClient(token)

	entID, err := c.enterpriseID(ctx, cfg.Enterprise)
	if err != nil {
		return err
	}

	live, err := c.fetchState(ctx, cfg.Enterprise)
	if err != nil {
		return err
	}

	plan := computePlan(cfg, live)
	if err := plan.check(cfg, g); err != nil {
		return err
	}

	mode := "DRY-RUN"
	if *confirm {
		mode = "APPLY"
	}
	fmt.Printf(">>> enterprise %q: %d action(s) [%s]\n", cfg.Enterprise, len(plan.actions), mode)

	if err := plan.apply(ctx, c, entID, *confirm); err != nil {
		return err
	}
	fmt.Println("Finished syncing enterprise configuration.")
	return nil
}
