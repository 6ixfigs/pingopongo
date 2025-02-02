/*
Copyright © 2025 6ixfigs
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// exampleCmd represents the example command
var exampleCmd = &cobra.Command{
	Use:     "example",
	Aliases: []string{"e", "ex"},
	Short:   "Example usage of the pongo CLI app.",
	Example: `This is how you would use pongo to record a match between two new players.

	pongo create-leaderboard pongers
	pongo register-webhook https://another_webhook (optional)
	pongo create-player zoran-primorac
	pongo create-player mile-kitic
	pongo register pongers zoran-primorac mile-kitic 7-0

If you had previously created a leaderboard and registered a webhook to it, you may skip the first 2 commands.
Also, creating players is unnecessary if they were previously created.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Example)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)

	exampleCmd.Flags().BoolP("help", "h", false, "Displays detailed instructions for the pongo CLI app.")
}
