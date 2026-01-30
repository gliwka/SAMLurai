package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gliwka/SAMLurai/internal/output"
	"github.com/gliwka/SAMLurai/internal/saml"
	"github.com/spf13/cobra"
)

var (
	inspectFile   string
	inspectKey    string
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect and display SAML assertion details",
	Long: `Parse and display SAML assertion details in a human-readable format.

The input can be:
  - A SAML XML file
  - A base64-encoded SAML file
  - A HAR (HTTP Archive) file - displays all SAML assertions in order
  - Data from stdin (pipe)

This command automatically:
  - Detects HAR files and extracts all SAML assertions
  - Decodes base64-encoded input (with optional deflate)
  - Decrypts encrypted assertions (if -k flag is provided)

This command displays:
  - Issuer information
  - Subject (NameID)
  - Conditions (validity period)
  - Attributes
  - Authentication statements

Examples:
  # Inspect from XML file
  samlurai inspect -f assertion.xml

  # Inspect all SAML from a HAR file (in order)
  samlurai inspect -f session.har

  # Inspect HAR file with decryption key
  samlurai inspect -f session.har -k private.pem

  # Inspect base64-encoded SAML (auto-decoded)
  echo "PHNhbWw+Li4uPC9zYW1sPg==" | samlurai inspect

  # Inspect encrypted assertion (auto-decrypted)
  samlurai inspect -f encrypted.xml -k private.pem

  # Output as JSON
  samlurai inspect -f assertion.xml -o json`,
	RunE: runInspect,
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().StringVarP(&inspectFile, "file", "f", "", "Read SAML from file (supports XML, base64, or HAR files)")
	inspectCmd.Flags().StringVarP(&inspectKey, "key", "k", "", "Path to private key for decryption (PEM format)")
}

func runInspect(cmd *cobra.Command, args []string) error {
	input, err := getInspectInput(cmd)
	if err != nil {
		return err
	}

	// Check if input is a HAR file
	if isHARFile(inspectFile, input) {
		return runInspectHAR(cmd, []byte(input))
	}

	// Regular SAML inspection
	return runInspectSAML(cmd, input)
}

// isHARFile checks if the input is likely a HAR file
func isHARFile(filename, content string) bool {
	// Check file extension
	if filename != "" {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext == ".har" {
			return true
		}
	}
	
	// Check content for HAR JSON structure
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "{") && strings.Contains(trimmed, `"log"`) && strings.Contains(trimmed, `"entries"`) {
		return true
	}
	
	return false
}

// runInspectHAR handles inspection of HAR files
func runInspectHAR(cmd *cobra.Command, data []byte) error {
	extractor := saml.NewHARExtractor()
	results, err := extractor.ExtractFromHAR(data)
	if err != nil {
		return fmt.Errorf("failed to parse HAR file: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No SAML assertions found in the HAR file.")
		return nil
	}

	formatter := output.NewFormatter(outputFormat)
	
	// Print header for HAR inspection
	fmt.Fprintf(cmd.OutOrStdout(), "Found %d SAML message(s) in HAR file:\n\n", len(results))

	for i, extracted := range results {
		// Print separator and context for each SAML message
		if i > 0 {
			fmt.Fprintln(cmd.OutOrStdout())
		}
		
		fmt.Fprintf(cmd.OutOrStdout(), "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(cmd.OutOrStdout(), " [%d/%d] %s from %s\n", i+1, len(results), extracted.Type, extracted.Source)
		if extracted.ParameterName != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "       Parameter: %s\n", extracted.ParameterName)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "       URL: %s\n", truncateURL(extracted.URL, 70))
		fmt.Fprintf(cmd.OutOrStdout(), "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		// Process the SAML data
		xmlData := extracted.DecodedXML

		// Auto-decrypt if encrypted and key is provided
		if saml.IsEncrypted(xmlData) {
			if inspectKey == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Encrypted assertion detected - provide -k flag to decrypt\n\n")
				// Still try to show what we can from the response wrapper
				parser := saml.NewParser()
				info, err := parser.ParsePartial(xmlData)
				if err == nil && info != nil {
					formatted, _ := formatter.FormatSAMLInfo(info)
					fmt.Fprint(cmd.OutOrStdout(), formatted)
				}
				continue
			}

			decryptor, err := saml.NewDecryptor(inspectKey)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Failed to load private key: %v\n\n", err)
				continue
			}

			xmlData, err = decryptor.Decrypt(xmlData)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Failed to decrypt: %v\n\n", err)
				continue
			}
		}

		// Parse and display
		parser := saml.NewParser()
		info, err := parser.Parse(xmlData)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Failed to parse: %v\n\n", err)
			continue
		}

		formatted, err := formatter.FormatSAMLInfo(info)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Failed to format: %v\n\n", err)
			continue
		}

		fmt.Fprint(cmd.OutOrStdout(), formatted)
	}

	return nil
}

// runInspectSAML handles inspection of regular SAML files
func runInspectSAML(cmd *cobra.Command, input string) error {
	// Step 1: Auto-decode if input is base64-encoded
	decoder := saml.NewDecoder()
	xmlData, err := decoder.SmartDecode(input)
	if err != nil {
		return fmt.Errorf("failed to decode input: %w", err)
	}

	// Step 2: Auto-decrypt if encrypted and key is provided
	if saml.IsEncrypted(xmlData) {
		if inspectKey == "" {
			return fmt.Errorf("encrypted SAML detected but no private key provided. Use -k flag to specify a key")
		}

		decryptor, err := saml.NewDecryptor(inspectKey)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}

		xmlData, err = decryptor.Decrypt(xmlData)
		if err != nil {
			return fmt.Errorf("failed to decrypt SAML: %w", err)
		}
	}

	// Step 3: Parse and display
	parser := saml.NewParser()
	info, err := parser.Parse(xmlData)
	if err != nil {
		return fmt.Errorf("failed to parse SAML: %w", err)
	}

	formatter := output.NewFormatter(outputFormat)
	formatted, err := formatter.FormatSAMLInfo(info)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Fprint(cmd.OutOrStdout(), formatted)
	return nil
}

func getInspectInput(cmd *cobra.Command) (string, error) {
	if inspectFile != "" {
		data, err := os.ReadFile(inspectFile)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// Check if stdin has data
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("no input provided. Use -f flag or pipe data to stdin")
}
