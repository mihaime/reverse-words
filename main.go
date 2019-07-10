package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Input struct {
	Word string `json:"word,omitempty"`
}
type Output struct {
	ReverseWord string `json:"reverse_word,omitempty"`
}

var (
	totalWordsReversed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_reversed_words",
			Help: "Total number of reversed words",
		},
	)
)

var (
	endpointsAccessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "endpoints_accessed",
			Help: "Total number of accessed to a given endpoint",
		},
		[]string{"endpoint"},
	)
)

// ReturnRelease returns the release configured by the user
func ReturnRelease(w http.ResponseWriter, r *http.Request) {
	release := getEnv("RELEASE", "NotSet")
	releaseString := "Reverse Words Release: " + release
	w.Write([]byte(releaseString))
	endpointsAccessed.WithLabelValues("release").Inc()
}

//ReturnHealth returns healthy string, can be used for monitoring pourposes
func ReturnHealth(w http.ResponseWriter, r *http.Request) {
	health := "Healthy"
	w.Write([]byte(health))
	endpointsAccessed.WithLabelValues("health").Inc()
}

//ReverseWord returns a reversed word based on an input word
func ReverseWord(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var word Input
	var reverseWord string
	err := decoder.Decode(&word)
	if err != nil {
		// Error EOF means no json data has been sent
		if err.Error() != "EOF" {
			panic(err)
		}
	}
	if len(word.Word) < 1 {
		log.Println("No word detected, sending default reverse word")
		reverseWord = "detceted drow oN"
	} else {
		log.Println("Detected word", word.Word)
		reverseWord = reverse(word.Word)
	}
	log.Println("Reverse word", reverseWord)
	totalWordsReversed.Inc()
	output := Output{reverseWord}
	js, err := json.Marshal(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	endpointsAccessed.WithLabelValues("reverseword").Inc()
}

//reverse returns input string reversed
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

//getEnv returns the value for a given Env Var
func getEnv(varName, defaultValue string) string {
	if varValue, ok := os.LookupEnv(varName); ok {
		return varValue
	}
	return defaultValue
}

func main() {
	release := getEnv("RELEASE", "NotSet")
	port := getEnv("APP_PORT", "8080")
	log.Println("Starting Reverse Api. Release:", release)
	log.Println("Listening on port", port)
	prometheus.MustRegister(totalWordsReversed)
	prometheus.MustRegister(endpointsAccessed)
	router := mux.NewRouter()
	router.HandleFunc("/", ReverseWord).Methods("POST")
	router.HandleFunc("/", ReturnRelease).Methods("GET")
	router.HandleFunc("/health", ReturnHealth).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	log.Fatal(http.ListenAndServe(":"+port, router))
}