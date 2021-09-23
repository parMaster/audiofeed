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

func index(w http.ResponseWriter, r *http.Request) {

	// move to config
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

func info(w http.ResponseWriter, r *http.Request) {

	html := fmt.Sprintf("%v", r)

	w.Write([]byte(html))
}

func title(w http.ResponseWriter, r *http.Request) {

	rw := NewResponse()

	chapterTemplate := template.New("Chapter")
	chapterTemplate.Parse(chapterTemplateBody)

	xmlTemplate := template.New("Title with chapters")
	xmlTemplate.Parse(xmlTemplateBody)

	params := mux.Vars(r)
	titleName := params["name"]
	titleURL := "http://" + r.Host + "/title/" + titleName // fix

	// move to config
	err := filepath.WalkDir("./audio/"+titleName+"/", func(path string, entry fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if !entry.IsDir() {

			var isChapter = regexp.MustCompile(`(?im)\.(mp3|m4a|m4b)$`)
			if isChapter.MatchString(path) {
				log.Println("visited: ", path)

				chapterTemplate.Execute(rw, map[string]string{
					"ChapterURL": "http://" + r.Host + "/" + path, // move to config
					"TitleURL":   titleURL,
				})
			}

			var isCover = regexp.MustCompile(`(?im)\.(jpg|jpeg|png)$`)
			if isCover.MatchString(path) {
				cover := path
				log.Println("cover found: ", cover)
			}
		}

		return nil
	})
	chapters := rw.body
	rw.Clear()

	xmlTemplate.Execute(rw, map[string]string{
		"Chapters":  chapters,
		"TitleURL":  titleURL,
		"TitleName": titleName,
	})

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
