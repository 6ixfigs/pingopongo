/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("list-webhooks", args[0])
	},
}

func init() {
	rootCmd.AddCommand(listWebhooksCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listWebhooksCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listWebhooksCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
