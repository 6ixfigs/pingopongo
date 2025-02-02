/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// deleteWebhooksCmd represents the deleteWebhooks command
var deleteWebhooksCmd = &cobra.Command{
	Use:     "delete-webhooks <leaderboard-name>",
	Aliases: []string{"dw", "del", "delete"},
	Short:   "Deletes all webhooks registered to the specified leaderboard.",
	Long: `Deletes all registered webhooks from the specified leaderboard.
	This action cannot be undone. You will be prompted an 'Are you sure?' before completion.`,
	Example:               "\tpongo delete-webhooks leaderboard\n\tpongo dw leaderboard",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		leaderboard := args[0]

		fmt.Printf("\n> Are you sure you want to delete all webhook-urls from '%s'? (y/n)\t", leaderboard)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			path := fmt.Sprintf("/leaderboards/%s/webhooks", args[0])
			return sendCommand(path, nil, http.MethodDelete)
		} else {
			fmt.Println("Delete operation cancelled.")
			return nil
		}

	},
}

func init() {
	rootCmd.AddCommand(deleteWebhooksCmd)

	deleteWebhooksCmd.Flags().BoolP("help", "h", false, "Deletes all webhooks from specified leaderboard. Cannot be undone.")
}
