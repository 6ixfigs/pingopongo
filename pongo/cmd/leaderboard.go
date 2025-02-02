/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// leaderboardCmd represents the leaderboard command
var leaderboardCmd = &cobra.Command{
	Use:                   "leaderboard <leaderboard-name>",
	Aliases:               []string{"l"},
	Short:                 "Displays the ranking inside the leaderboard.",
	Long:                  `Shows the top 15 players and their scores in the specified leaderboard.`,
	Example:               "  pongo leaderboard pongers\n  pongo l pongers",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

func init() {
	rootCmd.AddCommand(leaderboardCmd)

	leaderboardCmd.Flags().BoolP("help", "h", false, "Display the ranking from the specified leaderboard.")
}
