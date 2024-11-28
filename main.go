package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
)

// APIResponse defines the structure of the API response
type APIResponse struct {
	ID    int    `json:"id"` // ID is now an integer
	Title string `json:"title"`
	Body  string `json:"body"`
}

// fetchData fetches posts from the JSONPlaceholder API
func fetchData() ([]APIResponse, error) {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var data []APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	return data, nil
}

// saveToFirestore saves the fetched data to Firestore
func saveToFirestore(data []APIResponse) error {
	projectID := os.Getenv("FIRESTORE_PROJECT")
	if projectID == "" {
		return fmt.Errorf("FIRESTORE_PROJECT environment variable not set")
	}

	ctx := context.Background()

	// Create Firestore client
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("firestore client creation failed: %v", err)
	}
	defer client.Close()

	// Log emulator usage if applicable
	if emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST"); emulatorHost != "" {
		log.Printf("Using Firestore emulator at %s", emulatorHost)
	}

	// Save data to Firestore
	for _, record := range data {
		docID := strconv.Itoa(record.ID) // Convert int ID to string
		_, err := client.Collection("posts").Doc(docID).Set(ctx, record)
		if err != nil {
			return fmt.Errorf("error saving record ID %d: %v", record.ID, err)
		}
	}
	return nil
}

// FunctionEntryPoint handles HTTP requests and processes data
func FunctionEntryPoint(w http.ResponseWriter, r *http.Request) {
	// Fetch data from the external API
	data, err := fetchData()
	if err != nil {
		log.Printf("Error fetching data: %v", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Save data to Firestore
	if err := saveToFirestore(data); err != nil {
		log.Printf("Error saving data: %v", err)
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Data successfully processed and stored!")
}

// main runs the server locally
func main() {
	http.HandleFunc("/", FunctionEntryPoint)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
