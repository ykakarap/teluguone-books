# Telugu books downloader for TeluguOne Grandhalayam

A book downloader to download books from [TeluguOne Grandhalayam](http://www.teluguone.com/grandalayam/).
Downloads the contents of the novel on the webpage and create a `book.txt` file out of it.

How to use : 
```
$ go run main.go --url="<place url here>" 
```

## TO-DO :
- Convert the `book.txt` file to a PDF
- Improve quality of code
  - Improve error handling
  - Improve logging
  - Use some kind of workerpool to control no.of goroutines fired
- Upload generated book to play books to read on mobile.
