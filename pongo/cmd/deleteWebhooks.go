/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// deleteWebhooksCmd represents the deleteWebhooks command
var deleteWebhooksCmd = &cobra.Command{
	Use:     "delete-webhooks <leaderboard-name>",
	Aliases: []string{"dw", "del", "delete"},
	Short:   "Deletes all webhooks registered to the given leaderboard.",
	Long: `Deletes all registered webhooks from the given leaderboard.
	This action cannot be undone. You will be prompted an 'Are you sure?' before completion.`,
	Example:               "\tpongo delete-webhooks leaderboard\n\tpongo dw leaderboard",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		leaderboard := args[0]

		fmt.Printf("\n> Are you sure you want to delete all webhook-urls from '%s'? (y/n)\t", leaderboard)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			sendCommand("delete-webhooks", leaderboard)
		} else {
			fmt.Println("Delete operation cancelled.")
		}

	},
}

func init() {
	rootCmd.AddCommand(deleteWebhooksCmd)

	deleteWebhooksCmd.Flags().BoolP("help", "h", false, "Deletes all webhooks from specified leaderboard. Cannot be undone.")
}
