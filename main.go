package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"pokedex/internal/pokecache"
	"strings"
	"time"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	next     string
	previous string
}

type Response struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []LocationArea `json:"results"`
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationAreaResponse struct {
	Pokemon []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name    string `json:"name"`
	BaseExp int    `json:"base_experience"`
	Height  int    `json:"height"`
	Weight  int    `json:"weight"`
	Stats   []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	if cfg != nil {
		fmt.Printf("Last seen page: next = %q, previous = %q\n", cfg.next, cfg.previous)
	}
	os.Exit(0)
	return nil // unreachable, but required for the signature
}

func checkCache(url string) ([]byte, bool) {
	// check cache before sending requests
	cacheResponse, ok := cache.Get(url)
	if !ok {
		return nil, ok
	}
	return cacheResponse, ok
}

func commandMap(cfg *config) error {
	// https://pokeapi.co/api/v2/location-area/{id or name}/
	if cfg.next == "" {
		fmt.Println(("You're on the last page of results."))
		return nil
	}
	// If cache has data
	if cachedData, ok := checkCache(cfg.next); ok {
		// Parse the cached data
		var locData Response
		err := json.Unmarshal(cachedData, &locData)
		if err != nil {
			fmt.Println("Error parsing cached data:", err)
			return err
		}

		// Update config and display results
		cfg.previous = cfg.next
		cfg.next = locData.Next

		for _, res := range locData.Results {
			fmt.Println(res.Name)
		}
		fmt.Println("cache used")
		return nil
	}

	res, err := http.Get(cfg.next)
	if err != nil {
		fmt.Println(err)
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return err
	}
	if err != nil {
		fmt.Println(err)
	}
	var locData Response
	json.Unmarshal(body, &locData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Add response to cache
	cache.Add(cfg.next, body)

	// sort cfg next and
	cfg.previous = cfg.next
	cfg.next = locData.Next

	for _, res := range locData.Results {
		fmt.Println(res.Name)
	}
	return nil
}

func commandMapB(cfg *config) error {
	// https://pokeapi.co/api/v2/location-area/{id or name}/
	if cfg.previous == "" {
		fmt.Println(("You're on the first page of results."))
		return nil
	}
	// If cache has data
	if cachedData, ok := checkCache(cfg.next); ok {
		// Parse the cached data
		var locData Response
		err := json.Unmarshal(cachedData, &locData)
		if err != nil {
			fmt.Println("Error parsing cached data:", err)
			return err
		}

		// Update config and display results
		cfg.previous = cfg.next
		cfg.next = locData.Next

		for _, res := range locData.Results {
			fmt.Println(res.Name)
		}
		fmt.Println("cache used")
		return nil
	}

	res, err := http.Get(cfg.previous)
	if err != nil {
		fmt.Println(err)
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return err
	}
	if err != nil {
		fmt.Println(err)
	}
	var locData Response
	json.Unmarshal(body, &locData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Add response to cache
	cache.Add(cfg.next, body)

	// sort cfg next and
	cfg.previous = locData.Previous
	cfg.next = locData.Next

	for _, res := range locData.Results {
		fmt.Println(res.Name)
	}
	return nil
}

func commandHelp(cfg *config) error {

	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range registry {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	if cfg != nil {
		fmt.Printf("Last seen page: next = %q, previous = %q\n", cfg.next, cfg.previous)
	}
	return nil
}

func commandExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("Please provide a location area name")
		return nil
	}
	locationArea := args[0]

	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", locationArea)
	// If cache has data
	if cachedData, ok := checkCache(url); ok {
		// Parse the cached data
		var locData Response
		err := json.Unmarshal(cachedData, &locData)
		if err != nil {
			fmt.Println("Error parsing cached data:", err)
			return err
		}

		for _, res := range locData.Results {
			fmt.Println(res.Name)
		}
		fmt.Println("cache used")
		return nil
	}

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		return err
	}
	if err != nil {
		fmt.Println(err)
	}
	var locationData LocationAreaResponse
	err = json.Unmarshal(body, &locationData)
	if err != nil {
		fmt.Println("Error parsing location data:", err)
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range locationData.Pokemon {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *config, args []string) error {

	// https://pokeapi.co/api/v2/pokemon/{id or name}/

	if len(args) == 0 {
		fmt.Println("Please provide a location area name")
		return nil
	}
	pokemonName := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// check pokedex first
	_, ok := Pokedex[pokemonName]
	if ok {
		fmt.Printf("%s already caught!!\n", pokemonName)
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	var pokemonToCatch Pokemon

	// If cache has data
	if cachedData, ok := checkCache(url); ok {
		err := json.Unmarshal(cachedData, &pokemonToCatch)
		if err != nil {
			fmt.Println("Error parsing cached data:", err)
			return err
		}

	} else {
		res, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return err
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			fmt.Printf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
			return err
		}
		if err != nil {
			fmt.Println(err)
		}

		err = json.Unmarshal(body, &pokemonToCatch)
		if err != nil {
			fmt.Println("Error parsing data:", err)
			return err
		}
		cache.Add(url, body)

	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate catch chance (0-100)
	// The higher the base experience, the lower the catch chance
	catchChance := 100 - (pokemonToCatch.BaseExp / 6)

	// Make sure we have a reasonable minimum chance (e.g., 5%)
	if catchChance < 5 {
		catchChance = 5
	}

	if r.Intn(100) < catchChance {
		fmt.Printf("%s was caught!\n", pokemonName)
		// Add to pokedex
		Pokedex[pokemonName] = pokemonToCatch

	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}
	return nil
}

func commandInspect(cli *config, args []string) error {
	if len(args) == 0 {
		fmt.Println("Please provide a location area name")
		return nil
	}
	pokemonName := args[0]

	entry, ok := Pokedex[pokemonName]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	// Print the Pokemon details in the required format
	fmt.Printf("Name: %s\n", entry.Name)
	fmt.Printf("Height: %d\n", entry.Height)
	fmt.Printf("Weight: %d\n", entry.Weight)

	fmt.Println("Stats:")
	for _, stat := range entry.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, pokemonType := range entry.Types {
		fmt.Printf("  - %s\n", pokemonType.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *config, args []string) error {
	fmt.Println("Your Pokedex:")
	for pokemonName, _ := range Pokedex {
		fmt.Printf("- %s\n", pokemonName)
	}
	return nil
}

// make global variable
var registry map[string]cliCommand
var cache = pokecache.NewCache(3)
var Pokedex = make(map[string]Pokemon)

func main() {
	// Create config to pass to each command
	cfg := config{
		next:     "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		previous: "",
	}

	registry = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    func(cfg *config, args []string) error { return commandExit(cfg) },
		},
	}
	// Now add the help command after registry is defined
	registry["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    func(cfg *config, args []string) error { return commandHelp(cfg) },
	}

	registry["map"] = cliCommand{
		name:        "map",
		description: "Get's the next 20 locationa areas",
		callback:    func(cfg *config, args []string) error { return commandMap(cfg) },
	}

	registry["mapb"] = cliCommand{
		name:        "mapb",
		description: "Get's the previous 20 locationa areas",
		callback:    func(cfg *config, args []string) error { return commandMapB(cfg) },
	}

	registry["explore"] = cliCommand{
		name:        "explore",
		description: "Explore a location-area",
		callback:    commandExplore,
	}

	registry["catch"] = cliCommand{
		name:        "catch",
		description: "Catch a pokemon",
		callback:    commandCatch,
	}

	registry["inspect"] = cliCommand{
		name:        "inspect",
		description: "Inspect a pokemon#s stat",
		callback:    commandInspect,
	}

	registry["pokedex"] = cliCommand{
		name:        "pokdex",
		description: "Show caught Pokemon",
		callback:    commandPokedex,
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		line := scanner.Text()
		cleaned := cleanInput(line)

		// Check if input is not empty
		if len(cleaned) == 0 {
			continue
		}

		command, ok := registry[cleaned[0]]

		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		// Pass the command arguments (everything after the command name)
		args := []string{}
		if len(cleaned) > 1 {
			args = cleaned[1:]
		}

		err := command.callback(&cfg, args)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func cleanInput(text string) []string {
	// split by whitespace, lowercase and trim trailing whspace
	t := strings.TrimSpace(text)
	tLower := strings.ToLower(t)
	return strings.Split(tLower, " ")
}
