package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/ragadeeshu/octo-canyon-freshness-league/datahandling"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/league" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	results, err := datahandling.GetOrFetchData()
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

	t, err := template.New("template.gohtml").Funcs(funcMap).ParseFiles("./web/template.gohtml")
	if err != nil {
		fmt.Println(err)
	}
	err = t.Execute(w, results)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	// results, err := datahandling.GetOrFetchData()
	// _, err := datahandling.GetOrFetchData()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	fileServer := http.FileServer(http.Dir("./web/static"))
	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static", fileServer))
	log.Fatal(http.ListenAndServe(":8080", nil))

	// output, _ := json.MarshalIndent(results, "", "\t")
	// fmt.Println(string(output))

}
