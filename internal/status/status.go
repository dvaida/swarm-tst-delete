package status

import (
	"fmt"
	"io"
)

// Run executes the status command for the given paths
func Run(paths []string, w io.Writer) error {
	fmt.Fprintln(w, "Status command not yet implemented")
	return nil
}
