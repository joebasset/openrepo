package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:                "openrepo",
		Short:              "Scaffold fullstack monorepos",
		SilenceUsage:       true,
		SilenceErrors:      true,
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootCmd.AddCommand(newCreateCmd())

	return rootCmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
