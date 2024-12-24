package main

import (
	_ "bytes"
	"flag"
	_ "fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	_ "path"
	"path/filepath"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var (
	startURL       string
	outputDir      string
	recursionDepth int
	visited        = make(map[string]bool)
	mutex          sync.Mutex
	sem            = make(chan struct{}, 10) // Semaphore for concurrency
)

func init() {
	flag.StringVar(&startURL, "url", "", "Starting URL of the website")
	flag.StringVar(&outputDir, "output", "output", "Output directory")
	flag.IntVar(&recursionDepth, "depth", 0, "Recursion depth (0 for unlimited)")
	flag.Parse()
}

func main() {
	if startURL == "" {
		log.Fatal("Starting URL is required.")
	}

	u, err := url.Parse(startURL)
	if err != nil {
		log.Fatalf("Invalid URL: %s", startURL)
	}

	// Ensure output directory exists
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Start downloading from the starting URL
	download(u, outputDir, 0)
}

func download(u *url.URL, baseDir string, depth int) {
	if recursionDepth > 0 && depth > recursionDepth {
		return
	}

	// Check if URL has been visited
	mutex.Lock()
	if visited[u.String()] {
		mutex.Unlock()
		return
	}
	visited[u.String()] = true
	mutex.Unlock()

	// Fetch the content
	resp, err := http.Get(u.String())
	if err != nil {
		log.Printf("Error fetching %s: %v", u, err)
		return
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to download %s: status code %d", u, resp.StatusCode)
		return
	}

	// Determine the local file path
	relPath := u.Path
	if relPath == "" || relPath == "/" {
		relPath = "index.html"
	} else if !filepath.IsAbs(relPath) {
		relPath = filepath.FromSlash(relPath)
	}

	filePath := filepath.Join(baseDir, relPath)
	fileDir := filepath.Dir(filePath)

	// Create directory if it doesn't exist
	err = os.MkdirAll(fileDir, 0755)
	if err != nil {
		log.Printf("Failed to create directory %s: %v", fileDir, err)
		return
	}

	// Save the content to the file
	f, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file %s: %v", filePath, err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Printf("Failed to write to file %s: %v", filePath, err)
		return
	}

	log.Printf("Downloaded %s to %s", u, filePath)

	// If it's an HTML page, parse and find links
	if resp.Header.Get("Content-Type") == "text/html; charset=utf-8" {
		go func() {
			defer func() { <-sem }()
			parseHTML(u, baseDir, depth, filePath)
		}()
		sem <- struct{}{} // Acquire semaphore
	}
}

func parseHTML(u *url.URL, baseDir string, depth int, filePath string) {
	doc, err := goquery.NewDocument(filePath)
	if err != nil {
		log.Printf("Error parsing HTML %s: %v", filePath, err)
		return
	}

	// Find all links in the page
	doc.Find("a, img, link, script").Each(func(i int, s *goquery.Selection) {
		attr := s.AttrOr("data-src", s.AttrOr("src", s.AttrOr("href", "")))

		link, exists := s.Attr(attr)
		if !exists {
			return
		}

		// Resolve relative URLs
		absURL, err := u.Parse(link)
		if err != nil {
			log.Printf("Error parsing link %s in %s: %v", link, filePath, err)
			return
		}

		// Check if the link is within the same domain
		if absURL.Host != u.Host {
			return
		}

		// Recursively download the linked resource
		download(absURL, baseDir, depth+1)
	})
}
