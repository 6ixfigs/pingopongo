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
	Use:                   "record <leaderboard-name> <player1> <player2> <score>",
	Aliases:               []string{"r", "rec"},
	Short:                 "Records a match between two players.",
	Long:                  `Sends the command containing the match recording to server.`,
	Example:               "\tpongo record CroPongClub zoran-milanovic dragan-primorac 21-0\t pongo r pongers marcel vux 1-1",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("record", strings.Join(args, " "))
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
