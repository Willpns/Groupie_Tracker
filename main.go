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
	Image        string   `json:"image"`
}

var artists []Artist

// Fonction pour joindre les éléments d'une liste avec un séparateur
func join(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

func main() {
	// Charger les données des artistes
	err := fetchArtists()
	if err != nil {
		fmt.Println("Error fetching artists:", err)
		return
	}

	// Déclarer les fonctions personnalisées
	funcMap := template.FuncMap{
		"join": join,
	}

	// Charger les templates avec les fonctions personnalisées
	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("./templates/*.html"))

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homeHandler(w, r, templates)
	})
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		searchHandler(w, r, templates)
	})
	http.HandleFunc("/artist/", func(w http.ResponseWriter, r *http.Request) {
		artistHandler(w, r, templates)
	})
	http.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		errorHandler(w, r, templates)
	})
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
func homeHandler(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
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

	totalPages := (len(artists) + artistsPerPage - 1) / artistsPerPage
	var pages []int
	for i := 1; i <= totalPages; i++ {
		pages = append(pages, i)
	}

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
		Query       string // Ajout du champ Query
	}{
		Artists:     artists[startIndex:endIndex],
		CurrentPage: page,
		TotalPages:  totalPages,
		Pages:       pages,
		PrevPage:    prevPage,
		NextPage:    nextPage,
		Query:       "", // Valeur vide, car la recherche n'est pas applicable ici
	}

	err = tmpl.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

// searchHandler gère la recherche d'artistes
func searchHandler(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
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

	err := tmpl.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

// artistHandler affiche les détails d'un artiste spécifique
func artistHandler(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	idStr := strings.TrimPrefix(r.URL.Path, "/artist/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 || id > len(artists) {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}

	artist := artists[id-1]
	err = tmpl.ExecuteTemplate(w, "artist.html", artist)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

// errorHandler affiche une page d'erreur
func errorHandler(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	err := tmpl.ExecuteTemplate(w, "error.html", nil)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}
