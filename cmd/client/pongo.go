/*
Copyright © 2025 6ixfigs
*/
package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra-cli/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "pongo",
	Short: "Client application for sending request to a ping-pong match tracking server.",
	Long: `Pongo is a command-line tool which helps users send data to ping-pong match tracking servers.
The server parses commands and forwards them to webhooks specified in the command-line.
For example:

	pongo create-leaderboard <leaderboard-name>
	pongo register-webhook <leaderboard-name> <another-url> (optional)
	pongo create-player <leaderboard-name> <username>
	pongo record <leaderboard-name> <player1> <player2> <score>`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var createLeaderboardCmd = &cobra.Command{
	Use:                   "create-leaderboard <leaderboard-name>",
	Aliases:               []string{"cl"},
	Short:                 "Creates a new leaderboard group for players to join.",
	Long:                  "Makes a new group for players to join into. Every leaderboard has it's own ranking.\nPlayers can be created inside multiple leaderboards with the same name.",
	Example:               "  pongo create-leaderboard MyLeaderboard\n  pongo cl MyLeaderboard",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/leaderboards"
		formData := map[string]string{"name": args[0]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

var createPlayerCmd = &cobra.Command{
	Use:                   "create-player <leaderboard-name> <username>",
	Aliases:               []string{"cp", "new-player"},
	Short:                 "Creates a new player inside the specified leaderboard.",
	Long:                  `Creates a player with <username> and registers it inside the <leaderboard-name> leaderboard.`,
	Example:               "  pongo create-player MyLeaderboard zoran-milanovic\n  pongo cp MyLeaderboard zoran-milanovic",
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/players", args[0])
		formData := map[string]string{"username": args[1]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

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

var exampleCmd = &cobra.Command{
	Use:     "example",
	Aliases: []string{"e", "ex"},
	Short:   "Example usage of the pongo CLI app.",
	Example: `This is how you would use pongo to record a match between two new players.

	pongo create-leaderboard pongers
	pongo register-webhook https://another_webhook (optional)
	pongo create-player zoran-primorac
	pongo create-player mile-kitic
	pongo register pongers zoran-primorac mile-kitic 7-0

If you had previously created a leaderboard and registered a webhook to it, you may skip the first 2 commands.
Also, creating players is unnecessary if they were previously created.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Example)
	},
}

var leaderboardCmd = &cobra.Command{
	Use:                   "leaderboard <leaderboard-name>",
	Aliases:               []string{"l"},
	Short:                 "Displays the ranking inside the specified leaderboard.",
	Long:                  `Shows the top 15 players and their scores in the given leaderboard.`,
	Example:               "  pongo leaderboard pongers\n  pongo l pongers",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

var listWebhooksCmd = &cobra.Command{
	Use:                   "list-webhooks <leaderboard-name>",
	Aliases:               []string{"lw", "list", "webhooks", "hooks"},
	DisableFlagsInUseLine: true,
	Example:               "pongo list-webhooks MyLeaderboard",
	Short:                 "List all webhooks registered to the specified leaderboard.",
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

var recordCmd = &cobra.Command{
	Use:                   "record <leaderboard-name> <player1> <player2> <score>",
	Aliases:               []string{"r", "rec"},
	Short:                 "Records a match between two players.",
	Long:                  `Stores match data. The score should be in the format 'player1_sets_won-player2_sets_won'`,
	Example:               "  pongo record CroPongClub zoran-milanovic dragan-primorac 21-0\n  pongo r pongers marcel vux 1-1",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/matches", args[0])
		formData := map[string]string{"player1": args[1], "player2": args[2], "score": args[3]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

var statsCmd = &cobra.Command{
	Use:                   "stats <leaderboard-name> <username>",
	Aliases:               []string{"s"},
	Short:                 "Displays user's stats inside the given leaderboard.",
	Long:                  `Dispays user's games won, win percentage, winning streak and more.`,
	Example:               "  pongo stats MyLeaderboard marcel-muslija\n  pongo s MyLeaderboard luka-bikota",
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/players/%s", args[0], args[1])
		return sendCommand(path, nil, http.MethodGet)
	},
}

func init() {
	rootCmd.Flags().BoolP("help", "h", false, "This is a client app for recording ping pong matches.")

	rootCmd.AddCommand(recordCmd)
	recordCmd.Flags().BoolP("help", "h", false, "Record a match between two players. Run 'pongo example' to see detailed instructions.")

	rootCmd.AddCommand(createLeaderboardCmd)
	createLeaderboardCmd.Flags().BoolP("help", "h", false, "Creates a new leaderboard and registers a webhook-url.")

	rootCmd.AddCommand(createPlayerCmd)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	createPlayerCmd.Flags().BoolP("help", "h", false, "Creates a player inside the specified leaderboard.")

	rootCmd.AddCommand(deleteWebhooksCmd)
	deleteWebhooksCmd.Flags().BoolP("help", "h", false, "Deletes all webhooks from specified leaderboard. Cannot be undone.")

	rootCmd.AddCommand(registerWebhookCmd)
	registerWebhookCmd.Flags().BoolP("help", "h", false, "Register a new webhook-url for the specified leaderboard.")

	rootCmd.AddCommand(exampleCmd)
	exampleCmd.Flags().BoolP("help", "h", false, "Displays detailed instructions for the pongo CLI app.")

	rootCmd.AddCommand(leaderboardCmd)
	leaderboardCmd.Flags().BoolP("help", "h", false, "Display the ranking from the specified leaderboard.")

	rootCmd.AddCommand(listWebhooksCmd)
	listWebhooksCmd.Flags().BoolP("help", "h", false, "List all webhooks from the specified leaderboard.")

	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolP("help", "h", false, "Shows stats for the specified player.")
}

const serverURL = "http://localhost:8080"

func sendCommand(path string, formData map[string]string, method string) error {

	form := url.Values{}
	for key, value := range formData {
		form.Set(key, value)
	}

	req, err := http.NewRequest(method, serverURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Response form server: ", resp.Status)
	return nil
}

func main() {
	cmd.Execute()
}
