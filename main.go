package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

const baseURL = "https://groupietrackers.herokuapp.com/api/"

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	RelationsURL string   `json:"relations"`
	Relations    Relations
}

type Relations struct {
	DatesLocations map[string][]string `json:"datesLocations"`
}

var artists []Artist

func fetchData(endpoint string, result interface{}) error {
	url := baseURL + endpoint
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error fetching %s: %v", endpoint, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response from %s: %v", endpoint, err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing JSON from %s: %v", endpoint, err)
	}

	log.Printf("Successfully fetched %s", endpoint)
	return nil
}

func fetchRelations(url string) (Relations, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Relations{}, fmt.Errorf("error fetching relations: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Relations{}, fmt.Errorf("error reading relations response: %v", err)
	}

	var relations Relations
	if err := json.Unmarshal(body, &relations); err != nil {
		return Relations{}, fmt.Errorf("error parsing relations JSON: %v", err)
	}

	return relations, nil
}

// Fonction pour trier les artistes
func sortArtists(sortBy, order string) {
	switch sortBy {
	case "name":
		if order == "asc" {
			sort.Slice(artists, func(i, j int) bool {
				return strings.ToLower(artists[i].Name) < strings.ToLower(artists[j].Name)
			})
		} else if order == "desc" {
			sort.Slice(artists, func(i, j int) bool {
				return strings.ToLower(artists[i].Name) > strings.ToLower(artists[j].Name)
			})
		}
	case "creationDate":
		if order == "asc" {
			sort.Slice(artists, func(i, j int) bool {
				return artists[i].CreationDate < artists[j].CreationDate
			})
		} else if order == "desc" {
			sort.Slice(artists, func(i, j int) bool {
				return artists[i].CreationDate > artists[j].CreationDate
			})
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")

	// Sort the artists if sorting parameters are provided
	if sortBy != "" && order != "" {
		sortArtists(sortBy, order)
	}

	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("Error loading home.html: %v", err)
		return
	}

	err = tmpl.Execute(w, artists)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering home.html: %v", err)
	}
}

func artistHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Artist ID is missing", http.StatusBadRequest)
		return
	}

	var selectedArtist Artist
	found := false
	for _, artist := range artists {
		if fmt.Sprintf("%d", artist.ID) == id {
			selectedArtist = artist
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Artist not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("Error loading artist.html: %v", err)
		return
	}

	err = tmpl.Execute(w, selectedArtist)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering artist.html: %v", err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var results []Artist

	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), strings.ToLower(query)) {
			results = append(results, artist)
		}
	}

	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("Error loading search.html: %v", err)
		return
	}

	err = tmpl.Execute(w, results)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering search.html: %v", err)
	}
}

func main() {
	log.Println("Loading data from API...")

	// Fetch artists data
	if err := fetchData("artists", &artists); err != nil {
		log.Fatalf("Failed to load artists: %v", err)
	}

	// Fetch relations data for each artist
	for i, artist := range artists {
		relations, err := fetchRelations(artist.RelationsURL)
		if err != nil {
			log.Printf("Failed to load relations for artist %d: %v", artist.ID, err)
			continue
		}
		artists[i].Relations = relations
	}

	log.Println("Successfully loaded all data from API.")

	// Set up routes and start the server
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/artist", artistHandler)
	http.HandleFunc("/search", searchHandler)

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
