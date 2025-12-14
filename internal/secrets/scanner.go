package secrets

// Finding represents a detected secret
type Finding struct {
	Line    int
	Column  int
	Match   string
	Type    string
}

// ScanResult contains the results of a secret scan
type ScanResult struct {
	Findings   []Finding
	ShouldSkip bool // Type A: entire file should be skipped
}

// Scanner scans content for secrets
type Scanner struct{}

// New creates a new Scanner
func New() *Scanner {
	return &Scanner{}
}

// ScanFile checks if a file should be skipped entirely (Type A secrets)
func (s *Scanner) ScanFile(path string) (*ScanResult, error) {
	return &ScanResult{
		Findings:   nil,
		ShouldSkip: false,
	}, nil
}

// ScanContent scans content for inline secrets (Type B)
func (s *Scanner) ScanContent(content string) (*ScanResult, error) {
	return &ScanResult{
		Findings:   nil,
		ShouldSkip: false,
	}, nil
}

// Redact removes secrets from content
func (s *Scanner) Redact(content string, findings []Finding) string {
	// For now, return content unchanged
	return content
}
