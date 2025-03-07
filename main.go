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

type Concert struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	City      string  `json:"city"`
	Date      string  `json:"date"`
}

var artists []Artist

func getConcertsHandler(w http.ResponseWriter, r *http.Request) {
	concerts := []Concert{
		{48.8566, 2.3522, "Paris", "2025-07-10"},
		{51.5074, -0.1278, "Londres", "2025-07-15"},
		{40.7128, -74.0060, "New York", "2025-07-20"},
	}
	json.NewEncoder(w).Encode(concerts)
}

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
	// Grab sort parameters from the URL (e.g., /home?sortBy=name&order=asc)
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")

	// If sorting parameters exist, sort the artists array
	if sortBy != "" && order != "" {
		sortArtists(sortBy, order)
	}

	// Detect if it's an AJAX request by checking the "X-Requested-With" header
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		tmpl, err := template.New("artists").Parse(`
			<div class="artists">
			{{ range . }}
				<div class="artist">
					<img src="{{ .Image }}" alt="{{ .Name }}">
					<h2>{{ .Name }}</h2>
					<p>Founded: {{ .CreationDate }}</p>
					<a href="/artist?id={{ .ID }}">View Details</a>
				</div>
			{{ end }}
			</div>
		`)
		if err != nil {
			http.Error(w, "Error loading partial template", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, artists)
		return
	}

	// Otherwise, render the full "home.html" template
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Execute the full page template (this includes the search form, sort menu, etc.)
	tmpl.Execute(w, artists)
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

func accueilHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/accueil.html")
	if err != nil {
		http.Error(w, "Error loading accueil.html", http.StatusInternalServerError)
		log.Printf("Error loading accueil.html: %v", err)
		return
	}

	err = tmpl.Execute(w, artists)
	if err != nil {
		http.Error(w, "Error rendering accueil.html", http.StatusInternalServerError)
		log.Printf("Error rendering accueil.html: %v", err)
	}
}

func main() {
	log.Println("Loading data from API...")

	if err := fetchData("artists", &artists); err != nil {
		log.Fatalf("Failed to load artists: %v", err)
	}

	for i, artist := range artists {
		relations, err := fetchRelations(artist.RelationsURL)
		if err != nil {
			log.Printf("Failed to load relations for artist %d: %v", artist.ID, err)
			continue
		}
		artists[i].Relations = relations
	}

	log.Println("Successfully loaded all data from API.")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", accueilHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/artist", artistHandler)
	http.HandleFunc("/search", searchHandler)

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
