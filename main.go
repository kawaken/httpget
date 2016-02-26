package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func yahooRanking(tch chan string) {
	log.Println("starg yahoo ranking")

	doc, err := goquery.NewDocument("http://searchranking.yahoo.co.jp/burst_ranking/")
	if err != nil {
		log.Printf("yahoo err: %s", err)
		close(tch)
		return
	}

	doc.Find("ul.patC li a").Each(func(i int, s *goquery.Selection) {
		t := s.Text()
		log.Printf("send title %s\n", t)
		tch <- t
	})

	log.Println("end yahoo ranking")
	close(tch)
}

func googleSearch(tch chan string, uch chan string) {
	log.Println("start google search")
	var wg sync.WaitGroup
loop:
	for {
		select {
		case t, ok := <-tch:
			if !ok {
				break loop
			}

			log.Printf("receive title %s\n", t)

			wg.Add(1)
			go func(t string) {
				log.Printf("start google query: %s\n", t)
				q := url.QueryEscape(t)
				u := fmt.Sprintf("https://www.google.co.jp/search?q=%s", q)
				doc, err := goquery.NewDocument(u)
				if err != nil {
					log.Printf("google err: %s, title %s\n", err, t)
					return
				}

				f := doc.Find("h3.r a")
				fmt.Println(f.Length())

				f.Each(func(i int, s *goquery.Selection) {
					u, ok := s.Attr("href")
					if ok {
						log.Printf("send url: %s\n", u)
						uch <- u
					}
				})

				wg.Done()

			}(t)
		}
	}
	wg.Wait()
	close(uch)
	log.Println("end google search")
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
			log.Printf("receive url: %s\n", u)
			list = append(list, u)
		}
	}

	log.Printf("send list: len %d\n", len(list))
	lch <- list
}

func main() {
	tch := make(chan string, 100)
	uch := make(chan string)
	lch := make(chan []string)

	go yahooRanking(tch)
	go googleSearch(tch, uch)
	go collectURL(uch, lch)

	log.Println("waiting collectURL")
	list := <-lch
	for i, u := range list {
		fmt.Printf("#%5d: %s", i, u)
	}
	log.Println("done")
}
