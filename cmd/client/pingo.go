/*
Copyright © 2025 6ixfigs
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var pingo = &cobra.Command{
	Use:          "pingo",
	Short:        "CLI for interacting with the Pongo server",
	SilenceUsage: true,
}

var version = &cobra.Command{
	Use:   "version",
	Short: "Print Pingo version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pingo v1.0.0")
	},
}

var leaderboard = &cobra.Command{
	Use:     "leaderboard {create,get}",
	Short:   "Create or retrieve leaderboards",
	Long:    "The leaderboard command allows you to create and retrieve leaderboards. Leaderboards are used to track player rankings and match results in a structured and competitive format.",
	Aliases: []string{"l"},
}

var leaderboardCreate = &cobra.Command{
	Use:                   "create <name>",
	Short:                 "Create a new leaderboard",
	Long:                  "Creates a new leaderboard with the specified name.",
	Aliases:               []string{"c"},
	Example:               "pingo leaderboard create OnlyRealGs",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/leaderboards"
		formData := map[string]string{"name": args[0]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

var leaderboardGet = &cobra.Command{
	Use:                   "get <name>",
	Short:                 "Retrieve a leaderboard",
	Long:                  "Retrieves player rankings on the specified leaderboard.",
	Aliases:               []string{"g"},
	Example:               "pingo leaderboard get OnlyRealGs",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

var player = &cobra.Command{
	Use:     "player {create,stats}",
	Short:   "Create a player or retrieve stats",
	Long:    "The player command enables you to create new players or retrieve their statistics. Players are the participants in your ping-pong matches, and their stats are tracked within leaderboards.",
	Aliases: []string{"p"},
}

var playerCreate = &cobra.Command{
	Use:                   "create <leaderboard> <username>",
	Short:                 "Creates a new player",
	Long:                  "Creates a new player with the specified username on the specified leaderboard.",
	Aliases:               []string{"c"},
	Example:               "pingo player create 2pac",
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/players", args[0])
		formData := map[string]string{"username": args[1]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

var playerStats = &cobra.Command{
	Use:                   "stats <leaderboard> <player>",
	Short:                 "Retrieve player stats",
	Long:                  "Retrieves detailed statistics for a specific player from the specified leaderboard.",
	Aliases:               []string{"s"},
	Example:               "pingo player get 2pac",
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/players/%s", args[0], args[1])
		return sendCommand(path, nil, http.MethodGet)
	},
}

var webhooks = &cobra.Command{
	Use:     "webhooks {register,list,delete}",
	Short:   "Manage webhooks",
	Long:    "The webhooks command allows you to manage webhooks for a leaderboard. Webhooks can be used to receive updates or notifications about match results and leaderboard changes.",
	Aliases: []string{"w"},
}

var webhooksRegister = &cobra.Command{
	Use:                   "register <leaderboard> <url>",
	Short:                 "Register a webhook",
	Long:                  "Registers a new webhook for a specified leaderboard. Use this command to set up a URL that will receive notifications about match results and other leaderboard events.",
	Aliases:               []string{"r"},
	Example:               "pingo webhooks register OnlyRealGs https://onlyrealgs.com/incoming",
	Args:                  cobra.ExactArgs(2),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/webhooks", args[0])
		formData := map[string]string{"url": args[1]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

var webhooksList = &cobra.Command{
	Use:                   "list <leaderboard>",
	Short:                 "List all registered webhooks",
	Long:                  "Lists all registered webhooks for a specified leaderboard. Use this command to view the URLs that are currently receiving updates from the leaderboard.",
	Aliases:               []string{"l"},
	Example:               "pingo webhooks list OnlyRealGs",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/webhooks", args[0])
		return sendCommand(path, nil, http.MethodGet)
	},
}

var webhooksDelete = &cobra.Command{
	Use:                   "delete <leaderboard>",
	Short:                 "Delete all webhooks",
	Long:                  "Deletes all registered webhooks for a specified leaderboard. Use this command to remove webhooks and stop receiving notifications for a leaderboard.",
	Aliases:               []string{"d"},
	Example:               "pingo webhooks delete OnlyRealGs",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		leaderboard := args[0]

		fmt.Printf("\n> Are you sure you want to delete all webhooks from '%s'? (y/n)\t", leaderboard)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
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

var record = &cobra.Command{
	Use:                   "record <leaderboard> <player1> <player2> <score>",
	Short:                 "Record a match between two players",
	Long:                  "Records the outcome of a match between two players in a specified leaderboard. Use this command to log match results, update player rankings, and maintain an accurate recordof played matches.",
	Aliases:               []string{"r"},
	Example:               "pingo record OnlyRealGs eazy-e 2pac 2-1",
	Args:                  cobra.ExactArgs(4),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fmt.Sprintf("/leaderboards/%s/matches", args[0])
		formData := map[string]string{"player1": args[1], "player2": args[2], "score": args[3]}
		return sendCommand(path, formData, http.MethodPost)
	},
}

func init() {
	pingo.CompletionOptions.DisableDefaultCmd = true

	pingo.AddCommand(version)

	leaderboard.AddCommand(leaderboardCreate)
	leaderboard.AddCommand(leaderboardGet)
	pingo.AddCommand(leaderboard)

	player.AddCommand(playerCreate)
	player.AddCommand(playerStats)
	pingo.AddCommand(player)

	webhooks.AddCommand(webhooksRegister)
	webhooks.AddCommand(webhooksList)
	webhooks.AddCommand(webhooksDelete)
	pingo.AddCommand(webhooks)

	pingo.AddCommand(record)
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

	text, err := io.ReadAll(resp.Body)
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
	err := pingo.Execute()
	if err != nil {
		os.Exit(1)
	}
}
