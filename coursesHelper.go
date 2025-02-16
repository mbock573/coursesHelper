package coursesHelper

import (
	"bytes"
	"fmt"
	"github.com/mbock573/httpClientHelper"
	"golang.org/x/net/html"
	"io"
	"strings"
)

const moduledb_htwsaarURL = "moduledb.htwsaar.de"

func Run() (map[string]string, error) {
	client := httpClientHelper.NewClient()
	httpResult, err := httpClientHelper.HttpGetRequest(client, moduledb_htwsaarURL)
	if err != nil {
		fmt.Printf("CoursesHelper Error: %v\n", err)
	}
	bodyBytes, err := io.ReadAll(httpResult.Body)
	if err != nil {
		fmt.Printf("CourseHelper Error: %v\n", err)
	}
	availableCourses, err := courseParser(bodyBytes)
	return availableCourses, err
}

func courseParser(htmlBody []byte) (map[string]string, error) {
	coursesMap := make(map[string]string)
	doc, err := html.Parse(bytes.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	// Finde die Tabelle mit der Klasse "pretty-table"
	var table *html.Node
	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "pretty-table" {
					table = n
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTable(c)
			if table != nil {
				return
			}
		}
	}
	findTable(doc)
	if table == nil {
		return nil, fmt.Errorf("Tabelle mit Klasse 'pretty-table' nicht gefunden")
	}

	// Finde tbody
	var tbody *html.Node
	for c := table.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "tbody" {
			tbody = c
			break
		}
	}
	if tbody == nil {
		return nil, fmt.Errorf("tbody in Tabelle nicht gefunden")
	}

	// Verarbeite jede Zeile, überspringe die Kopfzeile
	rowIndex := 0
	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type != html.ElementNode || tr.Data != "tr" {
			continue
		}
		rowIndex++
		if rowIndex == 1 { // Überspringe Kopfzeile
			continue
		}

		// Sammle alle td-Elemente der Zeile
		var tds []*html.Node
		for td := tr.FirstChild; td != nil; td = td.NextSibling {
			if td.Type == html.ElementNode && td.Data == "td" {
				tds = append(tds, td)
			}
		}
		if len(tds) < 7 {
			continue // Nicht genug Spalten
		}

		// Extrahiere URL aus dem ersten td
		var anchor *html.Node
		for c := tds[0].FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "a" {
				anchor = c
				break
			}
		}
		if anchor == nil {
			continue // Kein Link gefunden
		}

		// Hole href-Attribut
		var href string
		for _, attr := range anchor.Attr {
			if attr.Key == "href" {
				href = html.UnescapeString(attr.Val)
				break
			}
		}
		if href == "" {
			continue // href ist leer
		}

		// Extrahiere Kürzel aus dem siebten td
		abbr := getTextContent(tds[6])
		if abbr == "" {
			continue // Kein Kürzel vorhanden
		}

		coursesMap[abbr] = href
	}

	return coursesMap, nil
}

// getTextContent extrahiert Textinhalte eines Nodes
func getTextContent(n *html.Node) string {
	var text strings.Builder
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return strings.TrimSpace(text.String())
}
