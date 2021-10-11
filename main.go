package main

import (
	"flag"
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

func (Feed *feed) readMediaFolder() []string {
	var titles []string

	eTitles, err := filepath.Glob("./" + Feed.MediaFolder + "/*")
	check(err)

	for _, t := range eTitles {
		title := filepath.Base(t)
		titles = append(titles, title)
	}
	return titles
}

func (Feed *feed) index(w http.ResponseWriter, r *http.Request) {
	rw := NewResponse()

	titlesTemplate := template.New("Title with chapters")
	titlesTemplate.Parse(titlesTemplateBody)

	titlesTemplate.Execute(rw, Feed.readMediaFolder())
	w.Write([]byte(rw.body))
}

type feed struct {
	HostName    string
	TitleName   string
	MediaFolder string
	TitlePath   string
	CoverPath   string
	Chapters    []string
}

func (Feed *feed) readTitle() {

	var isChapter = regexp.MustCompile(`(?im)\.(mp3|m4a|m4b)$`)
	var isCover = regexp.MustCompile(`(?im)\.(jpg|jpeg|png)$`)

	var chapters []string
	err := filepath.WalkDir("./"+Feed.MediaFolder+"/"+Feed.TitleName+"/", func(path string, entry fs.DirEntry, err error) error {
		check(err)

		if !entry.IsDir() {
			if isChapter.MatchString(path) {
				chapters = append(chapters, path)
				log.Println("visited: ", path)
			} else if isCover.MatchString(path) {
				Feed.CoverPath = path
				log.Println("cover found: ", Feed.CoverPath)
			}
		}
		return nil
	})
	check(err)

	Feed.Chapters = chapters
}

func (Feed *feed) title(w http.ResponseWriter, r *http.Request) {

	rw := NewResponse()

	xmlTemplate := template.New("Title with chapters")
	xmlTemplate.Parse(xmlTemplateBody)

	params := mux.Vars(r)
	Feed.TitleName = params["name"]
	Feed.HostName = r.Host
	Feed.TitlePath = "/title/" + Feed.TitleName

	Feed.readTitle()

	xmlTemplate.Execute(rw, Feed)

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

func getMediaFolderCmd(folder string) string {
	if folder != "" {
		return folder
	} else {
		var cmdFolder string
		flag.StringVar(&cmdFolder, "folder", "audio", "Name of a folder with media")
		flag.Parse()
		return cmdFolder
	}
}

func main() {
	var Feed feed

	r := mux.NewRouter()
	r.HandleFunc("/index", Feed.index).Methods("GET")
	r.HandleFunc("/info", info).Methods("GET")
	r.HandleFunc("/feed.xsl", stylesheet).Methods("GET")
	r.HandleFunc("/title/{name}", Feed.title).Methods("GET")

	Feed.MediaFolder = getMediaFolderCmd("")

	http.Handle("/", r)
	http.Handle("/"+Feed.MediaFolder+"/", http.StripPrefix("/"+Feed.MediaFolder+"/", http.FileServer(http.Dir("./"+Feed.MediaFolder))))

	log.Println("Listening...")
	err := http.ListenAndServe(":8080", nil)
	check(err)
}
