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
	decryptFile    string
	decryptKeyFile string
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt an encrypted SAML assertion",
	Long: `Decrypt an encrypted SAML assertion using a private key.

The encrypted assertion can be provided from:
  - A file using the -f flag
  - From stdin (pipe)

The private key must be in PEM format.

If the input is base64-encoded (with optional deflate compression),
it will be automatically decoded before decryption.

Examples:
  # Decrypt from file
  samlurai decrypt -k private.pem -f encrypted_assertion.xml

  # Decrypt from stdin
  cat encrypted.xml | samlurai decrypt -k private.pem

  # Decrypt base64-encoded input (auto-detected)
  echo "PHNhbWw6RW5jcnlwdGVkQXNzZXJ0aW9uPi4uLg==" | samlurai decrypt -k private.pem

  # Output as JSON
  samlurai decrypt -k private.pem -f encrypted.xml -o json`,
	RunE: runDecrypt,
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	decryptCmd.Flags().StringVarP(&decryptFile, "file", "f", "", "Read encrypted SAML from file")
	decryptCmd.Flags().StringVarP(&decryptKeyFile, "key", "k", "", "Path to private key (PEM format)")
	_ = decryptCmd.MarkFlagRequired("key")
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	input, err := getDecryptInput(cmd)
	if err != nil {
		return err
	}

	// Auto-decode if input is base64-encoded
	decoder := saml.NewDecoder()
	xmlData, err := decoder.SmartDecode(input)
	if err != nil {
		return fmt.Errorf("failed to decode input: %w", err)
	}

	decryptor, err := saml.NewDecryptor(decryptKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	decrypted, err := decryptor.Decrypt(xmlData)
	if err != nil {
		return fmt.Errorf("failed to decrypt SAML assertion: %w", err)
	}

	formatter := output.NewFormatter(outputFormat)
	formatted, err := formatter.FormatXML(decrypted)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	fmt.Fprint(cmd.OutOrStdout(), formatted)
	return nil
}

func getDecryptInput(cmd *cobra.Command) (string, error) {
	if decryptFile != "" {
		data, err := os.ReadFile(decryptFile)
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
