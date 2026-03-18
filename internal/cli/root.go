package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "openrepo",
		Short:         "Scaffold opinionated monorepos with Cobra and Huh",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newCreateCmd())

	return rootCmd
}

func Execute() error {
	return NewRootCmd().Execute()
}
