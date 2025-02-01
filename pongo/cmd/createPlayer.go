/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// createPlayerCmd represents the createPlayer command
var createPlayerCmd = &cobra.Command{
	Use:                   "create-player <leaderboard-name> <username>",
	Aliases:               []string{"cp", "new-player"},
	Short:                 "Creates a new player inside a leaderboard.",
	Long:                  `Creates a player with <username> and registers it inside the <leaderboard-name> leaderboard.`,
	Example:               "  pongo create-player MyLeaderboard zoran-milanovic\n  pongo cp MyLeaderboard zoran-milanovic",
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("create-player", strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(createPlayerCmd)

	createPlayerCmd.Flags().BoolP("help", "h", false, "Creates a player inside the specified leaderboard.")
}
