package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"archive/zip"
	"path/filepath"

	"net/http"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	log "github.com/llimllib/loglevel"
)

type Link struct {
	url   string
	text  string
	depth int
}

type HttpError struct {
	original string
}

func (self Link) String() string {
	spacer := strings.Repeat("\t", self.depth)
	return fmt.Sprintf("%s%s (%d) - %s", spacer, self.text, self.depth, self.url)

	//return fmt.Sprintf("%s", self.url)
}

func (self Link) Valid() bool {
	if self.depth >= MaxDepth {
		return false
	}

	if len(self.text) == 0 {
		return false
	}
	if len(self.url) == 0 || strings.Contains(strings.ToLower(self.url), "javascript") {
		return false
	}

	if len(self.url) != 0 && strings.Contains(strings.ToLower(self.url), "sinit") && strings.Contains(strings.ToLower(self.url), ".zip") {
		return true
	}
	return false
}

func (self HttpError) Error() string {
	return self.original
}

var MaxDepth = 1

func linkReader(resp *http.Response, depth int) []Link {
	page := html.NewTokenizer(resp.Body)
	links := []Link{}

	var start *html.Token
	var text string

	for {
		_ = page.Next()
		token := page.Token()
		if token.Type == html.ErrorToken {
			break
		}

		if start != nil && token.Type == html.TextToken {
			text = fmt.Sprintf("%s%s", text, token.Data)
		}

		if token.DataAtom == atom.A {
			switch token.Type {
			case html.StartTagToken:
				if len(token.Attr) > 0 {
					start = &token
				}
			case html.EndTagToken:
				if start == nil {
					log.Warnf("Link End found without Start: %s", text)
					continue
				}
				link := newLink(*start, text, depth)
				if link.Valid() {
					links = append(links, link)
					log.Debugf("Link Found %v", link)
				}

				start = nil
				text = ""
			}
		}
	}

	log.Debug(links)
	return links
}

func newLink(tag html.Token, text string, depth int) Link {
	link := Link{text: strings.TrimSpace(text), depth: depth}

	for i := range tag.Attr {
		if tag.Attr[i].Key == "data-id-url" {
			link.url = strings.TrimSpace(tag.Attr[i].Val)
		}
	}
	return link
}

func recurUrlLoader(url string, depth int) []Link {
	page, err := urlLoader(url)
	if err != nil {
		log.Error(err)
		return nil
	}
	links := linkReader(page, depth)

	return links
}

func urlLoader(url string) (resp *http.Response, err error) {
	log.Debugf("Downloading %s", url)
	resp, err = http.Get(url)
	if err != nil {
		log.Debugf("Error: %s", err)
		return
	}

	if resp.StatusCode > 299 {
		err = HttpError{fmt.Sprintf("Error (%d): %s", resp.StatusCode, url)}
		log.Debug(err)
		return
	}
	return

}

func downloadAndUnzipFiles(link Link, of string, clean bool) {
	dir, err := os.Getwd()

	if err != nil {
		log.Debugf("Error: %s", err)
		return
	}

	if of == "./" {
		of = dir
	} else if of[0:2] == "./" {
		of = dir + of[1:len(of)]
	}

	var zipFolder string
	if clean {
		zipFolder = "/tmp/"
	} else {
		zipFolder = of + "/zip/"
	}

	os.MkdirAll(filepath.Dir(zipFolder), os.ModePerm)

	filename := getFilename(link.url)
	err = downloadFile(zipFolder+filename, link.url)
	//fmt.Println(filename)
	if err != nil {
		log.Debugf("Error: %s", err)
		return
	}

	err = unzip(zipFolder+filename, of)
	if err != nil {
		log.Debug(err)
	}

	//fmt.Println("Unzipped:\n" + strings.Join(files, "\n"))

	if clean {
		err = os.Remove("/tmp/" + filename)
		if err != nil {
			log.Debug(err)
		}
	}

}

func downloadFile(filepath string, url string) error {

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(src string, dest string) error {

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.ToLower(f.Name[strings.LastIndex(f.Name, "."):len(f.Name)]) != ".bin" {
			continue
		}
		// Store filename/path for returning and using later on
		index := strings.LastIndex(f.Name, "/")

		fpath := filepath.Join(dest, f.Name[index+1:len(f.Name)])

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
func getFilename(url string) string {
	index := strings.LastIndex(url, "/")
	return url[index+1 : (len(url))]
}

func getFoldername(dir string, filename string) string {
	index := strings.LastIndex(filename, ".")
	return dir + "/" + filename[0:index]
}

func main() {
	ptrSite := flag.String("url", "https://software.intel.com/content/www/us/en/develop/articles/intel-trusted-execution-technology.html", "url to be searched for Links")
	ptrClean := flag.Bool("noClean", false, "keep zip files after unzipping")
	ptrOutputFolder := flag.String("of", "./", "output folder")

	flag.Parse()

	log.SetPriorityString("info")
	log.SetPrefix("crawler")

	log.Debug(os.Args)

	of := *ptrOutputFolder

	links := recurUrlLoader(*ptrSite, 0)

	for _, link := range links {
		downloadAndUnzipFiles(link, of, !*ptrClean)
	}
}
