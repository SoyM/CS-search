package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/PuerkitoBio/goquery"
	"unicode"
	"net/url"
	"encoding/json"
	"os"
)

type SearchResult struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}

type SearchResultsSlice struct {
	SearchResults []SearchResult
}

func searchZhihu(searchKey string) (srs SearchResultsSlice) {
	reqUrl := "https://www.zhihu.com/search?type=content&q=" + searchKey
	fmt.Println(reqUrl)
	res, err := http.Get(reqUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	//data, err2 := ioutil.ReadAll(res.Body)
	//if err2 != nil {
	//	log.Fatalf("readAll error: %s", err2)
	//	return
	//}
	//fmt.Print(string(data))

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	} else {
		var srs SearchResultsSlice
		doc.Find(".Search-container .AnswerItem").Each(func(i int, s *goquery.Selection) {
			a := s.Find("a").Eq(0)
			a.Each(func(i2 int, content *goquery.Selection) {
				linkUrl, _ := content.Attr("href")
				title := a.Text()
				srs.SearchResults = append(srs.SearchResults, SearchResult{linkUrl, title,})
			})
		})
		return srs
	}
}

func searchLeiphone(searchKey string) (SearchResultsSlice) {
	reqUrl := "https://www.leiphone.com/search?s=" + searchKey + "&site=article"

	res, err := http.Get(reqUrl)
	fmt.Println(reqUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// 输出html
	//data, err2 := ioutil.ReadAll(res.Body)
	//if err2 != nil {
	//	log.Fatalf("readAll error: %s", err2)
	//	return
	//}
	//fmt.Print(string(data))

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var srs SearchResultsSlice
	doc.Find("ul[class=\"articleList\"]").Eq(0).Find("li").Each(func(i int, s *goquery.Selection) {
		a := s.Find("a").Eq(0)
		a.Next()
		a.Each(func(i2 int, content *goquery.Selection) {
			linkUrl, _ := content.Attr("href")
			title := a.Text()
			srs.SearchResults = append(srs.SearchResults, SearchResult{linkUrl, title,})
		})
	})
	return srs
}

func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	//start := time.Now()
	searchKey := r.URL.Query()["q"][0]
	searchWhere := r.URL.Query()["w"][0]
	var searchKeyEn string
	if IsChineseChar(searchKey) {
		searchKeyEn, _ = UrlEncoded(searchKey)
	} else {
		searchKeyEn = searchKey
	}
	var srs SearchResultsSlice
	if searchWhere == "zhihu" {
		srs = searchZhihu(searchKeyEn)
	} else if searchWhere == "leiphone" {
		srs = searchLeiphone(searchKeyEn)
	}

	if data, err := json.Marshal(srs); err == nil {
		os.Stdout.Write(data)
		fmt.Fprintln(w, string(data))
	} else {
		fmt.Println("error:", err)
	}

	//fmt.Fprintln(w, time.Since(start))
}

func main() {
	//http.HandleFunc("/", search)
	http.HandleFunc("/search", search)

	if err := http.ListenAndServe("0.0.0.0:8081", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("start")

}
