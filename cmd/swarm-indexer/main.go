package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for swarm-indexer
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "swarm-indexer",
		Short: "Index text files for AI context retrieval",
		Long:  "A CLI tool that indexes text files from registered paths into Typesense for AI context retrieval (RAG).",
	}

	rootCmd.AddCommand(newIndexCmd())
	rootCmd.AddCommand(newSearchCmd())
	rootCmd.AddCommand(newStatusCmd())

	return rootCmd
}

func newIndexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "index [path]",
		Short: "Index files from a path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("index command not yet implemented")
			return nil
		},
	}
}

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search indexed files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("search command not yet implemented")
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show indexer status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("status command not yet implemented")
			return nil
		},
	}
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
