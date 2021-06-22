package Crawling

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type news struct {
	title   string
	summary string
	link    string
}

var baseurl string = "https://www.hankyung.com/economy?date=" + strings.Split(time.Now().String(), " ")[0]

func Crawl() {

	mainC := make(chan []news)

	var news_site []news
	totalpages := getPages()
	for i := 1; i <= totalpages; i++ {
		go getPage(i, mainC)
	}
	for i := 0; i < totalpages; i++ {
		news_page := <-mainC
		news_site = append(news_site, news_page...)
	}
	writeNews(news_site)
}

func getPages() int {

	pagenum := 0
	res, err := http.Get(baseurl)
	getErr(err)
	getErrStatus(res)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	getErr(err)
	doc.Find(".paging").Each(func(i int, s *goquery.Selection) {
		a, _ := doc.Find(".end").Attr("href")
		re := regexp.MustCompile("[0-9]+")
		num := re.FindAllString(a, -1)[3]

		pagenum, _ = strconv.Atoi(num)

	})
	return pagenum
}

func writeNews(news_site []news) {

	file, err := os.Create("news.csv")
	getErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Title", "Summary", "Link"}

	wErr := w.Write(headers)
	getErr(wErr)

	for _, news := range news_site {
		news_slice := []string{news.title, news.summary, news.link}
		wErr := w.Write(news_slice)
		getErr(wErr)
	}

}

func getPage(i int, mainC chan<- []news) {

	c := make(chan news)

	var news_page []news
	pageURL := baseurl + "&page=" + strconv.Itoa(i)
	fmt.Println(pageURL)
	res, err := http.Get(pageURL)

	getErr(err)
	defer res.Body.Close()
	getErrStatus(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	getErr(err)
	articles := doc.Find(".article")
	articles.Each(func(i int, s *goquery.Selection) {
		go extractNews(s, c)
	})

	for i := 0; i < articles.Length(); i++ {
		news_s := <-c
		news_page = append(news_page, news_s)
	}
	mainC <- news_page
}

func extractNews(s *goquery.Selection, c chan<- news) {

	title := cleanstr(s.Find("h3").Text())
	summary := cleanstr(s.Find(".read").Text())
	link_o, _ := s.Find(".tit>a").Attr("href")
	link := cleanstr(link_o)

	c <- news{
		title:   title,
		summary: summary,
		link:    link}
}

func cleanstr(str string) string {
	return strings.TrimSpace(str)
}

func getErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getErrStatus(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalf("Status code error: %d", res.StatusCode)
	}
}
