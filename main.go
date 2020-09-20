package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func GenerateFromTemplate(header string, list string) string {
	t := "<style> * { font-family: sans-serif; } body { padding: 0; margin: 0; } h1 { font-size: 60px; } a { text-decoration: none; color: black; font-size: 25px; } a:hover { color: #005288; } #container { height: 100vh; width: 100vw; display: flex; justify-content: center; align-items: center; }</style><div id='container'><div><h1>?</h1></div><div><ul>?</ul></div><div>"

	r1 := strings.Replace(t, "?", header, 1)
	r2 := strings.Replace(r1, "?", list, 1)

	return r2
}

func ParseToc(title string, path string) (*Directory, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("unable to read toc file")
	}

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(f)

	var itms []Item
	e1 := json.Unmarshal(buf.Bytes(), &itms)
	if e1 != nil {
		return nil, e1
	}

	return &Directory{
		Title:    title,
		FilePath: path,
		Items:    itms,
	}, nil
}

type Directory struct {
	Title    string
	FilePath string
	Items    []Item
}

func (d *Directory) toHTMLList() string {
	var str string

	for _, v := range d.Items {
		str = str + v.toHTMLListItem()
	}

	return str
}

func (d *Directory) addItem(i Item) error {
	f, err := os.OpenFile(d.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 770)
	defer f.Close()
	if err != nil {
		return err
	}

	itm := fmt.Sprintf("%s::%s::%s::%s\n", i.Title, i.URL)
	f.Write([]byte(itm))
	d.Items = append(d.Items, i)
	return nil
}

type Item struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func (i *Item) toHTMLListItem() string {
	return fmt.Sprintf("<li><a href='%s'>%s</a></li>", i.URL, i.Title)
}

func logRequest(r *http.Request) {
	log, err := os.OpenFile("./access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 770)
	if err != nil {
		fmt.Printf("there was an issue opening the log file: %s\n", err)
		return
	}

	nlog := fmt.Sprintf("[%s] [request] %+v\n", time.Now().UTC().String(), r)
	log.Write([]byte(nlog))
	fmt.Print(nlog)
}

func main() {
	dir, err := ParseToc("Dash", "./toc.json")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		res := GenerateFromTemplate(dir.Title, dir.toHTMLList())
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(res))
	})

	http.HandleFunc("/dash/item/new", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		buf := bytes.NewBuffer([]byte{})
		buf.ReadFrom(r.Body)

		var itm Item
		json.Unmarshal(buf.Bytes(), &itm)
		dir.addItem(itm)

		js := json.RawMessage(`{ "status": "ok", "message": "successfully added new item" }`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil))
}
