package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
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
	http.HandleFunc("/artist/", artistHandler)
	http.HandleFunc("/error", errorHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Lancer le serveur
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

// homeHandler affiche la liste paginée des artistes
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/home.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Pagination
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	artistsPerPage := 5
	startIndex := (page - 1) * artistsPerPage
	endIndex := startIndex + artistsPerPage

	if startIndex >= len(artists) {
		startIndex = len(artists)
	}
	if endIndex > len(artists) {
		endIndex = len(artists)
	}

	// Calcul des pages de pagination
	totalPages := (len(artists) + artistsPerPage - 1) / artistsPerPage
	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

	// Calcul des pages précédente et suivante
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	data := struct {
		Artists     []Artist
		CurrentPage int
		TotalPages  int
		Pages       []int
		PrevPage    int
		NextPage    int
	}{
		Artists:     artists[startIndex:endIndex],
		CurrentPage: page,
		TotalPages:  totalPages,
		Pages:       pages,
		PrevPage:    prevPage,
		NextPage:    nextPage,
	}

	tmpl.Execute(w, data)
}

// searchHandler gère la recherche d'artistes
func searchHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/search.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
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

// artistHandler affiche les détails d'un artiste spécifique
func artistHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/artist.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 || id > len(artists) {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	artist := artists[id-1]
	tmpl.Execute(w, artist)
}

// errorHandler affiche une page d'erreur
func errorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/error.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
