/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// registerWebhookCmd represents the registerWebhook command
var registerWebhookCmd = &cobra.Command{
	Use:                   "register-webhook <leaderboard-name> <webhook-url>",
	Short:                 "Adds a webhook-url to the specified leaderboard.",
	Aliases:               []string{"rw", "reg"},
	Example:               "  pongo register-webhook MyLeaderboard https://slack-webhook-url\n  pongo rw MyLeaderboard https://discord-webhook-url",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/webhooks", args[0])
		formData := map[string]string{"url": args[1]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

func init() {
	rootCmd.AddCommand(registerWebhookCmd)
	registerWebhookCmd.Flags().BoolP("help", "h", false, "Register a new webhook-url for the specified leaderboard.")
}
