package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func index(w http.ResponseWriter, r *http.Request) {

	eTitles, err := filepath.Glob("./audio/*")
	if err != nil {
		panic(err.Error())
	}

	html := ""
	for _, t := range eTitles {
		title := filepath.Base(t)
		html += fmt.Sprintf("<a href=\"/title/%v\">%v</a><br>\n", title, title)
	}
	w.Write([]byte(html))
}

func title(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	name := params["name"]

	eChapters, err := filepath.Glob("./audio/" + name + "/*")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	xml := ""
	for _, c := range eChapters {
		xml += fmt.Sprintf("<b>%v</b><br>\n", c)
	}
	w.Header().Add("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(xml))
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/index", index).Methods("GET")
	r.HandleFunc("/title/{name}", title).Methods("GET")

	http.Handle("/", r)

	log.Println("Listening...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err.Error())
	}
}
