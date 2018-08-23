package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {

	urlPtr := flag.String("url", "", "url of the novel")
	flag.Parse()

	if *urlPtr == "" {
		fmt.Printf("Please specify a url of the book to download using the '--url' option")
		return
	}

	downloadBook(*urlPtr)
}

func downloadBook(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Could not make the get call. Error : %v\n", err)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("Could not parse the html to get doc. Error : %v\n", err)
	}
	defer resp.Body.Close()

	links := parseForPageLinks(doc)
	fmt.Printf("No.of pages is : %v\n", len(links))

	// use a tmp directory to hold the individual pages
	tmpdir := "temp_" + randomString(6)
	os.Mkdir(tmpdir, os.ModeDir)
	pageFileNames := make([]string, 0, len(links))

	wg := &sync.WaitGroup{}

	for index, link := range links {
		// create a temp file
		f, err := ioutil.TempFile(tmpdir, "page_")
		if err != nil {
			fmt.Printf("Could not create temp file for page %v. Error : %v\n", index, err)
			return
		}
		pageFileNames = append(pageFileNames, f.Name())
		wg.Add(1)
		go func(url string, n int, w io.WriteCloser) {
			getPage(url, n, w)
			defer w.Close()
			wg.Done()
		}(link, index, f)
	}

	wg.Wait()

	// now merge all the files to one
	book, err := os.Create("book.txt")
	if err != nil {
		fmt.Printf("Could not create the 'book.txt' file. Error : %v\n", err)
		return
	}
	for _, page := range pageFileNames {
		// wd, _ := os.Getwd()
		p, err := os.Open(page)
		if err != nil {
			fmt.Printf("Could not open the page file. Error: %v\n", err)
			return
		}
		io.Copy(book, p)
		p.Close()
		err = os.Remove(page)
		if err != nil {
			fmt.Printf("Could not delete file %v. Error : %v\n", page, err)
		}
	}

	book.Close()
	err = os.Remove(tmpdir)
	if err != nil {
		fmt.Printf("Could not remove the tmp directory. Error : %v\n", err)
	}

}

func parseForPageLinks(doc *html.Node) []string {

	links := make([]string, 0)
	var f func(*html.Node)
	f = func(n *html.Node) {
		// look for the correct "select" node
		if n.Type == html.ElementNode && n.Data == "select" {
			for _, a := range n.Attr {
				if a.Key == "id" && a.Val == "catid" {

					// now loop throught of it child "option" nodes and collect the links
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						if c.Type == html.ElementNode && c.Data == "option" {
							for _, aa := range c.Attr {
								if aa.Key == "value" && strings.Index(aa.Val, "http://") != -1 {
									links = append(links, aa.Val)
								}
							}
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return links[0 : len(links)/2]
}

func getPage(url string, n int, w io.Writer) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Could not get page %v. Error : %v\n", n, err)
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Printf("Could not parse the html of page %v to get doc. Error : %v\n", n, err)
	}
	defer resp.Body.Close()

	writePagetoFile(doc, w)
}

func writePagetoFile(doc *html.Node, w io.Writer) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "class" {
					if strings.Index(a.Val, "tel_content") != -1 {

						// now loop throught all the child "TextNodes" and write the value to the writer
						readTextRecursivelyandWritetoFile(n, w)

					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
}

func readTextRecursivelyandWritetoFile(n *html.Node, w io.Writer) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			_, err := w.Write([]byte(c.Data))
			if err != nil {
				fmt.Printf("Failed while writing to file. Error : %v\n", err)
				fmt.Printf("Aborting...")
				break
			}
		}
		readTextRecursivelyandWritetoFile(c, w)
	}
}

func randomString(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
