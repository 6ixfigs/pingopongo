/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// registerWebhookCmd represents the registerWebhook command
var registerWebhookCmd = &cobra.Command{
	Use:                   "register-webhook <leaderboard-name> <webhook-url>",
	Short:                 "Adds a webhook-url to the specified leaderboard.",
	Aliases:               []string{"rw", "reg"},
	Example:               "\tpongo register-webhook MyLeaderboard https://slack-webhook-url\n\tpongo rw MyLeaderboard https://discord-webhook-url",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("register-webhook", strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(registerWebhookCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// registerWebhookCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// registerWebhookCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
