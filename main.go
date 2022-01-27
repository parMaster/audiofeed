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

	"github.com/go-pkgz/lgr"
	"github.com/gorilla/mux"
)

type feedServer struct {
	HostName    string
	MediaFolder string
	Port        string
	Logger
	title
}

func NewServer(l Logger) *feedServer {
	s := &feedServer{}
	s.Logger = l
	return s
}

type title struct {
	Name      string
	Path      string
	CoverPath string
	Chapters  []string
}

type Logger interface {
	Logf(format string, args ...interface{})
}

func (s *feedServer) index(w http.ResponseWriter, r *http.Request) {
	titlesTemplate := template.New("Title with chapters")
	titlesTemplate.Parse(titlesTemplateBody)

	titles, err := s.fromMediaFolder(s.MediaFolder)
	if err != nil {
		s.Logf("WARNING Reading folder error: %s", err.Error())
	}
	titlesTemplate.Execute(w, titles)
}

func (s *feedServer) displayTitle(w http.ResponseWriter, r *http.Request) {
	xmlTemplate := template.New("Title with chapters")
	xmlTemplate.Parse(xmlTemplateBody)

	params := mux.Vars(r)
	s.HostName = r.Host
	s.Name = params["name"] // filter somehow?
	s.Path = filepath.ToSlash(filepath.Join("title", s.Name))

	s.Logf("INFO Reading title '%s'", s.Name)
	s.readTitle(filepath.Join(s.MediaFolder, s.Name))

	w.Header().Add("Content-Type", "text/xml; charset=utf-8")
	xmlTemplate.Execute(w, s)
}

func (*feedServer) info(w http.ResponseWriter, r *http.Request) {
	b := fmt.Sprintf("%v", r)
	w.Write([]byte(b))
}

func (*feedServer) stylesheet(w http.ResponseWriter, r *http.Request) {
	if b, err := os.ReadFile("feed.xsl"); err == nil {
		w.Write([]byte(b))
	}
}

func (s *feedServer) readCmdParams() {
	flag.StringVar(&s.Port, "port", "8080", "Server port")
	flag.StringVar(&s.MediaFolder, "folder", "audio", "Name of a folder with media")
	flag.Parse()
}

func (*title) fromMediaFolder(mediaFolder string) ([]string, error) {
	var titles []string

	eTitles, err := filepath.Glob(filepath.Join(mediaFolder, "/*"))
	if err != nil {
		return nil, err
	}

	for _, t := range eTitles {
		title := filepath.Base(t)
		titles = append(titles, title)
	}
	return titles, nil
}

func (t *title) readTitle(titlePath string) error {

	var isChapter = regexp.MustCompile(`(?im)\.(mp3|m4a|m4b)$`)
	var isCover = regexp.MustCompile(`(?im)\.(jpg|jpeg|png)$`)

	var chapters []string

	_, err := os.ReadDir(titlePath)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(titlePath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			if isChapter.MatchString(path) {
				chapters = append(chapters, filepath.ToSlash(path))
				log.Println("visited: ", path)
			} else if isCover.MatchString(path) {
				t.CoverPath = path
				log.Println("cover found: ", t.CoverPath)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	t.Chapters = chapters
	return nil
}

func main() {
	l := lgr.New(lgr.Format(lgr.FullDebug))

	FeedServer := NewServer(l)

	r := mux.NewRouter()
	r.HandleFunc("/index", FeedServer.index).Methods("GET")
	r.HandleFunc("/info", FeedServer.info).Methods("GET")
	r.HandleFunc("/feed.xsl", FeedServer.stylesheet).Methods("GET")
	r.HandleFunc("/title/{name}", FeedServer.displayTitle).Methods("GET")

	FeedServer.readCmdParams()

	http.Handle("/", r)
	http.Handle("/"+FeedServer.MediaFolder+"/", http.StripPrefix("/"+FeedServer.MediaFolder+"/", http.FileServer(http.Dir(FeedServer.MediaFolder))))

	l.Logf("Listening: %s", FeedServer.Port)
	err := http.ListenAndServe(":"+FeedServer.Port, nil)
	l.Logf("ERROR ListenAndServe: %s", err.Error())
}
