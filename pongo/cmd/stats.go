/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:                   "stats <leaderboard-name> <username>",
	Aliases:               []string{"s"},
	Short:                 "Displays user's stats inside the given leaderboard.",
	Long:                  `Dispays user's games won, win percentage, winning streak and more.`,
	Example:               "  pongo stats MyLeaderboard marcel-muslija\n  pongo s MyLeaderboard luka-bikota",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := fmt.Sprintf("/leaderboards/%s/players/%s", args[0], args[1])
		sendCommand(path, nil, http.MethodGet)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
