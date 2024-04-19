package link_parser

import (
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

func traverseTree(n *html.Node, p func(n *html.Node)) {
	if p != nil {
		p(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverseTree(c, p)
	}
}

func ExtractURLs(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("запрос URL [%s] : %s", url, resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге HTML : URL [%s] : %v", url, err)
	}

	var links []string
	// Обработчик узла
	processNode := func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key != "href" {
					continue
				}
				link, err := resp.Request.URL.Parse(a.Val)
				if err != nil {
					continue
				}
				links = append(links, link.String())
			}
		}
	}
	traverseTree(doc, processNode)
	return links, nil

}
