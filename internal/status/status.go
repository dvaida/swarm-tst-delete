package status

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dvaida/swarm-indexer/internal/indexer"
	"github.com/dvaida/swarm-indexer/internal/metadata"
)

// Run executes the status command for the given paths
func Run(paths []string, client indexer.TypesenseClient, w io.Writer) error {
	fmt.Fprintln(w, "Indexed Paths:")
	fmt.Fprintln(w, "─────────────────────────────────────────────────────────────")
	fmt.Fprintln(w)

	for _, path := range paths {
		if err := showPathStatus(path, w); err != nil {
			return err
		}
	}

	fmt.Fprintln(w, "─────────────────────────────────────────────────────────────")

	// Show Typesense stats
	showTypesenseStats(client, w)

	return nil
}

func showPathStatus(path string, w io.Writer) error {
	fmt.Fprintf(w, "📁 %s\n", path)

	meta, err := metadata.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(w, "   Status: Not indexed (no metadata found)")
			fmt.Fprintln(w)
			return nil
		}
		return fmt.Errorf("failed to load metadata for %s: %w", path, err)
	}

	// Display metadata info
	languages := strings.Join(meta.Languages, ", ")
	if languages == "" {
		languages = "none"
	}

	fmt.Fprintf(w, "   Type: %s | Files: %s | Languages: %s\n",
		meta.ProjectType,
		formatNumber(meta.FileCount),
		languages)

	lastIndexed := time.Unix(meta.LastIndexed, 0).Format("2006-01-02 15:04:05")
	fmt.Fprintf(w, "   Last indexed: %s\n", lastIndexed)

	// Check if re-index is needed
	currentHash, err := metadata.ComputeContentHash(path)
	if err != nil {
		fmt.Fprintf(w, "   Status: ⚠ Could not compute hash: %v\n", err)
	} else if currentHash != meta.ContentHash {
		fmt.Fprintln(w, "   Status: ⚠ Changes detected (re-index needed)")
	} else {
		fmt.Fprintln(w, "   Status: ✓ Up to date")
	}

	fmt.Fprintln(w)
	return nil
}

func showTypesenseStats(client indexer.TypesenseClient, w io.Writer) {
	stats, err := client.GetCollectionStats()
	if err != nil {
		fmt.Fprintf(w, "Typesense: Connection failed (%v)\n", err)
		return
	}

	fmt.Fprintf(w, "Typesense Collection: %s\n", stats.CollectionName)
	fmt.Fprintf(w, "   Documents: %s\n", formatNumber(int(stats.DocumentCount)))
	fmt.Fprintf(w, "   URL: %s\n", client.GetURL())
}

// formatNumber formats an integer with thousand separators
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// Format with commas
	str := fmt.Sprintf("%d", n)
	result := make([]byte, 0, len(str)+len(str)/3)

	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}

	return string(result)
}
