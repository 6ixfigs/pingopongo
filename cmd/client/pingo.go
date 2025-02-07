/*
Copyright © 2025 6ixfigs
*/
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pingo",
	Short: "Client application for sending request to a ping-pong match tracking server.",
	Long: `pingo is a command-line tool which helps users send data to ping-pong match tracking servers.
The server parses commands and forwards them to webhooks specified in the command-line.
For example:

	pingo create-leaderboard <leaderboard-name>
	pingo register-webhook <leaderboard-name> <another-url> (optional)
	pingo create-player <leaderboard-name> <username>
	pingo record <leaderboard-name> <player1> <player2> <score>`,
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
	Example:               "  pingo create-leaderboard MyLeaderboard\n  pingo cl MyLeaderboard",
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
	Example:               "  pingo create-player MyLeaderboard zoran-milanovic\n  pingo cp MyLeaderboard zoran-milanovic",
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
	Example:               "\tpingo delete-webhooks leaderboard\n\tpingo dw leaderboard",
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
	Example:               "  pingo register-webhook MyLeaderboard https://slack-webhook-url\n  pingo rw MyLeaderboard https://discord-webhook-url",
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
	Short:   "Example usage of the pingo CLI app.",
	Example: `This is how you would use pingo to record a match between two new players.

	pingo create-leaderboard pongers
	pingo register-webhook https://another_webhook (optional)
	pingo create-player zoran-primorac
	pingo create-player mile-kitic
	pingo register pongers zoran-primorac mile-kitic 7-0

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
	Example:               "  pingo leaderboard pongers\n  pingo l pongers",
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
	Example:               "pingo list-webhooks MyLeaderboard",
	Short:                 "List all webhooks registered to the specified leaderboard.",
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/webhooks", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

var recordCmd = &cobra.Command{
	Use:                   "record <leaderboard-name> <player1> <player2> <score>",
	Aliases:               []string{"r", "rec"},
	Short:                 "Records a match between two players.",
	Long:                  `Stores match data. The score should be in the format 'player1_sets_won-player2_sets_won'`,
	Example:               "  pingo record CroPongClub zoran-milanovic dragan-primorac 21-0\n  pingo r pongers marcel vux 1-1",
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
	Example:               "  pingo stats MyLeaderboard marcel-muslija\n  pingo s MyLeaderboard luka-bikota",
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
	recordCmd.Flags().BoolP("help", "h", false, "Record a match between two players. Run 'pingo example' to see detailed instructions.")

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
	exampleCmd.Flags().BoolP("help", "h", false, "Displays detailed instructions for the pingo CLI app.")

	rootCmd.AddCommand(leaderboardCmd)
	leaderboardCmd.Flags().BoolP("help", "h", false, "Display the ranking from the specified leaderboard.")

	rootCmd.AddCommand(listWebhooksCmd)
	listWebhooksCmd.Flags().BoolP("help", "h", false, "List all webhooks from the specified leaderboard.")

	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolP("help", "h", false, "Shows stats for the specified player.")
}

func sendCommand(path string, formData map[string]string, method string) error {

	form := url.Values{}
	for key, value := range formData {
		form.Set(key, value)
	}

	serverURL, err := getServerURL()
	if err != nil {
		return err
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

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n%s", resp.Status, string(text))
	return nil
}

func getServerURL() (string, error) {

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configFile := filepath.Join(configDir, "pingo.cfg")
	file, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	serverURL := scanner.Text()

	if serverURL == "" {
		fmt.Print("> Please enter base URL of pongo host: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		serverURL = scanner.Text()

		file.Write([]byte(serverURL))

		fmt.Println("Base URL saved at ", configFile)
	}

	return serverURL, nil
}

func main() {
	Execute()
}
