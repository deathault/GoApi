package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Item struct {
	Name  string `json:"name"`
	Size  int    `json:"size"`
	Price int    `json:"price"`
	Year  int    `json:"year"`
	SKU   int    `json:"sku"`
}

var items []Item

func main() {
	file, err := os.ReadFile("data.json")
	if err != nil {
		log.Fatalf("Не удалось открыть data.json: %v", err)
	}
	if err := json.Unmarshal(file, &items); err != nil {
		log.Fatalf("Не удалось распарсить JSON: %v", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/items", apiHandler)

	log.Println("Сервер запущен на порту 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Ошибка при выполнении шаблона", http.StatusInternalServerError)
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(r.URL.Query().Get("name"))
	minPrice, _ := strconv.Atoi(r.URL.Query().Get("minPrice"))
	maxPrice, _ := strconv.Atoi(r.URL.Query().Get("maxPrice"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))

	filtered := []Item{}
	for _, item := range items {
		if name != "" && !strings.Contains(strings.ToLower(item.Name), name) {
			continue
		}
		if minPrice != 0 && item.Price < minPrice {
			continue
		}
		if maxPrice != 0 && item.Price > maxPrice {
			continue
		}
		if year != 0 && item.Year != year {
			continue
		}
		filtered = append(filtered, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}
