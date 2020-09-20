package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var PORT = os.Args[1]
var TOC_DIR string = "./toc"
var TOC_TEMPLATE string = "<h1>?</h1><div>?</div>"

func GenerateFromTemplate(header string, list string) string {
	r1 := strings.Replace(TOC_TEMPLATE, "?", header, 1)
	r2 := strings.Replace(r1, "?", list, 1)

	return r2
}

func ParseToc(title string) (*Directory, error) {
	f, err := os.Open(TOC_DIR)
	if err != nil {
		return nil, errors.New("unable to read toc file")
	}

	finf, err := f.Stat()
	if err != nil {
		return nil, errors.New("unable to read file size")
	}

	fsiz := finf.Size()
	toc := make([]byte, fsiz)
	f.Read(toc)

	itms := strings.SplitN(string(toc), "\n", -1)

	var itmslc []Item
	for _, v := range itms {
		itm := strings.SplitN(v, "::", -1)
		if len(itm) != 4 {
			fmt.Printf("item does not conform to struct, skipping: '%v'\n", v)
			break
		}

		islocal, _ := strconv.ParseBool(itm[2])
		hasssl, _ := strconv.ParseBool(itm[3])

		itmslc = append(itmslc, Item{
			Title:   itm[0],
			URL:     itm[1],
			isLocal: islocal,
			hasSSL:  hasssl,
		})
	}

	return &Directory{
		Title: title,
		Items: itmslc,
	}, nil
}

type Directory struct {
	Title string
	Items []Item
}

func (d *Directory) toHTMLList() string {
	var str string

	for _, v := range d.Items {
		str = str + v.toHTMLListItem()
	}

	return str
}

func (d *Directory) addItem(i Item) {
	d.Items = append(d.Items, i)
}

type Item struct {
	Title   string
	URL     string
	isLocal bool
	hasSSL  bool
}

func (i *Item) toHTMLListItem() string {
	return fmt.Sprintf("<li><a href='%s'>%s</a></li>", i.URL, i.Title)
}

func logRequest(r *http.Request) {
	log, err := os.OpenFile("./access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 770)
	defer log.Close()
	if err != nil {
		fmt.Printf("there was an error opening the log file: %s\n", err)
		return
	}

	nlog := fmt.Sprintf("[%s] [request] %+v\n", time.Now().UTC().String(), r)
	log.Write([]byte(nlog))
	fmt.Print(nlog)
}

func main() {
	dir, err := ParseToc("Dash")
	if err != nil {
		log.Fatal("unable to parse toc, exiting...")
	}

	http.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		res := GenerateFromTemplate(dir.Title, dir.toHTMLList())
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(res))
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil))
}
