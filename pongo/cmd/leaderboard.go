/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// leaderboardCmd represents the leaderboard command
var leaderboardCmd = &cobra.Command{
	Use:                   "leaderboard <leaderboard-name>",
	Aliases:               []string{"l"},
	Short:                 "Displays the ranking inside the leaderboard.",
	Long:                  `Shows the top 15 players and their scores in the given leaderboard.`,
	Example:               "pongo leaderboard pongers",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("leaderboard", strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(leaderboardCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// leaderboardCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// leaderboardCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
