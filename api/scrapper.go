package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type UrlInfo struct {
	Site       string
	Url        string
	StartIndex int
}

type Chapter struct {
	Index    int
	Title    string
	Contents []string
	Err      error
}

type Credits struct {
	Author string
	Title  string
	URL    string
}

func (c *Client) GetSingleAddress(index int, address string, ch chan Chapter, wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		ch <- Chapter{Err: err}
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:158.0) Gecko/20100101 Firefox/158.0")

	res, err := c.HttpClient.Do(req)
	if err != nil {
		ch <- Chapter{Err: err}
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		ch <- Chapter{Err: fmt.Errorf("Error reaching %v: %v", address, res.Status)}
		return
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		ch <- Chapter{Err: err}
		return
	}
	chapterData, err := GetChapterData(string(data), index)
	if err != nil {
		ch <- Chapter{Err: err}
		return
	}
	ch <- chapterData
}

func GetSiteInfo(baseURL string) (UrlInfo, error) {
	urlS, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return UrlInfo{}, err
	}
	if urlS.Host != "ncode.syosetu.com" {
		return UrlInfo{}, errors.New("The url must come from ncode.syosetu.com")
	}
	urlParts := strings.Split(urlS.Path, "/")
	startIndex := 1
	if urlParts[2] != "" {
		index, err := strconv.Atoi(urlParts[2])
		if err != nil {
			return UrlInfo{}, err
		}
		startIndex = index
	}
	fullURL := url.URL{
		Scheme: urlS.Scheme,
		Host:   urlS.Host,
		Path:   urlParts[1],
	}
	return UrlInfo{
		Site:       urlParts[0],
		Url:        fullURL.String(),
		StartIndex: startIndex,
	}, nil
}

func (c *Client) GetChapterAuthor(url string) (Credits, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Credits{}, err
	}
	req.Header.Set("User-Agent", "export-novels-to-eReader-app")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return Credits{}, err
	}
	defer res.Body.Close()
	if res.StatusCode > 400 {
		return Credits{}, err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return Credits{}, err
	}
	titleFind := regexp.MustCompile(`<meta property="og:title" content="\s*(.*?)\s*">`)
	title := titleFind.FindAllStringSubmatch(string(data), 1)
	authorFind := regexp.MustCompile(`<meta name="twitter:creator" content="\s*(.*?)\s*">`)
	author := authorFind.FindAllStringSubmatch(string(data), 1)
	return Credits{
		Title:  title[0][1],
		Author: author[0][1],
		URL:    url,
	}, nil
}

func GetChapterData(data string, index int) (Chapter, error) {
	var lines []string
	titleFind := regexp.MustCompile(`<h1 class="p-novel__title p-novel__title--rensai">\s*(.*?)\s*</h1>`)
	title := titleFind.FindAllStringSubmatch(string(data), 1)
	topCrop := strings.Split(string(data), "<div class=\"js-novel-text p-novel__text\">")
	botCrop := strings.Split(topCrop[1], "</div>")
	filteredBody := strings.Split(botCrop[0], "</p>")
	reg := regexp.MustCompile(`<\s*(.*?)\s*>`)
	for _, line := range filteredBody {
		res := reg.ReplaceAllString(line, "")
		clean := strings.TrimSpace(res)
		lines = append(lines, clean)
	}
	if len(title[0]) < 2 {
		return Chapter{}, errors.New("Regex failed to find title")
	}
	return Chapter{
		Index:    index,
		Title:    title[0][1],
		Contents: lines,
	}, nil

}

func (c *Client) GetFromRange(urlData UrlInfo, chapterAmount int, ch chan Chapter, wg *sync.WaitGroup) {
	chapIndex := 0

	downloadBar := progressbar.Default(int64(chapterAmount), "Downloading chapters...")
	for i := urlData.StartIndex; i < urlData.StartIndex+chapterAmount; i++ {
		wg.Add(1)
		fullURL := fmt.Sprintf("%s/%d/", urlData.Url, i)
		go c.GetSingleAddress(chapIndex, fullURL, ch, wg)
		chapIndex++
		downloadBar.Add(1)
		if i%5 == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}
