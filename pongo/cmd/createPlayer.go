/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// createPlayerCmd represents the createPlayer command
var createPlayerCmd = &cobra.Command{
	Use:                   "create-player <leaderboard-name> <username>",
	Aliases:               []string{"cp", "new-player"},
	Short:                 "Creates a new player inside the specified leaderboard.",
	Long:                  `Creates a player with <username> and registers it inside the <leaderboard-name> leaderboard.`,
	Example:               "  pongo create-player MyLeaderboard zoran-milanovic\n  pongo cp MyLeaderboard zoran-milanovic",
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/players", args[0])
		formData := map[string]string{"username": args[1]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

func init() {
	rootCmd.AddCommand(createPlayerCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	createPlayerCmd.Flags().BoolP("help", "h", false, "Creates a player inside the specified leaderboard.")
}
