package main // Giữ package main để chạy được trên GoLand

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	// Nạp file .env nếu chạy local. Trên Vercel không có file này nên nó sẽ bỏ qua.
	_ = godotenv.Load()

	opt, err := redis.ParseURL(os.Getenv("UPSTASH_REDIS_URL"))
	if err == nil {
		rdb = redis.NewClient(opt)
	}
}

// Hàm xử lý chính - Vercel sẽ gọi hàm này (phải viết hoa chữ H)
func Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// 1. Nếu vào trang chủ thì hiện giao diện
	if (path == "" || path == "api/main") && r.Method == "GET" {
		http.ServeFile(w, r, "public/index.html")
		return
	}

	// 2. Luồng REDIRECT
	if path != "" && path != "api/main" && r.Method == "GET" && !strings.Contains(path, ".") {
		url, err := rdb.Get(ctx, path).Result()
		if err != nil {
			http.Error(w, "Link không tồn tại", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	// 3. Luồng RÚT GỌN
	if r.Method == "POST" {
		longURL := r.FormValue("url")
		customPath := strings.TrimSpace(r.FormValue("custom"))
		salt := os.Getenv("URL_SALT")

		if longURL == "" {
			http.Error(w, "Thiếu URL gốc", http.StatusBadRequest)
			return
		}

		var shortID string
		if customPath != "" {
			exists, _ := rdb.Exists(ctx, customPath).Result()
			if exists > 0 {
				http.Error(w, "Bí danh đã tồn tại", http.StatusConflict)
				return
			}
			shortID = customPath
		} else {
			shortID = hashURL(longURL, salt)
		}

		_ = rdb.Set(ctx, shortID, longURL, 0).Err()
		fmt.Fprintf(w, "%s", shortID)
		return
	}
}

// Hàm Hash
func hashURL(longURL string, salt string) string {
	data := longURL + salt
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:7]
}

// HÀM MAIN: Chỉ chạy khi bạn nhấn Run trên GoLand
func main() {
	http.HandleFunc("/", Handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🚀 Local server: http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
