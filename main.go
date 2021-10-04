package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gorilla/mux"
)

const AudioFolder string = "audio"

func index(w http.ResponseWriter, r *http.Request) {
	rw := NewResponse()
	var titles []string

	titlesTemplate := template.New("Title with chapters")
	titlesTemplate.Parse(titlesTemplateBody)

	eTitles, err := filepath.Glob("./" + AudioFolder + "/*")
	check(err)

	for _, t := range eTitles {
		title := filepath.Base(t)
		titles = append(titles, title)
	}

	titlesTemplate.Execute(rw, titles)
	w.Write([]byte(rw.body))
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

		check(err)

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

	w.Header().Add("Content-Type", "text/xml; charset=utf-8")
	w.Write([]byte(rw.body))
}

func info(w http.ResponseWriter, r *http.Request) {
	b := fmt.Sprintf("%v", r)
	w.Write([]byte(b))
}

func stylesheet(w http.ResponseWriter, r *http.Request) {
	b, err := os.ReadFile("./feed.xsl")
	check(err)
	w.Write([]byte(b))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/index", index).Methods("GET")
	r.HandleFunc("/info", info).Methods("GET")
	r.HandleFunc("/feed.xsl", stylesheet).Methods("GET")
	r.HandleFunc("/title/{name}", title).Methods("GET")

	http.Handle("/", r)

	log.Println("Listening...")
	err := http.ListenAndServe(":8080", nil)
	check(err)
}
