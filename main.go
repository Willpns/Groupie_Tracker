package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	Relations    string   `json:"relations"`
}

var artists []Artist

func main() {
	// Charger les données des artistes
	err := fetchArtists()
	if err != nil {
		fmt.Println("Error fetching artists:", err)
		return
	}

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// fetchArtists récupère les données des artistes depuis l'API
func fetchArtists() error {
	response, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %d", response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(&artists)
	if err != nil {
		return err
	}

	return nil
}

// homeHandler gère la page d'accueil
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Artists []Artist
	}{
		Artists: artists,
	}

	tmpl.Execute(w, data)
}

// searchHandler gère la recherche d'artistes
func searchHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	var results []Artist

	if query != "" {
		for _, artist := range artists {
			if strings.Contains(strings.ToLower(artist.Name), query) {
				results = append(results, artist)
			} else {
				for _, member := range artist.Members {
					if strings.Contains(strings.ToLower(member), query) {
						results = append(results, artist)
						break
					}
				}
			}
		}
	}

	data := struct {
		Query   string
		Results []Artist
	}{
		Query:   query,
		Results: results,
	}

	tmpl.Execute(w, data)
}
