package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/go-pkgz/lgr"
	"github.com/gorilla/mux"
	"github.com/kennygrant/sanitize"
)

type feedServer struct {
	HostName    string
	MediaFolder string
	Port        string
	*lgr.Logger
}

func New(l *lgr.Logger) *feedServer {
	return &feedServer{
		Logger: l,
	}
}

type title struct {
	Host      string
	Name      string
	Path      string
	CoverPath string
	Chapters  []string
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
	titleName := sanitize.BaseName(params["name"])

	s.Logf("INFO Reading title '%s'", titleName)

	сhapters, coverPath, err := s.readTitle(filepath.Join(s.MediaFolder, titleName))
	if err != nil {
		s.Logf("ERROR reading title: %s", err.Error())
		http.Error(w, "Error reading title", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "text/xml; charset=utf-8")
	xmlTemplate.Execute(w, title{
		r.Host,
		titleName,
		filepath.ToSlash(filepath.Join("title", titleName)),
		coverPath,
		сhapters,
	})
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

func (*feedServer) fromMediaFolder(mediaFolder string) ([]string, error) {
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

func (t *feedServer) readTitle(titlePath string) (chapters []string, coverPath string, err error) {

	var isChapter = regexp.MustCompile(`(?im)\.(mp3|m4a|m4b)$`)
	var isCover = regexp.MustCompile(`(?im)\.(jpg|jpeg|png)$`)

	_, err = os.ReadDir(titlePath)
	if err != nil {
		return nil, "", err
	}

	err = filepath.WalkDir(titlePath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			if isChapter.MatchString(path) {
				chapters = append(chapters, filepath.ToSlash(path))
			} else if isCover.MatchString(path) {
				coverPath = path
			}
		}
		return nil
	})

	if err != nil {
		return nil, "", err
	}
	return
}

func main() {

	l := lgr.New(lgr.Format(lgr.FullDebug))
	FeedServer := New(l)

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
