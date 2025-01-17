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
		//"formatArtistName": formatArtistName, // Ajouter cette fonction
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
	// On ne gère plus la pagination, on affiche tous les artistes
	data := struct {
		Artists []Artist
		Query   string
	}{
		Artists: artists, // Afficher tous les artistes, sans pagination
		Query:   "",      // Valeur vide car il n'y a pas de recherche ici
	}

	// Exécuter le template avec les artistes récupérés
	err := tmpl.ExecuteTemplate(w, "home.html", data)
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

// // Fonction pour générer l'URL de l'image à partir du nom de l'artiste
// func (a *Artist) GetImageUrl() string {
// 	// Remplacer les espaces par des underscores et mettre en minuscule
// 	imageName := strings.ToLower(strings.ReplaceAll(a.Name, " ", "_"))
// 	return fmt.Sprintf("https://groupietrackers.herokuapp.com/api/images/%s.jpeg", imageName)
// }

// // Fonction pour formater le nom de l'artiste en une URL valide
// func formatArtistName(name string) string {
// 	// Remplacer les espaces par des tirets et tout mettre en minuscule
// 	name = strings.ToLower(name)
// 	name = strings.ReplaceAll(name, " ", "")
// 	return name
// }
