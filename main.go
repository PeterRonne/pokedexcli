package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type clicommand struct {
	name        string
	description string
	callback    func(config *Config) error
}

// type Config struct {
// 	NextUrl     string `json:"next_url"`
// 	PreviousUrl string
// }

type locationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Count         int            `json:"count"`
	Next          string         `json:"next"`
	Previous      *string        `json:"previous"`
	LocationAreas []locationArea `json:"results"`
}

func commandExit(config *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	for _, command := range cliCommands {
		text := fmt.Sprintf("%s: %s", command.name, command.description)
		fmt.Println(text)
	}
	return nil
}

func commandMap(config *Config) error {
	var url string
	if config.Next == "" {
		url = "https://pokeapi.co/api/v2/location-area"
	} else {
		url = config.Next
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return err
	}

	for _, location := range config.LocationAreas {
		fmt.Println(location.Name)
	}
	return nil
}

func commandMapb(config *Config) error {
	if config.Previous == nil {
		println("you're on the first page")
		return nil
	}

	res, err := http.Get(*config.Previous)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return err
	}

	for _, location := range config.LocationAreas {
		fmt.Println(location.Name)
	}

	return nil
}

func cleanInput(text string) []string {
	if text == "" {
		return []string{""}
	}
	words := strings.Split(strings.TrimSpace(strings.ToLower(text)), " ")
	return words
}

var cliCommands map[string]clicommand

func main() {
	cliCommands = map[string]clicommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 locations if not one the first page",
			callback:    commandMapb,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	config := Config{}

	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			cleanedInput := cleanInput(scanner.Text())
			command, ok := cliCommands[cleanedInput[0]]
			if !ok {
				fmt.Println("Unknown command")
				continue
			}
			if err := command.callback(&config); err != nil {
				fmt.Println(err)
			}
		}
	}
}
