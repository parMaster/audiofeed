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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type feedServer struct {
	HostName    string
	MediaFolder string
	Port        string
	title
}

type title struct {
	Name      string
	Path      string
	CoverPath string
	Chapters  []string
}

func (FeedServer *feedServer) index(w http.ResponseWriter, r *http.Request) {
	titlesTemplate := template.New("Title with chapters")
	titlesTemplate.Parse(titlesTemplateBody)

	titlesTemplate.Execute(w, FeedServer.fromMediaFolder(FeedServer.MediaFolder))
}

func (FeedServer *feedServer) displayTitle(w http.ResponseWriter, r *http.Request) {
	xmlTemplate := template.New("Title with chapters")
	xmlTemplate.Parse(xmlTemplateBody)

	params := mux.Vars(r)
	FeedServer.HostName = r.Host
	FeedServer.Name = params["name"] // filter somehow?
	FeedServer.Path = filepath.ToSlash(filepath.Join("title", FeedServer.Name))

	FeedServer.readTitle(filepath.Join(FeedServer.MediaFolder, FeedServer.Name))

	w.Header().Add("Content-Type", "text/xml; charset=utf-8")
	xmlTemplate.Execute(w, FeedServer)
}

func (FeedServer *feedServer) info(w http.ResponseWriter, r *http.Request) {
	b := fmt.Sprintf("%v", r)
	w.Write([]byte(b))
}

func (FeedServer *feedServer) stylesheet(w http.ResponseWriter, r *http.Request) {
	if b, err := os.ReadFile("feed.xsl"); err == nil {
		w.Write([]byte(b))
	}
}

func (FeedServer *feedServer) readCmdParams() {
	flag.StringVar(&FeedServer.Port, "port", "8080", "Server port")
	flag.StringVar(&FeedServer.MediaFolder, "folder", "audio", "Name of a folder with media")
	flag.Parse()
}

func (Title *title) fromMediaFolder(mediaFolder string) []string {
	var titles []string

	eTitles, err := filepath.Glob(filepath.Join(mediaFolder, "/*"))
	check(err)

	for _, t := range eTitles {
		title := filepath.Base(t)
		titles = append(titles, title)
	}
	return titles
}

func (Title *title) readTitle(titlePath string) {

	var isChapter = regexp.MustCompile(`(?im)\.(mp3|m4a|m4b)$`)
	var isCover = regexp.MustCompile(`(?im)\.(jpg|jpeg|png)$`)

	var chapters []string

	_, err := os.ReadDir(titlePath)
	if err != nil {
		return
	}

	err = filepath.WalkDir(titlePath, func(path string, entry fs.DirEntry, err error) error {
		check(err)

		if !entry.IsDir() {
			if isChapter.MatchString(path) {
				chapters = append(chapters, filepath.ToSlash(path))
				log.Println("visited: ", path)
			} else if isCover.MatchString(path) {
				Title.CoverPath = path
				log.Println("cover found: ", Title.CoverPath)
			}
		}
		return nil
	})
	check(err)

	Title.Chapters = chapters
}

func main() {
	var FeedServer feedServer

	r := mux.NewRouter()
	r.HandleFunc("/index", FeedServer.index).Methods("GET")
	r.HandleFunc("/info", FeedServer.info).Methods("GET")
	r.HandleFunc("/feed.xsl", FeedServer.stylesheet).Methods("GET")
	r.HandleFunc("/title/{name}", FeedServer.displayTitle).Methods("GET")

	FeedServer.readCmdParams()

	http.Handle("/", r)
	http.Handle("/"+FeedServer.MediaFolder+"/", http.StripPrefix("/"+FeedServer.MediaFolder+"/", http.FileServer(http.Dir(FeedServer.MediaFolder))))

	log.Println("Listening :" + FeedServer.Port)
	err := http.ListenAndServe(":"+FeedServer.Port, nil)
	check(err)
}
