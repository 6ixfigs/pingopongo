/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// recordCmd represents the record command
var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Records a match between two players.",
	Long: `Sends the command containing the match recording to server.
For example:

	record <leaderboard-name> <player1> <player2> <score>`,
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("record", strings.Join(args, " "))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(4)(cmd, args); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// recordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
