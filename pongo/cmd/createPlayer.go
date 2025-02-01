/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// createPlayerCmd represents the createPlayer command
var createPlayerCmd = &cobra.Command{
	Use:     "create-player",
	Aliases: []string{"cp"},
	Short:   "Creates a new player inside a leaderboard.",
	Long:    `Creates a player with <username> and registers it inside the <leaderboard-name> leaderboard.`,
	Example: `pongo create-player MyLeaderboard zoran-milanovic
	pongo cp MyLeaderboard zoran-milanovic
	`,
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("create-player", strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(createPlayerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createPlayerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createPlayerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
