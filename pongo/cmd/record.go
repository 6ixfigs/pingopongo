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
	Long:                  `Stores match data. The score should be in the format 'player1_sets_won-player2_sets_won'`,
	Example:               "  pongo record CroPongClub zoran-milanovic dragan-primorac 21-0\n  pongo r pongers marcel vux 1-1",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		sendCommand("record", strings.Join(args, " "))
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)

	recordCmd.Flags().BoolP("help", "h", false, "Record a match between two players. Run 'pongo example' to see detailed instructions.")
}
