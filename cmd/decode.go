package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gliwka/SAMLurai/internal/output"
	"github.com/gliwka/SAMLurai/internal/saml"
	"github.com/spf13/cobra"
)

var (
	decodeFile    string
	decodeDeflate bool
)

var decodeCmd = &cobra.Command{
	Use:   "decode [base64-encoded-saml]",
	Short: "Decode a base64-encoded SAML assertion",
	Long: `Decode a base64-encoded SAML assertion or response.

The input can be provided as:
  - A command line argument
  - From a file using the -f flag
  - From stdin (pipe)

For SAML requests using HTTP-Redirect binding, use the --deflate flag
to decompress the deflated content.

Examples:
  # Decode from argument
  samlurai decode "PHNhbWxwOlJlc3BvbnNl..."

  # Decode from file
  samlurai decode -f response.txt

  # Decode from stdin
  echo "PHNhbWxwOlJlc3BvbnNl..." | samlurai decode

  # Decode with deflate decompression
  samlurai decode --deflate -f request.txt`,
	RunE: runDecode,
}

func init() {
	rootCmd.AddCommand(decodeCmd)

	decodeCmd.Flags().StringVarP(&decodeFile, "file", "f", "", "Read base64-encoded SAML from file")
	decodeCmd.Flags().BoolVar(&decodeDeflate, "deflate", false, "Apply deflate decompression (for HTTP-Redirect binding)")
}

func runDecode(cmd *cobra.Command, args []string) error {
	input, err := getDecodeInput(cmd, args)
	if err != nil {
		return err
	}

	decoder := saml.NewDecoder()
	var decoded []byte

	if decodeDeflate {
		decoded, err = decoder.DecodeDeflate(input)
	} else {
		decoded, err = decoder.Decode(input)
	}

	if err != nil {
		return fmt.Errorf("failed to decode SAML: %w", err)
	}

	formatter := output.NewFormatter(outputFormat)
	formatted, err := formatter.FormatXML(decoded)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Fprint(cmd.OutOrStdout(), formatted)
	return nil
}

func getDecodeInput(cmd *cobra.Command, args []string) (string, error) {
	// Priority: file flag > argument > stdin
	if decodeFile != "" {
		data, err := os.ReadFile(decodeFile)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	if len(args) > 0 {
		return strings.TrimSpace(args[0]), nil
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

	return "", fmt.Errorf("no input provided. Use -f flag, provide an argument, or pipe data to stdin")
}
