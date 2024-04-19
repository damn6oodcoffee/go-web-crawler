package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/damn6oodcoffee/go-web-crawler/link_parser"
)

type linkWithDepth struct {
	depth int
	url   string
}

type batchWithDepth struct {
	depth int
	urls  []string
}

func crawl(link *linkWithDepth) *batchWithDepth {
	fmt.Println(link.url, " <=== ", link.depth)
	links, err := link_parser.ExtractURLs(link.url)
	if err != nil {
		log.Print(err)
	}
	return &batchWithDepth{link.depth + 1, links}
}

func main() {

	writeToFile := flag.Bool("w", false, "запись в файл")
	concurRate := flag.Int("rate", 20, "число горутин для поиска")
	maxLinkDepth := flag.Int("depth", 3, "глубина поиска ссылок")
	flag.Parse()

	links := make(chan *batchWithDepth)
	unseenLinks := make(chan *linkWithDepth)

	var n int // Число обрабатываемых ссылок в данный момент
	n++       // +1 : начинаем с ссылок из командной строки

	// Добавление ссылок из командной строки на обработку
	go func() {
		links <- &batchWithDepth{0, flag.Args()}
	}()

	// Запуск ограниченного числа горутин для обработки новых ссылок
	for i := 0; i < *concurRate; i++ {
		go func() {
			for link := range unseenLinks {
				foundLinks := crawl(link)
				go func() {
					links <- foundLinks
				}()
			}
		}()
	}

	// Запись ссылок в файл
	linksWrite := make(chan *batchWithDepth)
	if *writeToFile {
		go func() {
			var w *bufio.Writer
			f, err := os.Create("./crawler_output.txt")
			if err != nil {
				log.Print(err)
			}
			defer f.Close()
			w = bufio.NewWriter(f)

			for list := range linksWrite {
				for _, link := range list.urls {
					if _, err := w.WriteString(link + "\n"); err != nil {
						log.Print(err)
					}
				}
				w.Flush()
			}
		}()
	}

	// main-горутина принимает найденные ссылки и ставит новые(необработанные)
	// ссылки на обработку.
	seen := make(map[string]bool)
	for ; n > 0; n-- {
		list := <-links
		if list.depth > *maxLinkDepth {
			continue
		}
		if *writeToFile {
			linksWrite <- list
		}
		for _, link := range list.urls {
			if !seen[link] {
				seen[link] = true
				n++
				unseenLinks <- &linkWithDepth{list.depth, link}
			}
		}

	}

}
