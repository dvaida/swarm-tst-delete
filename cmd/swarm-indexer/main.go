package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "swarm-indexer",
		Short: "Index text files for AI context retrieval",
		Long:  "A CLI tool that indexes text files from registered paths into Typesense for AI context retrieval (RAG), using semantic chunking and Gemini embeddings for hybrid search.",
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
		Long:  "Index text files from the specified path into Typesense.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search indexed files",
		Long:  "Search the indexed files using the specified query.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show indexer status",
		Long:  "Show the current status of the swarm-indexer.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
