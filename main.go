package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PeterRonne/pokedexcli/internal/pokecache"
)

const LOCATION_AREA_URL = "https://pokeapi.co/api/v2/location-area/"
const POKEMON_URL = "https://pokeapi.co/api/v2/pokemon/"

const BASE_URL = "https://pokeapi.co/api/v2/"

type clicommand struct {
	name        string
	description string
	callback    func(config *Config, args string) error
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
	Cache         *pokecache.Cache
}

type Pokemon struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
}

type Area struct {
	PokemonEncounters []struct {
		Pokemon Pokemon `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

func commandExit(config *Config, args string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config, args string) error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	for _, command := range cliCommands {
		text := fmt.Sprintf("%s: %s", command.name, command.description)
		fmt.Println(text)
	}
	return nil
}

func commandMap(config *Config, args string) error {
	// Check cache first
	if data, found := config.Cache.Get(config.Next); found {
		// Cache hit! Use cached data
		if err := json.Unmarshal(data, config); err != nil {
			return err
		}
	} else {
		// Cache miss - make HTTP request
		res, err := http.Get(config.Next)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		// Add to cache
		config.Cache.Add(config.Next, body)

		// Unmarshal the response
		if err := json.Unmarshal(body, config); err != nil {
			return err
		}
	}

	for _, location := range config.LocationAreas {
		fmt.Println(location.Name)
	}
	return nil
}

func commandMapb(config *Config, args string) error {
	if config.Previous == nil {
		println("you're on the first page")
		return nil
	}

	if data, found := config.Cache.Get(*config.Previous); found {
		if err := json.Unmarshal(data, config); err != nil {
			return err
		}
	} else {
		res, err := http.Get(*config.Previous)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		config.Cache.Add(*config.Previous, body)

		if err := json.Unmarshal(body, config); err != nil {
			return err
		}
	}

	for _, location := range config.LocationAreas {
		fmt.Println(location.Name)
	}
	return nil
}

// func exstractNames(areas []locationArea) []string {
// 	names := make([]string, 0, len(areas))
// 	for _, area := range areas {
// 		names = append(names, area.Name)
// 	}
// 	return names
// }

func commandExplore(config *Config, args string) error {
	// if !Contains(exstractNames(config.LocationAreas), args) {
	// 	return fmt.Errorf("%s Is not availible for exploration", args)
	// }

	fmt.Printf("Exploring %s...\nFound Pokemon:\n", args)
	var area Area
	url := LOCATION_AREA_URL + args
	if data, found := config.Cache.Get(url); found {
		if err := json.Unmarshal(data, &area); err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		config.Cache.Add(url, body)

		if err := json.Unmarshal(body, &area); err != nil {
			return err
		}
	}

	for _, encounter := range area.PokemonEncounters {
		fmt.Println(" -", encounter.Pokemon.Name)
	}

	return nil
}

func tryCatchPokemon(baseExp int) bool {
	roll := rand.Intn(baseExp)
	return roll < 50
}

func commandCatch(config *Config, args string) error {
	fmt.Printf("Throwing a Pokeball at %s...\n", args)

	var pokemon Pokemon
	url := POKEMON_URL + args

	if data, found := config.Cache.Get(url); found {
		if err := json.Unmarshal(data, &pokemon); err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		config.Cache.Add(url, body)

		if err := json.Unmarshal(body, &pokemon); err != nil {
			return err
		}
	}

	if tryCatchPokemon(pokemon.BaseExperience) {
		pokedex[args] = pokemon
		println(args, "was caught!")
	} else {
		println(args, "escaped!")
	}

	return nil
}

func commandInspect(Config *Config, args string) error {
	pokemon, ok := pokedex[args]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Println("Name:", pokemon.Name)
	fmt.Println("Height:", pokemon.Height)
	fmt.Println("Weight:", pokemon.Weight)
	return nil
}

func commandPokedex(config *Config, args string) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range pokedex {
		fmt.Println(pokemon.Name)
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
var pokedex map[string]Pokemon

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
		"explore": {
			name:        "explore",
			description: "Explore a area to view the availible pokemon",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inpect",
			description: "inpect a pokemon in pokedex",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "View all the pokemons in your pokedex",
			callback:    commandPokedex,
		},
	}

	pokedex = map[string]Pokemon{}

	scanner := bufio.NewScanner(os.Stdin)
	config := Config{
		Next:     LOCATION_AREA_URL,
		Previous: nil,
		Cache:    pokecache.NewCache(90 * time.Second),
	}

	for {
		fmt.Print("Pokedex > ")

		if scanner.Scan() {
			cleanedInput := cleanInput(scanner.Text())
			command, ok := cliCommands[cleanedInput[0]]

			var args string
			if len(cleanedInput) > 1 {
				args = cleanedInput[1]
			}

			if !ok {
				fmt.Println("Unknown command")
				continue
			}
			if err := command.callback(&config, args); err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Utility Functions
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if item == v {
			return true
		}
	}
	return false
}
