package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/joho/godotenv/autoload"
	uuid "github.com/satori/go.uuid"
)

// CSPReport is a single report according to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy-Report-Only
type CSPReport struct {
	ID                 string    `gorm:"primary_key" json:"_id"`
	BlockedURI         string    `json:"blocked-uri"`
	Disposition        string    `json:"disposition"`
	DocumentURI        string    `json:"document-uri"`
	EffectiveDirective string    `json:"effective-directive"`
	OriginalPolicy     string    `json:"original-policy"`
	Referrer           string    `json:"referrer"`
	ScriptSample       string    `json:"script-sample"`
	StatusCode         int       `json:"status-code"`
	ViolatedDirective  string    `json:"violated-directive"`
	CreatedAt          time.Time `gorm:"index" json:"timestamp"`
}

// BeforeCreate sets a unique UUIDv4 for the report
func (report *CSPReport) BeforeCreate(scope *gorm.Scope) (err error) {
	scope.SetColumn("ID", uuid.NewV4().String())
	return nil
}

var databaseURL string
var db *gorm.DB

func init() {
	var ok bool

	databaseURL, ok = os.LookupEnv("DATABASE_URL")
	if ok {
		return
	}

	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgServer := os.Getenv("POSTGRES_SERVER")
	pgPort := os.Getenv("POSTGRES_PORT")
	pgDatabase := os.Getenv("POSTGRES_DB")

	databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", pgUser, pgPassword, pgServer, pgPort, pgDatabase)
}

func loggingMiddleware(h http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, h)
}

func limiterMiddleware(h http.Handler) http.Handler {
	lmt := tollbooth.NewLimiter(10, nil)

	return tollbooth.LimitFuncHandler(lmt, h.ServeHTTP)
}

func responseWithError(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	w.Write([]byte(fmt.Sprintf("{\"error\":\"%v\"}", err)))
}

func insertReportHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var report struct {
		CSPReport CSPReport `json:"csp-report"`
	}
	err := decoder.Decode(&report)
	if err != nil {
		responseWithError(err, w)
		return
	}

	err = db.Create(&report.CSPReport).Error
	if err != nil {
		responseWithError(err, w)
		return
	}

	w.WriteHeader(204)
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	var err error
	db, err = gorm.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.LogMode(true)
	db.AutoMigrate(&CSPReport{})

	router := mux.NewRouter()
	router.Use(loggingMiddleware, limiterMiddleware)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	router.HandleFunc("/csp-reports", insertReportHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), handlers.RecoveryHandler()(router)))
}
