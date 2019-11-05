package words

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/glibsm/words/internal/highlight"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

// Serve ...
func Serve(title string, opts ...Option) error {
	// initialize with defaults, then run through options to allow for overrides.
	s := &server{
		dir:  "posts",
		out:  "_build",
		port: 9000,
	}

	for _, opt := range opts {
		opt(s)
	}
	if len(opts) > 0 {
		log.Println("Applied", len(opts), "server options")
	}

	// drop and re-create build dir
	os.RemoveAll(s.out)
	if err := os.MkdirAll(s.out, os.ModePerm); err != nil {
		return fmt.Errorf("failed ot create build dir: %v", err)
	}
	log.Println("Created", s.out)

	files, err := findMDFiles(s.dir)
	if err != nil {
		return fmt.Errorf("failed to locate blog posts")
	}

	hr := highlight.New()

	var posts []*post
	for _, f := range files {
		p, err := read(f)
		if err != nil {
			return fmt.Errorf("failed to parse %v: %v", f, err)
		}

		p.html = blackfriday.Run(p.content, blackfriday.WithRenderer(hr))
		dest := filepath.Join(s.out, p.Slug)

		rendered, err := renderPost(p)
		if err != nil {
			return fmt.Errorf("failed to render content: %v", err)
		}

		ioutil.WriteFile(dest, rendered, os.ModePerm)
		log.Println("Wrote out", dest)

		posts = append(posts, p)
	}

	// sort posts in reverse chronological orders, as one would expect so the
	// newest ones float to the top.
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	indexContent, err := ioutil.ReadFile("index.md")
	if err != nil {
		log.Println("No (or malformed) index.md found, generating only links to posts")
	}

	indexRender, err := renderIndex(index{
		title:   title,
		content: indexContent,
		html:    blackfriday.Run(indexContent, blackfriday.WithRenderer(hr)),
	}, posts)
	if err != nil {
		return fmt.Errorf("failed to render index content: %v", err)
	}

	dest := "_build/index.html"
	ioutil.WriteFile(dest, indexRender, os.ModePerm)
	log.Println("Wrote out", dest)

	// move static content (if it exists) over to the build folder
	if _, err := os.Stat("static"); err == nil {
		// to copy one directory into another in pure Go is kind of wonky, so for now
		// to save some time just syscall out. Sorry, Windows...
		cmd := exec.Command("cp", "-r", "static", "_build/")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to move the static folder over: %v", err)
		}
		log.Println("Copied over the static folder")
	}

	fs := http.FileServer(http.Dir(s.out))
	http.Handle("/", fs)

	listen := fmt.Sprintf(":%d", s.port)
	log.Println("Listening on", listen)
	return http.ListenAndServe(listen, nil)
}

type post struct {
	*metadata

	content   []byte // original MD content
	html      []byte // processed through blackfriday
	HumanDate string
}

type index struct {
	title   string
	content []byte //index.md content
	html    []byte // processed through blackfriday
}

// read the blog post as *metadata and content
func read(path string) (*post, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	mdStart := bytes.Index(b, sep)
	if mdStart != 0 {
		// blog post doesn't stat with `---` thus missing the metadata section.
		// It's require for things like the slug. Plzor providezor.
		return nil, fmt.Errorf("%v doesn't start with ---", path)
	}

	// nasty nasty byteses...
	mdEnd := bytes.Index(b[3:], sep)
	mdBytes := b[3 : mdEnd+3]

	// parse out md as yaml and fuss if not
	var md *metadata
	if err := yaml.Unmarshal(mdBytes, &md); err != nil {
		return nil, err
	}

	if md.Date.IsZero() {
		return nil, errors.New("`date:` metadata is required")
	}

	if md.Slug == "" {
		return nil, errors.New("`slug:` metadata is required")
	}

	if md.Title == "" {
		return nil, errors.New("`title:` metadata is required")
	}

	return &post{
		metadata:  md,
		content:   b[mdEnd+6:],
		HumanDate: md.Date.Format("2006-01-02"),
	}, nil
}

type server struct {
	dir   string
	out   string
	port  int
	blurb string // blurb for the index page before the list of posts
}

type metadata struct {
	Slug  string    `yaml:"slug"`
	Title string    `yaml:"title"`
	Date  time.Time `yaml:"date"`
}

func findMDFiles(dir string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find any posts: %v", err)
	}
	return files, nil
}
