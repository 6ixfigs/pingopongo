/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// createLeaderboardCmd represents the createLeaderboard command
var createLeaderboardCmd = &cobra.Command{
	Use:     "create-leaderboard <leaderboard-name>",
	Aliases: []string{"cl"},
	Short:   "Creates a new leaderboard group for players to join.",
	Long: `Makes a new group for players to join into. Every leaderboard has it's own
ranking. Players can be created inside multiple leaderboards with the same name.`,

	Example: `pongo create-leaderboard MyLeaderboard
	pongo cl MyLeaderboard`,

	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,

	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("create-leaderboard", strings.Join(args, ""))
	},
}

func init() {
	rootCmd.AddCommand(createLeaderboardCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createLeaderboardCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createLeaderboardCmd.Flags().BoolP("help", "h", false, "Creates a new leaderboard.")
}
