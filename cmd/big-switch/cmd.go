package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "big-switch",
		Short: "Big-switch production deployer",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Flags().Lookup("debug").Changed {
				log.SetLevel(log.DebugLevel)
			}
		},
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newEncryptCmd())
	rootCmd.AddCommand(newDecryptCmd())
	rootCmd.PersistentFlags().Bool("debug", false, "Turn on debug logging.")

	return rootCmd
}

var (
	buildTime    = "unknown"
	buildVersion = "dev"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s (built: %s)\n", buildVersion, buildTime)
		},
	}
}

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts the deployer server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}
}

func newEncryptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "encrypt <filename>",
		Short: "Encrypt a configuration file for later use",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := encryptConfig(args[0])
			if err != nil {
				log.Error(err)
			}
		},
	}
}

func newDecryptCmd() *cobra.Command {
	decryptPassphrase := ""
	cmd := cobra.Command{
		Use:   "decrypt <filename>",
		Short: "Decrypt a previously encrypted configuration file. The passphrase flag must be set.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if decryptPassphrase == "" {
				log.Fatal("The passphrase cannot be empty.")
			}
			plain, err := deryptConfig(args[0], decryptPassphrase)
			if err != nil {
				log.Error(err)
			}
			fmt.Println(string(plain))
		},
	}

	cmd.Flags().StringVarP(&decryptPassphrase, "passphrase", "p", "", "Use the given passphrase to decrypt the file.")

	return &cmd
}
