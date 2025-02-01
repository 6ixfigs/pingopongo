/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// exampleCmd represents the example command
var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Example usage of the pongo CLI app.",
	Long: `This is how you would use pongo to record a match between two new players.

	pongo create-leaderboard pongers
	pongo register-webhook https://slack.com/pingypongy_webhook
	pongo create-player zoran-primorac
	pongo create-player mile-kitic
	pongo register pongers zoran-primorac mile-kitic 7-0

If you had previously created a leaderboard and registered a webhook to it, you may skip the first 2 commands.
Also, creating players is unnecessary if they had previously been created.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exampleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exampleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
