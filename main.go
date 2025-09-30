package main

import (
	"context"
	_ "embed"
	"encoding/json"
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

	"github.com/go-pkgz/lgr"
	"github.com/go-pkgz/routegroup"
	"github.com/jessevdk/go-flags"
)

//go:embed web/feed.xml
var feed_xml string

//go:embed web/feed.xsl
var feed_xsl string

//go:embed web/titles.html
var titles_html string

type feedServer struct {
	Options
}

func NewFeedServer(opts Options) *feedServer {
	return &feedServer{Options: opts}
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
		code := r.PathValue("code")
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

	titleName := r.PathValue("title")
	log.Printf("[INFO] Reading title '%s'", titleName)

	сhapters, coverPath, err := s.readTitle(filepath.Join(s.MediaFolder, titleName))
	if err != nil {
		log.Printf("[ERROR] reading title: %s", err.Error())
		http.Error(w, "Error reading title", http.StatusInternalServerError)
		return
	}

	for i, c := range сhapters {
		// remove s.MediaFolder prefix from each chapter path
		сhapters[i], _ = strings.CutPrefix(c, filepath.ToSlash(s.MediaFolder)+"/")
		// add "audio/" prefix to each chapter path for http access
		сhapters[i] = "audio/" + сhapters[i]
		log.Printf("[DEBUG] file: %s", сhapters[i])
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
		"date":            time.Now().String(),
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
		// skip .gitignore and other dot-files
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
	s.MediaFolder, _ = filepath.Abs(s.MediaFolder)

	mux := http.NewServeMux()
	router := routegroup.New(mux)

	router.HandleFunc("/index", s.index)
	router.HandleFunc("/index/{code}", s.index)
	router.HandleFunc("/info", s.info)
	router.HandleFunc("/feed.xsl", s.stylesheet)
	router.HandleFunc("/title/{title}", s.displayTitle)

	router.With(filesOnly).HandleFiles("/audio", http.Dir(s.MediaFolder))

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Port),
		Handler:           router,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Second,
		ErrorLog:          lgr.ToStdLogger(lgr.Default(), "WARN"),
	}
	log.Printf("[INFO] Listening: %d", s.Port)

	titles, err := s.fromMediaFolder(s.MediaFolder)
	if err != nil {
		log.Printf("[WARN] Reading folder error: %s", err.Error())
	}

	log.Printf("[INFO] Found: %d titled in '%s' folder", len(titles), s.MediaFolder)

	go func() {
		<-ctx.Done()
		if httpServer != nil {
			if err := httpServer.Shutdown(ctx); err != nil {
				log.Printf("[ERROR] failed to close http server, %v", err)
			}
		}
	}()

	return httpServer.ListenAndServe()
}

// middleware to allow only files, no folders listing
func filesOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type Options struct {
	Dbg         bool   `long:"dbg" description:"Debug mode"`
	HostName    string `long:"host" default:"localhost" description:"Server host name"`
	Port        uint   `long:"port" default:"8080" description:"Server port"`
	MediaFolder string `long:"folder" default:"./audio" description:"Name of a folder with media, ./audio by default"`
	AccessCode  string `long:"code" description:"(optional) Access Code, if set, will be required for access to /index/{code} to list titles"`
}

func main() {
	var cfg Options

	p := flags.NewParser(&cfg, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		p.WriteHelp(os.Stderr)
		os.Exit(2)
	}

	feedServer := NewFeedServer(cfg)

	logOpts := []lgr.Option{
		lgr.LevelBraces,
		lgr.StackTraceOnError,
	}
	if feedServer.AccessCode != "" {
		logOpts = append(logOpts, lgr.Secret(feedServer.AccessCode))
	}
	if feedServer.Dbg {
		logOpts = append(logOpts, lgr.Debug, lgr.CallerFile, lgr.CallerFunc)
	}
	lgr.SetupStdLogger(logOpts...)

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
		log.Printf("[INFO] shutting down")
		cancel()
	}()

	if err := feedServer.Run(ctx); err != nil && err.Error() != "http: Server closed" {
		log.Printf("[ERROR] %s", err)
	}
}
