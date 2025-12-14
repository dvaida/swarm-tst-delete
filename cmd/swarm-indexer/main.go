package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dvaida/swarm-indexer/internal/search"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
	var limit int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search indexed content",
		Long:  "Search indexed content using hybrid text and vector search.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			ctx := context.Background()

			// TODO: Create real Typesense searcher when indexer is implemented
			// For now, return empty results
			searcher := &search.MockSearcher{
				Results:    []search.SearchResult{},
				EmptyIndex: true,
			}

			results, err := search.Search(ctx, searcher, query, limit)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			output := search.FormatResults(results, jsonOutput)
			fmt.Fprint(cmd.OutOrStdout(), output)

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results to return")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
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
