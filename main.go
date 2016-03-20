package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func yahooRanking(tch chan string) {

	doc, err := goquery.NewDocument("http://searchranking.yahoo.co.jp/burst_ranking/")
	if err != nil {
		log.Printf("yahoo err: %s", err)
		close(tch)
		return
	}

	doc.Find("ul.patC li a").Each(func(i int, s *goquery.Selection) {
		t := s.Text()
		tch <- t
	})

	close(tch)
}

func getURL(doc *goquery.Document, uch chan string) {
	f := doc.Find("h3.r a")
	f.Each(func(i int, s *goquery.Selection) {
		u, ok := s.Attr("href")
		if ok {
			uch <- u
		}
	})
}

func googleSearch(tch chan string, uch chan string) {
	var wg sync.WaitGroup
loop:
	for {
		select {
		case t, ok := <-tch:
			if !ok {
				break loop
			}

			wg.Add(1)
			go func(t string) {
				defer wg.Done()

				q := url.QueryEscape(t)
				u := fmt.Sprintf("https://www.google.co.jp/search?q=%s", q)
				doc, err := goquery.NewDocument(u)
				if err != nil {
					log.Printf("google err: %s, title %s\n", err, t)
					return
				}

				getURL(doc, uch)

				f := doc.Find("td a.fl")
				f.Each(func(i int, s *goquery.Selection) {
					u, ok := s.Attr("href")
					if ok {
						go func() {
							doc, err := goquery.NewDocument(fmt.Sprintf("https://www.google.co.jp%s", u))
							if err != nil {
								log.Printf("google err: %s, title %s\n", err, t)
								return
							}
							getURL(doc, uch)
						}()
					}
				})

			}(t)
		}
	}
	wg.Wait()
	close(uch)
}

func collectURL(uch chan string, lch chan []string) {
	list := make([]string, 0, 10000)
loop:
	for {
		select {
		case u, ok := <-uch:
			if !ok {
				break loop
			}
			list = append(list, u)
		}
	}

	lch <- list
}

func accessURL(u string) {
	target := fmt.Sprintf("https://google.co.jp%s", u)
	fmt.Println(base64.RawURLEncoding.EncodeToString([]byte(target)))
}

func main() {
	tch := make(chan string, 100)
	uch := make(chan string)
	lch := make(chan []string)

	go yahooRanking(tch)
	go googleSearch(tch, uch)
	go collectURL(uch, lch)

	list := <-lch
	/*for i, u := range list {
		//fmt.Printf("#%05d: %s\n", i, u)
	}*/

	for _, u := range list {
		accessURL(u)
	}
}
