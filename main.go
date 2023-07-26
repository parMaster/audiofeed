package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-pkgz/lgr"
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
	AccessCode  string
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

	// if access code is set, check it
	if s.AccessCode != "" {
		code := chi.URLParam(r, "code")
		if code != s.AccessCode {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

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

	titleName := chi.URLParam(r, "title")
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"Host":            r.Host,
		"RequestURI":      r.RequestURI,
		"RemoteAddr":      r.RemoteAddr,
		"date":            fmt.Sprintf("%s", time.Now()),
		"UserAgent":       r.UserAgent(),
		"Accept-Encoding": r.Header.Get("Accept-Encoding"),
		"Accept-Language": r.Header.Get("Accept-Language"),
		"Connection":      r.Header.Get("Connection"),
		"Accept":          r.Header.Get("Accept"),
		"Accept-Charset":  r.Header.Get("Accept-Charset"),
	})
}

func (*feedServer) stylesheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xsl; charset=utf-8")
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
		// skip .gitignore and other files
		if strings.HasPrefix(title, ".") {
			continue
		}
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
	r := chi.NewRouter()

	r.Route("/index", func(r chi.Router) {
		r.Get("/{code}", s.index)
		r.Get("/", s.index)
	})

	r.Get("/info", s.info)
	r.Get("/feed.xsl", s.stylesheet)
	r.Get("/title/{title}", s.displayTitle)

	fs := http.FileServer(http.Dir("./" + s.MediaFolder))
	r.Handle("/"+s.MediaFolder+"/*", http.StripPrefix("/"+s.MediaFolder, filesOnly(fs)))

	httpServer := &http.Server{
		Addr:              ":" + s.Port,
		Handler:           r,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Second,
		ErrorLog:          lgr.ToStdLogger(lgr.Default(), "WARN"),
	}
	log.Printf("[INFO] Listening: %s", s.Port)

	titles, err := s.fromMediaFolder(s.MediaFolder)
	if err != nil {
		log.Printf("[WARN] Reading folder error: %s", err.Error())
	}

	log.Printf("[INFO] Found: %d titled in '%s' folder", len(titles), s.MediaFolder)

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

func filesOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		log.Printf("[INFO] %s (%s)", r.URL.Path, r.Header.Get("X-Real-Ip"))
		next.ServeHTTP(w, r)
	})
}

func main() {
	feedServer := NewFeedServer()

	var dbg = flag.Bool("dbg", false, "Debug mode")
	flag.StringVar(&feedServer.Port, "port", "8080", "Server port")
	flag.StringVar(&feedServer.MediaFolder, "folder", "audio", "Name of a folder with media, ./audio by default")
	flag.StringVar(&feedServer.AccessCode, "code", "", "(optional) Access Code, if set, will be required for access to /index/{code} to list titles")
	flag.Parse()

	logOpts := []lgr.Option{lgr.LevelBraces}
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
