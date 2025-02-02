/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// listWebhooksCmd represents the listWebhooks command
var listWebhooksCmd = &cobra.Command{
	Use:                   "list-webhooks <leaderboard-name>",
	Aliases:               []string{"lw", "list", "webhooks", "hooks"},
	DisableFlagsInUseLine: true,
	Example:               "pongo list-webhooks MyLeaderboard",
	Short:                 "List all webhooks registered to the given leaderboard.",
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

func init() {
	rootCmd.AddCommand(listWebhooksCmd)

	listWebhooksCmd.Flags().BoolP("help", "h", false, "List all webhooks from the specified leaderboard.")
}
