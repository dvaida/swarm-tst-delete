package main

import (
	"fmt"
	"os"

	"github.com/dvaida/swarm-indexer/internal/config"
	"github.com/dvaida/swarm-indexer/internal/indexer"
	"github.com/dvaida/swarm-indexer/internal/status"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "swarm-indexer",
		Short: "Index text files into Typesense for AI context retrieval",
	}

	statusCmd := &cobra.Command{
		Use:   "status [paths...]",
		Short: "Show indexing status for paths",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()

			client, err := indexer.NewClient(cfg.TypesenseURL, cfg.TypesenseAPIKey, cfg.TypesenseCollection)
			if err != nil {
				return fmt.Errorf("failed to create Typesense client: %w", err)
			}

			return status.Run(args, client, os.Stdout)
		},
	}

	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
