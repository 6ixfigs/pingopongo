/*
Copyright © 2025 6ixfigs
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pongo",
	Short: "Client application for sending request to a ping-pong match tracking server.",
	Long: `Pongo is a command-line tool which helps users send data to ping-pong match tracking servers.
The server parses commands and forwards them to webhooks specified in the command-line.
For example:

	pongo create-leaderboard <leaderboard-name>
	pongo register-webhook <leaderboard-name> <another-url> (optional)
	pongo create-player <leaderboard-name> <username>
	pongo record <leaderboard-name> <player1> <player2> <score>`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("help", "h", false, "This is a client app for recording ping pong matches.")
}
