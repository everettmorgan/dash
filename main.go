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
)

func generateFromTemplate(header string, list string) string {
	t := "<style> * { font-family: sans-serif; } body { padding: 0; margin: 0; } h1 { font-size: 60px; } a { text-decoration: none; color: black; font-size: 25px; } a:hover { color: #005288; } #container { height: 100vh; width: 100vw; display: flex; justify-content: center; align-items: center; }</style><div id='container'><div><h1>?</h1></div><div><ul>?</ul></div><div>"

	r1 := strings.Replace(t, "?", header, 1)
	r2 := strings.Replace(r1, "?", list, 1)

	return r2
}

func parseToc(title string, path string) (*Directory, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("unable to read toc file")
	}

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(f)

	var itms []Item
	e1 := json.Unmarshal(buf.Bytes(), &itms)
	if e1 != nil {
		log.Fatal("unable to parse toc, exiting...", buf.String())
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

func (d *Directory) addItem(i Item) {
	f, e1 := os.OpenFile(d.FilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 770)
	defer f.Close()
	if e1 != nil {
		fmt.Println("unable to marshal toc json object, skipping...", e1)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	buf.ReadFrom(f)

	var itms []Item
	e2 := json.Unmarshal(buf.Bytes(), &itms)
	if e2 != nil {
		log.Fatal("unable to parse toc, exiting...", buf.String())
	}

	itms = append(itms, i)
	js, e3 := json.Marshal(itms)
	if e3 != nil {
		fmt.Println("unable to marshal new item slice, skipping...", e3)
		return
	}

	_, e4 := f.Write(js)
	if e4 != nil {
		fmt.Println("unable to write new item to file, skipping...", e4)
		return
	}
	d.Items = append(d.Items, i)
}

type Item struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func (i *Item) toHTMLListItem() string {
	return fmt.Sprintf("<li><a href='%s'>%s</a></li>", i.URL, i.Title)
}

func main() {
	dir, err := parseToc("Dash", "./toc.json")
	if err != nil {
		log.Fatal("unable to parse toc file, exiting...", err)
	}

	http.HandleFunc("/dash", func(w http.ResponseWriter, r *http.Request) {
		res := generateFromTemplate(dir.Title, dir.toHTMLList())
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(res))
	})

	http.HandleFunc("/dash/item/new", func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer([]byte{})
		buf.ReadFrom(r.Body)

		var itm Item
		json.Unmarshal(buf.Bytes(), &itm)

		w.Header().Set("Content-Type", "application/json")

		if itm.Title == "" && itm.URL == "" {
			js := json.RawMessage(`{ "status": "failed", "message": "missing item title and url" }`)
			w.Write(js)
			return
		}

		dir.addItem(itm)
		js := json.RawMessage(`{ "status": "ok", "message": "successfully added new item" }`)
		w.Write(js)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil))
}
