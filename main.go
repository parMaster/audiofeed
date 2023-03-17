package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"text/template"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/gorilla/mux"
	"github.com/kennygrant/sanitize"
)

//go:embed web/feed.xml
var feed_xml string

//go:embed web/feed.xsl
var feed_xsl string

//go:embed web/titles.html
var titles_html string

type feedServer struct {
	HostName    string
	MediaFolder string
	Port        string
}

func NewFeedServer() *feedServer {
	return &feedServer{}
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
	titlesTemplate.Parse(titles_html)

	titles, err := s.fromMediaFolder(s.MediaFolder)
	if err != nil {
		log.Printf("[WARN] Reading folder error: %s", err.Error())
	}
	titlesTemplate.Execute(w, titles)
}

func (s *feedServer) displayTitle(w http.ResponseWriter, r *http.Request) {
	xmlTemplate := template.New("Title with chapters")
	xmlTemplate.Parse(feed_xml)

	params := mux.Vars(r)
	titleName := sanitize.BaseName(params["name"])

	log.Printf("[INFO] Reading title '%s'", titleName)

	сhapters, coverPath, err := s.readTitle(filepath.Join(s.MediaFolder, titleName))
	if err != nil {
		log.Printf("[ERROR] reading title: %s", err.Error())
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
	w.Write([]byte(feed_xsl))
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

func (s *feedServer) Run(ctx context.Context) error {
	r := mux.NewRouter()
	r.HandleFunc("/index", s.index).Methods("GET")
	r.HandleFunc("/info", s.info).Methods("GET")
	r.HandleFunc("/feed.xsl", s.stylesheet).Methods("GET")
	r.HandleFunc("/title/{name}", s.displayTitle).Methods("GET")
	fs := http.StripPrefix("/"+s.MediaFolder+"/", http.FileServer(http.Dir("./"+s.MediaFolder+"/")))
	r.PathPrefix("/" + s.MediaFolder + "/").Handler(fs)

	httpServer := &http.Server{
		Addr:              ":" + s.Port,
		Handler:           r,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Second,
		ErrorLog:          lgr.ToStdLogger(lgr.Default(), "WARN"),
	}
	log.Printf("Listening: %s", s.Port)

	go func() {
		<-ctx.Done()
		if httpServer != nil {
			if err := httpServer.Close(); err != nil {
				log.Printf("[ERROR] failed to close http server, %v", err)
			}
		}
	}()

	return httpServer.ListenAndServe()
}

func main() {
	feedServer := NewFeedServer()

	var dbg = flag.Bool("dbg", false, "Debug mode")
	flag.StringVar(&feedServer.Port, "port", "8080", "Server port")
	flag.StringVar(&feedServer.MediaFolder, "folder", "audio", "Name of a folder with media")
	flag.Parse()

	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	if *dbg {
		logOpts = append(logOpts, lgr.Debug, lgr.CallerFile, lgr.CallerFunc)
	}
	lgr.SetupStdLogger(logOpts...)
	lgr.Setup(logOpts...)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic:\n%v", x)
			panic(x)
		}

		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	if err := feedServer.Run(ctx); err != nil && err.Error() != "http: Server closed" {
		log.Printf("[ERROR] %s", err)
	}
}
