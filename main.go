package main

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	Name  string `json:"name"`
	Size  int    `json:"size"`
	Price int    `json:"price"`
	Year  int    `json:"year"`
	SKU   int    `json:"sku"`
}

var items []Item

// --- MAIN ---
func main() {
	file, err := ioutil.ReadFile("data.json")
	if err != nil {
		log.Fatalf("Не удалось открыть data.json: %v", err)
	}

	if err := json.Unmarshal(file, &items); err != nil {
		log.Fatalf("Не удалось распарсить JSON: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/api/items", apiHandler)

	// Оборачиваем mux в middleware для логирования
	loggedMux := loggingMiddleware(mux)

	log.Println("Сервер запущен на порту 3000")
	log.Fatal(http.ListenAndServe(":3000", loggedMux))
}

// --- HANDLERS ---
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

// --- LOGGING MIDDLEWARE ---
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

		// --- IP ---
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		// --- Заголовки ---
		headers, _ := json.MarshalIndent(r.Header, "", "  ")

		// --- Query params ---
		query, _ := json.MarshalIndent(r.URL.Query(), "", "  ")

		// --- Тело запроса (если есть) ---
		var body string
		if r.Body != nil {
			data, _ := io.ReadAll(r.Body)
			body = string(data)
			// Важно: чтобы другие хендлеры могли снова читать тело, его надо вернуть
			r.Body = io.NopCloser(strings.NewReader(body))
		}

		// --- Передаём запрос дальше ---
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		log.Printf(`
----------------------------------------------------
🕒 Время: %v
📡 Метод: %s
🌍 URL: %s
🔢 Статус: %d
⏱️ Длительность: %v
👤 IP: %s
🧭 User-Agent: %s
🧩 Query параметры: %s
📬 Заголовки: %s
📦 Тело запроса: %s
----------------------------------------------------
`, time.Now().Format("2006-01-02 15:04:05"),
			r.Method, r.URL.String(), lrw.statusCode, duration,
			ip, r.UserAgent(), query, headers, body)
	})
}
