package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/ragadeeshu/octo-canyon-freshness-league/datahandling"
)

var iksmMutex sync.Mutex

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/league" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	if r.URL.Path == "/league" {
		results, err := datahandling.GetOrFetchData(&iksmMutex)
		if err != nil {
			fmt.Println(err)
		}

		funcMap := template.FuncMap{
			"indexfunc": func(i int) int {
				return i%10 + 1
			},
			"timedisplay": func(s uint) string {
				min := s / 60
				sec := s % 60
				return fmt.Sprintf("%02d:%02d", min, sec)
			},
		}

		t, err := template.New("league.gohtml").Funcs(funcMap).ParseFiles("./web/league.gohtml")
		if err != nil {
			fmt.Println(err)
		}
		err = t.Execute(w, results)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		body, _ := ioutil.ReadFile("web/index.html")
		fmt.Fprintf(w, "%s", body)
	}
}

func main() {
	go func() {
		log.Println("Starting idle poller")
		index := 0
		for {
			var err error
			fmt.Println("fetching index ", index)
			index, err = datahandling.FetchContestant(&iksmMutex, index)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Hour*8 + time.Minute*time.Duration(rand.Intn(60)))
		}
	}()

	fileServer := http.FileServer(http.Dir("./web/static"))
	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static", fileServer))
	var port string
	if len(os.Args) > 1 {
		port = os.Args[1]
	} else {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
