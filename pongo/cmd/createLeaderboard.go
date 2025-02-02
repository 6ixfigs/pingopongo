/*
Copyright © 2025 6ixfigs
*/
package cmd

import (
	"net/http"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/leaderboards"
		formData := map[string]string{"name": args[0]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

func init() {
	rootCmd.AddCommand(createLeaderboardCmd)

	createLeaderboardCmd.Flags().BoolP("help", "h", false, "Creates a new leaderboard and registers a webhook-url.")
}
