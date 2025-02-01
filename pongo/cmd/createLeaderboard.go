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
	Use:     "create-leaderboard",
	Aliases: []string{"cl"},
	Short:   "Creates a new leaderboard group for players to join.",
	Long: `Makes a new group for players to join into. Every leaderboard has it's own
ranking. Players are allowed to join multiple leaderboards.
Example usage:
	pongo create-leaderboard pongers

Short form:
	pongo cl pongers`,
	Example: `pongo create-leaderboard MyLeaderboard
	pongo cl MyLeaderboard`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
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
