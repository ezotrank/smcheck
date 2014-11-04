package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	concurrency = flag.Int("c", 5, "Concurency level")
)

type SitemapIndex struct {
	Indexes []*Index `xml:"sitemap"`
}

type Index struct {
	Loc string `xml:"loc"`
}

type Pages struct {
	XMLName    xml.Name `xml:"urlset"`
	XmlNS      string   `xml:"xmlns,attr"`
	XmlImageNS string   `xml:"xmlns:image,attr"`
	XmlNewsNS  string   `xml:"xmlns:news,attr"`
	Pages      []*Page  `xml:"url"`
}

type Page struct {
	XMLName  xml.Name `xml:"url"`
	Loc      string   `xml:"loc"`
	Name     string   `xml:"news:news>news:publication>news:name"`
	Language string   `xml:"news:news>news:publication>news:language"`
	Title    string   `xml:"news:news>news:title"`
	Keywords string   `xml:"news:news>news:keywords"`
	Image    string   `xml:"image:image>image:loc"`
}

func unzip(body []byte) []byte {
	if hex.EncodeToString(body[:2]) == "1f8b" {
		var a bytes.Buffer
		buf := bytes.NewBuffer(body)
		r, _ := gzip.NewReader(buf)
		a.ReadFrom(r)
		return a.Bytes()
	}
	return body
}

func getUrl() <-chan string {
	out := make(chan string)

	go func() {
		resp, err := http.Get(flag.Args()[0])
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		index := &SitemapIndex{}
		pages := &Pages{}
		err = xml.Unmarshal(unzip(body), index)
		if err != nil {
			err = xml.Unmarshal(unzip(body), pages)
			if err != nil {
				panic(err)
			}
			for _, page := range pages.Pages {
				out <- page.Loc
			}
		} else {
			for _, index := range index.Indexes {
				fmt.Println("check index ", index.Loc)
				respi, erri := http.Get(index.Loc)
				if erri != nil {
					panic(erri)
				}
				defer respi.Body.Close()
				body, erri := ioutil.ReadAll(respi.Body)
				err = xml.Unmarshal(unzip(body), pages)
				if err != nil {
					panic(err)
				}
				for _, page := range pages.Pages {
					out <- page.Loc
				}
			}
		}
		close(out)
	}()
	return out
}

func main() {
	flag.Parse()

	sem := make(chan bool, *concurrency)
	for url := range getUrl() {
		sem <- true
		go func(url string) {
			tNow := time.Now()
			var resp *http.Response
			var err error
			defer func() {
				<-sem
				fmt.Println(resp.StatusCode, time.Now().Sub(tNow), url, err)
			}()
			resp, err = http.Get(url)
		}(url)
	}
}
