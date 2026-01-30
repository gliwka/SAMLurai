package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information
	version = "dev"

	// Global flags
	outputFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "samlurai",
	Short: "A CLI tool for debugging SAML assertions",
	Long: `SAMLurai is a command-line tool for decoding, decrypting, and inspecting 
SAML assertions. It helps developers and security professionals debug 
SAML-based authentication flows.

Examples:
  # Decode a base64-encoded SAML response
  echo "PHNhbWw..." | samlurai decode

  # Decode with deflate decompression (HTTP-Redirect binding)
  samlurai decode --deflate -f request.txt

  # Decrypt an encrypted SAML assertion
  samlurai decrypt -k private.pem -f encrypted.xml

  # Inspect SAML assertion details
  samlurai inspect -f assertion.xml`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "pretty", "Output format: pretty, json, xml")
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

// SetVersion sets the version for the root command (used in tests)
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

// GetRootCmd returns the root command (used in tests)
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// OutputWriter returns the writer for command output
func OutputWriter() *os.File {
	return os.Stdout
}

// ErrorWriter returns the writer for error output
func ErrorWriter() *os.File {
	return os.Stderr
}

// Errorf prints an error message to stderr
func Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
