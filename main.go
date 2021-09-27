package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gorilla/mux"
)

const AudioFolder string = "audio"

func index(w http.ResponseWriter, r *http.Request) {

	eTitles, err := filepath.Glob("./" + AudioFolder + "/*")
	if err != nil {
		panic(err.Error())
	}

	// ToDo: use html template
	html := ""
	for _, t := range eTitles {
		title := filepath.Base(t)
		html += fmt.Sprintf("<a href=\"/title/%v\">%v</a><br>\n", title, title)
	}
	w.Write([]byte(html))
}

func info(w http.ResponseWriter, r *http.Request) {

	html := fmt.Sprintf("%v", r)

	w.Write([]byte(html))
}

type feed struct {
	TitleName   string
	TitleURL    string
	CoverURL    string
	ChapterURLs []string
}

func title(w http.ResponseWriter, r *http.Request) {

	rw := NewResponse()
	var Feed feed

	xmlTemplate := template.New("Title with chapters")
	xmlTemplate.Parse(xmlTemplateBody)

	params := mux.Vars(r)
	Feed.TitleName = params["name"]
	Feed.TitleURL = "http://" + r.Host + "/title/" + Feed.TitleName

	var isChapter = regexp.MustCompile(`(?im)\.(mp3|m4a|m4b)$`)
	var isCover = regexp.MustCompile(`(?im)\.(jpg|jpeg|png)$`)

	err := filepath.WalkDir("./"+AudioFolder+"/"+Feed.TitleName+"/", func(path string, entry fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if !entry.IsDir() {

			if isChapter.MatchString(path) {

				Feed.ChapterURLs = append(Feed.ChapterURLs, "http://"+r.Host+"/"+path)
				log.Println("visited: ", path)

			} else if isCover.MatchString(path) {

				Feed.CoverURL = "http://" + r.Host + "/" + path
				log.Println("cover found: ", Feed.CoverURL)
			}
		}

		return nil
	})

	xmlTemplate.Execute(rw, Feed)

	if err != nil {
		log.Fatal("error walking audio path: \n", err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write([]byte(rw.body))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/index", index).Methods("GET")
	r.HandleFunc("/info", info).Methods("GET")
	r.HandleFunc("/title/{name}", title).Methods("GET")

	http.Handle("/", r)

	log.Println("Listening...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err.Error())
	}
}
