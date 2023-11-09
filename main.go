package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

// PageData struct holds the data to be passed to the template.
type PageData struct {
	AsciiArt string
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderErrorPage(w, http.StatusNotFound)
		return
	}
	renderTemplate(w, "templates/main.html", nil)
}

func handleAsciiArt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderErrorPage(w, http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	banner := r.FormValue("banner")
	result := generateAsciiArt(text, banner)

	data := PageData{
		AsciiArt: result,
	}

	renderTemplate(w, "templates/main.html", data)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	asciiArt := r.URL.Query().Get("data")
	if asciiArt == "" {
		renderErrorPage(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprint(len(asciiArt)))
	w.Header().Set("Content-Disposition", `attachment; filename="ascii_art.txt"`)

	if _, err := w.Write([]byte(asciiArt)); err != nil {
		renderErrorPage(w, http.StatusInternalServerError)
		return
	}
}

func renderTemplate(w http.ResponseWriter, templateFile string, data interface{}) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		renderErrorPage(w, http.StatusNotFound)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		renderErrorPage(w, http.StatusInternalServerError)
	}
}

func generateAsciiArt(text, banner string) string {
	words := strings.Split(text, "\n")
	rawBytes, err := os.ReadFile(fmt.Sprintf("banner/%s.txt", banner))
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(strings.ReplaceAll(string(rawBytes), "\r\n", "\n"), "\n")

	var result strings.Builder
	for i, word := range words {
		if word == "" {
			if i < len(words)-1 {
				result.WriteString("\n")
			}
			continue
		}
		for h := 1; h < 9; h++ {
			for _, l := range word {
				for lineIndex, line := range lines {
					if lineIndex == (int(l)-32)*9+h {
						result.WriteString(line)
					}
				}
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}

func renderErrorPage(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	tmpl, err := template.ParseFiles(fmt.Sprintf("templates/%d.html", status))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, status)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", handleDefault)
	http.HandleFunc("/ascii-art", handleAsciiArt)
	http.HandleFunc("/export", handleExport) // New endpoint for exporting ASCII art

	fmt.Println("Server started. Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
