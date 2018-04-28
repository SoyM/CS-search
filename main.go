package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"log"
	"time"
	"github.com/PuerkitoBio/goquery"
)

var (
	ptnIndexItem    = regexp.MustCompile(`<a target="_blank" href="(.+\.html)" title=".+" >(.+)</a>`)
	ptnContentRough = regexp.MustCompile(`(?s).*<div class="artcontent">(.*)<div id="zhanwei">.*`)
	ptnBrTag        = regexp.MustCompile(`<br>`)
	ptnHTMLTag      = regexp.MustCompile(`(?s)</?.*?>`)
	ptnSpace        = regexp.MustCompile(`(^\s+)|( )`)
)

func Get(url string) (content string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	data, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		log.Fatalf("readAll error: %s", err2)
		return
	}
	content = string(data)
	return
}

type IndexItem struct {
	url   string
	title string
}

func findIndex(content string) (index []IndexItem, err error) {
	matches := ptnIndexItem.FindAllStringSubmatch(content, 10000)
	index = make([]IndexItem, len(matches))
	for i, item := range matches {
		index[i] = IndexItem{"http://www.yifan100.com" + item[1], item[2]}
	}
	return
}

func readContent(url string) (content string) {
	raw := Get(url)

	match := ptnContentRough.FindStringSubmatch(raw)
	if match != nil {
		content = match[1]
	} else {
		return
	}

	content = ptnBrTag.ReplaceAllString(content, "\r\n")
	content = ptnHTMLTag.ReplaceAllString(content, "")
	content = ptnSpace.ReplaceAllString(content, "")
	return
}

func search_zhihu()(dataUrl []string, dataTitle []string) {
	url := "https://www.zhihu.com/search?type=content&q=%E6%9C%BA%E5%99%A8%E5%AD%A6%E4%B9%A0"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	//data, err2 := ioutil.ReadAll(res.Body)
	//if err2 != nil {
	//	log.Fatalf("readAll error: %s", err2)
	//	return
	//}
	//fmt.Print(string(data))

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".Search-container .AnswerItem").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a")
		a.Each(func(i2 int, content *goquery.Selection) {
			url, _ := content.Attr("href")
			title := a.Text()
			dataUrl = append(dataUrl, url)
			dataTitle = append(dataTitle, title)
		})
	})
	return dataUrl, dataTitle
}

func search_leiphone() (dataUrl []string, dataTitle []string) {
	url := "https://www.leiphone.com/search?site=&s=%E6%9C%BA%E5%99%A8%E5%AD%A6%E4%B9%A0"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("ul[class=\"articleList\"]").Eq(0).Find("li").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a").Eq(0)
		a.Next()
		a.Each(func(i2 int, content *goquery.Selection) {
			url, _ := content.Attr("href")
			title := a.Text()
			dataUrl = append(dataUrl, url)
			dataTitle = append(dataTitle, title)
		})
	})
	return dataUrl, dataTitle
}

func search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	dataUrl, dataTitle := search_leiphone()
	fmt.Fprintln(w, dataUrl)
	fmt.Fprintln(w, dataTitle)
	fmt.Fprintln(w,"-------------------------------------------------")
	dataUrl, dataTitle = search_zhihu()
	fmt.Fprintln(w, dataUrl)
	fmt.Fprintln(w, dataTitle)
	//输出执行时间，单位为毫秒。
	fmt.Fprintln(w,time.Since(start))
}

func main() {
	http.HandleFunc("/", search)

	err := http.ListenAndServe(":9090", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	fmt.Println("start")

	//index, _ := findIndex(s)
	//fmt.Println(`Get contents and write to file ...`)
	//for _, item := range index {
	//	fmt.Printf("Get content %s from %s and write to file.\n", item.title, item.url)
	//	fileName := fmt.Sprintf("%s.txt", item.title)
	//	content := readContent(item.url)
	//	ioutil.WriteFile(fileName, []byte(content), 0644)
	//	fmt.Printf("Finish writing to %s.\n", fileName)
	//}
}
