package main
import (
	"fmt"
	"bufio"
	"strings"
	"os"
	"net/http"
	"encoding/json"
	"github.com/glitchdawg/pokedex/internal/pokecache"
	"time"
	"math/rand"
)



type config struct {
	Next string
	Previous interface{}
	Cache *pokecache.Cache
	Pokedex map[string]Pokemon
}
type cliCommand struct {
	name        string
	description string
	callback    func(*config,[]string) error
}


func fetchLocations(url string,c *config) (LocationStruct, error) {
	locations := LocationStruct{}

	if cachedData, ok := c.Cache.Get(url); ok {
		err := json.Unmarshal(cachedData, &locations)
        if err == nil {
            return locations, nil
        }
	}

	res,err:=http.NewRequest("GET",url,nil)
	if err != nil {
		return locations,fmt.Errorf("failed to create request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(res)
	if err != nil {
		return locations,fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return locations,fmt.Errorf("failed to get response: %v", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	
	
	err = decoder.Decode(&locations)
	if err != nil {
		return locations,fmt.Errorf("failed to decode response: %v", err)
	
	}
	responseBytes, _ := json.Marshal(locations)
    c.Cache.Add(url, responseBytes)
	return locations, nil
}

func CleanInput(text string) []string{
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	words := strings.Fields(text)
	return words
}

func commandExit(c *config, args []string) error{
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}
func commandHelp(c *config, args []string) error{
    fmt.Println(`Welcome to the Pokedex!
Usage:

help: Displays a help message
exit: Exit the Pokedex
map: Display the names of 20 location areas in the Pokemon world
mapb: Display the previous 20 location areas`)
    return nil
}
func commandMap(c *config, args []string) error{
	var url string
	if c.Next != "" {
		url = fmt.Sprintf("%s", c.Next)
	}else{
		url = "https://pokeapi.co/api/v2/location-area/"
	}
	locations, err := fetchLocations(url,c)
	if err != nil {
		return fmt.Errorf("failed to fetch locations: %v", err)
	}
	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	c.Next = locations.Next
	c.Previous = locations.Previous
	return nil
}
func commandMapb(c *config, args []string) error{
	if c.Previous == nil {
        fmt.Println("you're on the first page")
        return nil
    }
	prevURL, ok := c.Previous.(string)
    if !ok {
        return fmt.Errorf("previous URL is not a string")
    }
    
    url := prevURL
	locations, err := fetchLocations(url,c)
	if err != nil {
		return fmt.Errorf("failed to fetch locations: %v", err)
	}
	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	c.Next = locations.Next
	c.Previous = locations.Previous
	return nil
}
func traverseLocations(locationData ExploredLocation,area string) error{
	fmt.Printf("Exploring %s...\n", area)
	fmt.Println("Found Pokemon:")
	for _, encounter := range locationData.PokemonEncounters {
		fmt.Printf("- %s\n", encounter.Pokemon.Name)
	}
	return nil
}
func commandExplore(c *config, args []string) error{
	if len(args) < 2 {
		return fmt.Errorf("please provide a location area name")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments, please provide only one location area name")
	}
	if args[1] == "" {
		return fmt.Errorf("please provide a valid location area name")
	}
	locationData := ExploredLocation{}
	area:=args[1]
	
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", area)
	if cachedData, ok := c.Cache.Get(url); ok {
		err := json.Unmarshal(cachedData, &locationData)
		if err == nil {
			err=traverseLocations(locationData, area)
			if err != nil {
				return fmt.Errorf("failed to traverse locations: %v", err)
			}
			return nil
		}
	}
	res, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(res)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get response: %v", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&locationData)
	if err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	err=traverseLocations(locationData, area)
	if err != nil {
			return fmt.Errorf("failed to traverse locations: %v", err)
	}
	responseBytes, _ := json.Marshal(locationData)
	c.Cache.Add(url, responseBytes)
	return nil
}

func catchPokemon(c *config,pokemon Pokemon) error {
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	rand.Seed(time.Now().UnixNano())
	catchChance := rand.Float64()
	catchRate := 1.0 - (float64(pokemon.BaseExperience) / 1000.0)
	if catchChance < catchRate {
		fmt.Printf("%s was caught!\n", pokemon.Name)
		c.Pokedex[pokemon.Name] = pokemon
	} else {
		fmt.Printf("%s excaped!\n", pokemon.Name)
	}
	return nil
}
func commandCatch(c *config, args []string) error{
	if len(args) < 2 {
		return fmt.Errorf("please provide a pokemon name")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments, please provide only one pokemon name")
	}
	if args[1] == "" {
		return fmt.Errorf("please provide a valid pokemon name")
	}
	pokemonName := args[1]
	pokemonInfo := Pokemon{}
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonName)
	if cachedData, ok := c.Cache.Get(url); ok {
		err := json.Unmarshal(cachedData, &pokemonInfo)
		if err == nil {
			err=catchPokemon(c,pokemonInfo)
			if err != nil {
				return fmt.Errorf("failed to catch pokemon: %v", err)
			}
			return nil
		}
	}
	res, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(res)	
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get response: %v", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&pokemonInfo)
	if err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	err=catchPokemon(c,pokemonInfo)
	if err != nil {
		return fmt.Errorf("failed to catch pokemon: %v", err)
	}
	responseBytes, _ := json.Marshal(pokemonInfo)
	c.Cache.Add(url, responseBytes)
	return nil
}

func commandInspect(c *config, args []string) error{
	if len(args) < 2 {
		return fmt.Errorf("please provide a pokemon name")
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments, please provide only one pokemon name")
	}
	if args[1] == "" {
		return fmt.Errorf("please provide a valid pokemon name")
	}
	pokemonName := args[1]
	if pokemonInfo,ok:= c.Pokedex[pokemonName]; ok {
		fmt.Printf("Name: %s\n", pokemonInfo.Name)
		fmt.Printf("Height: %d\n", pokemonInfo.Height)
		fmt.Printf("Weight: %d\n", pokemonInfo.Weight)
		fmt.Printf("Stats:\n")
		for _, stat := range pokemonInfo.Stats {
			fmt.Printf("  %s: %d\n", stat.Stat.Name, stat.BaseStat)
		}
		fmt.Printf("Types:\n")
		for _, t := range pokemonInfo.Types {
			fmt.Printf("  %s\n", t.Type.Name)
		}		

	} else {
		return fmt.Errorf("pokemon not found in your Pokedex")
	}
	return nil
}

func commandPokedex(c *config, args []string) error{
	if len(c.Pokedex) == 0 {
		fmt.Println("Your Pokedex is empty.")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for name, _ := range c.Pokedex {
		fmt.Printf("- %s\n", name)
	}
	return nil
}

func main() {
	cache:=pokecache.NewCache(5*time.Minute)
	scanner := bufio.NewScanner(os.Stdin)
	cfg := &config{
		Cache: cache,
		Pokedex: make(map[string]Pokemon),
	}
	commands:=map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help":{
			name:        "help",
			description: "Show help information",
			callback:    commandHelp,
		},
		"map":{
			name: "map",
			description: "Show the map of the region",
			callback: commandMap,
		},
		"mapb":{
			name: "mapb",
			description: "go back to previous map of the region",
			callback: commandMapb,
		},
		"explore":{
			name: "explore",
			description: "Explore the pokemons of the region",
			callback: commandExplore,
		},
		"catch":{
			name: "catch",
			description: "Catch a pokemon",
			callback: commandCatch,
		},
		"inspect":{
			name: "inspect",
			description: "Inspect a pokemon",
			callback: commandInspect, 
		},
		"pokedex":{
			name: "pokedex",
			description: "Show the pokedex",
			callback: commandPokedex,
		},

	}
	
	for {
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			text := scanner.Text()
			words := CleanInput(text)
			if len(words) == 0 {
				continue
			}
			if command:= commands[words[0]]; command.name != "" {
				
				if err := command.callback(cfg,words); err != nil {
					fmt.Println("Error executing command:", err)
				}
			} else {
				fmt.Println("Unknown command:", words[0])
			}
		}
	
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
	}
}