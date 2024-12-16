package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	err := fetchArtists()
	if err != nil {
		fmt.Println("Error fetching artists:", err)
		return
	}
	http.HandleFunc("/", homeHandler)
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

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
func homeHandler(w http.ResponseWriter, r *http.Request) {
	for _, artist := range artists {
		fmt.Fprintf(w, "<p>%s (%d)</p>", artist.Name, artist.CreationDate)
	}
}
