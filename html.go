package main

import (
	"html/template"
	"log"
	"net/http"
)

// NavLink represents a navigation menu item
type NavLink struct {
	Text string
	URL  string
}

// PageData holds all the data needed for our template
type PageData struct {
	Title    string
	NavLinks []NavLink
	Content  string
}

func HTML() {
	// Parse the HTML template
	tmpl, err := template.ParseFiles("layout.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Handle the home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Sample data for our template
		data := PageData{
			Title: "Acme Corporation",
			NavLinks: []NavLink{
				{Text: "Home", URL: "/"},
				{Text: "About", URL: "/about"},
				{Text: "Services", URL: "/services"},
				{Text: "Products", URL: "/products"},
				{Text: "Contact", URL: "/contact"},
			},
			Content: "Welcome to Acme Corporation. We provide innovative solutions for all your needs.",
		}

		// Render the template with our data
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Start the web server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
