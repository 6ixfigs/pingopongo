/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// createLeaderboardCmd represents the createLeaderboard command
var createLeaderboardCmd = &cobra.Command{
	Use:                   "create-leaderboard <leaderboard-name>",
	Aliases:               []string{"cl"},
	Short:                 "Creates a new leaderboard group for players to join.",
	Long:                  "Makes a new group for players to join into. Every leaderboard has it's own ranking.\nPlayers can be created inside multiple leaderboards with the same name.",
	Example:               "  pongo create-leaderboard MyLeaderboard\n  pongo cl MyLeaderboard",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("create-leaderboard", args[0])
	},
}

func init() {
	rootCmd.AddCommand(createLeaderboardCmd)

	createLeaderboardCmd.Flags().BoolP("help", "h", false, "Creates a new leaderboard and registers a webhook-url.")
}
