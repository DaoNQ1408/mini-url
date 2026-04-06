package handler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

// Kết nối Redis khi khởi tạo (Singleton)
func init() {
	opt, _ := redis.ParseURL(os.Getenv("UPSTASH_REDIS_URL"))
	rdb = redis.NewClient(opt)
}

// Hàm Hash URL kết hợp với Salt từ biến môi trường
func hashURL(longURL string, salt string) string {
	data := longURL + salt
	hash := sha256.Sum256([]byte(data))
	// Lấy 7 ký tự đầu của mã Hex làm ID
	return fmt.Sprintf("%x", hash)[:7]
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// 1. LUỒNG REDIRECT (GET /:id)
	path := strings.TrimPrefix(r.URL.Path, "/")
	// Tránh redirect các file tĩnh hoặc api
	if path != "" && path != "api/main" && !strings.Contains(path, ".") && r.Method == "GET" {
		url, err := rdb.Get(ctx, path).Result()
		if err != nil {
			http.Error(w, "URL không tồn tại", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	// 2. LUỒNG RÚT GỌN (POST /)
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
			// Kiểm tra trùng lặp nếu dùng Custom Alias
			exists, _ := rdb.Exists(ctx, customPath).Result()
			if exists > 0 {
				http.Error(w, "Bí danh này đã được sử dụng", http.StatusConflict)
				return
			}
			shortID = customPath
		} else {
			// Tạo ID bằng thuật toán Hash + Salt
			shortID = hashURL(longURL, salt)
		}

		// Lưu vào Redis (Không bao giờ hết hạn)
		err := rdb.Set(ctx, shortID, longURL, 0).Err()
		if err != nil {
			http.Error(w, "Lỗi kết nối Database", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", shortID)
		return
	}
}
