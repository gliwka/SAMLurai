package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gliwka/SAMLurai/internal/output"
	"github.com/gliwka/SAMLurai/internal/saml"
	"github.com/spf13/cobra"
)

var (
	extractFile      string
	extractOutputDir string
	extractList      bool
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract SAML assertions from HAR/XHR files",
	Long: `Extract all SAML assertions found in HAR (HTTP Archive) files.

HAR files are commonly exported from browser developer tools and contain 
all HTTP requests and responses from a browsing session. This command 
scans the HAR file for SAML assertions in:
  - POST request bodies (SAMLResponse, SAMLRequest parameters)
  - URL query parameters (HTTP-Redirect binding)
  - HTML responses containing hidden form fields

Each extracted SAML assertion is saved to a separate file with a 
descriptive name indicating its type and source.

Examples:
  # Extract all SAML assertions from a HAR file
  samlurai extract -f session.har

  # Extract to a specific directory
  samlurai extract -f session.har -d ./extracted

  # List SAML assertions without extracting
  samlurai extract -f session.har --list

  # Extract from Chrome DevTools HAR export
  samlurai extract -f chrome_network.har -d ./saml_assertions`,
	RunE: runExtract,
}

func init() {
	rootCmd.AddCommand(extractCmd)

	extractCmd.Flags().StringVarP(&extractFile, "file", "f", "", "HAR file to extract SAML from (required)")
	extractCmd.Flags().StringVarP(&extractOutputDir, "dir", "d", ".", "Output directory for extracted files")
	extractCmd.Flags().BoolVar(&extractList, "list", false, "List found SAML assertions without extracting")
	_ = extractCmd.MarkFlagRequired("file")
}

func runExtract(cmd *cobra.Command, args []string) error {
	// Read HAR file
	data, err := os.ReadFile(extractFile)
	if err != nil {
		return fmt.Errorf("failed to read HAR file: %w", err)
	}

	// Extract SAML assertions
	extractor := saml.NewHARExtractor()
	results, err := extractor.ExtractFromHAR(data)
	if err != nil {
		return fmt.Errorf("failed to extract SAML: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No SAML assertions found in the HAR file.")
		return nil
	}

	// List mode - just show what was found
	if extractList {
		return listExtractedSAML(cmd, results)
	}

	// Extract mode - save to files
	return saveExtractedSAML(cmd, extractor, results)
}

func listExtractedSAML(cmd *cobra.Command, results []saml.ExtractedSAML) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Found %d SAML assertion(s):\n\n", len(results))

	for _, r := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "  [%d] %s\n", r.Index, r.Type)
		fmt.Fprintf(cmd.OutOrStdout(), "      Source: %s\n", r.Source)
		if r.ParameterName != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "      Parameter: %s\n", r.ParameterName)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "      URL: %s\n", truncateURL(r.URL, 60))
		if r.WasDeflated {
			fmt.Fprintf(cmd.OutOrStdout(), "      Encoding: base64 + deflate\n")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "      Encoding: base64\n")
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	return nil
}

func saveExtractedSAML(cmd *cobra.Command, extractor *saml.HARExtractor, results []saml.ExtractedSAML) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(extractOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	formatter := output.NewFormatter("pretty")
	savedFiles := []string{}

	for _, r := range results {
		filename := extractor.GenerateFilename(r)
		filepath := filepath.Join(extractOutputDir, filename)

		// Format the XML nicely
		formatted, err := formatter.FormatXML(r.DecodedXML)
		if err != nil {
			// If formatting fails, use raw XML
			formatted = string(r.DecodedXML)
		}

		// Write to file
		if err := os.WriteFile(filepath, []byte(formatted), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		savedFiles = append(savedFiles, filename)
	}

	// Print summary
	fmt.Fprintf(cmd.OutOrStdout(), "Extracted %d SAML assertion(s) to %s:\n\n", len(results), extractOutputDir)

	for i, r := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "  [%d] %s → %s\n", r.Index, r.Type, savedFiles[i])
		fmt.Fprintf(cmd.OutOrStdout(), "      Source: %s", r.Source)
		if r.ParameterName != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " (%s)", r.ParameterName)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	return nil
}

// truncateURL truncates a URL for display
func truncateURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}
