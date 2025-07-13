package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type clicommand struct {
	name        string
	description string
	callback    func() error
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	for _, command := range cliCommands {
		text := fmt.Sprintf("%s: %s", command.name, command.description)
		fmt.Println(text)
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
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			cleanedInput := cleanInput(scanner.Text())
			command, ok := cliCommands[cleanedInput[0]]
			if !ok {
				fmt.Println("Unknown command")
				continue
			}
			if err := command.callback(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
