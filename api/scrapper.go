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
)

type UrlInfo struct {
	Site       string
	Url        string
	StartIndex int
}

type Chapter struct {
	Title    string
	Contents []string
}

type Credits struct {
	Author string
	Title  string
	URL    string
}

func (c *Client) GetSingleAddress(address UrlInfo) (Chapter, error) {
	fullURL := fmt.Sprintf("%s/%d/", address.Url, address.StartIndex)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return Chapter{}, err
	}
	req.Header.Set("User-Agent", "export-novels-to-eReader-app")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return Chapter{}, err
	}
	defer res.Body.Close()
	if res.StatusCode > 400 {
		return Chapter{}, fmt.Errorf("Error: %v\n", res.Status)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return Chapter{}, err
	}
	chapterData, err := GetChapterData(string(data))
	if err != nil {
		return Chapter{}, err
	}
	return chapterData, err
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

func GetChapterData(data string) (Chapter, error) {
	var lines []string
	titleFind := regexp.MustCompile(`<h1 class="p-novel__title p-novel__title--rensai">\s*(.*?)\s*</h1>`)
	title := titleFind.FindAllStringSubmatch(string(data), 1)
	topCrop := strings.Split(string(data), "<div class=\"js-novel-text p-novel__text\">")
	botCrop := strings.Split(topCrop[1], "</div>")
	filteredBody := strings.Split(botCrop[0], "</p>")
	reg := regexp.MustCompile(`<\s*(.*?)\s*>`)
	for _, line := range filteredBody {
		res := reg.ReplaceAllString(line, "")
		lines = append(lines, res)
	}
	if len(title[0]) < 2 {
		return Chapter{}, errors.New("Regex failed to find title")
	}
	return Chapter{
		Title:    title[0][1],
		Contents: lines,
	}, nil

}

func (c *Client) GetHTML(url, start string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "export-novels-to-eReader-app")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("Error reaching %v: %v\n", url, res.Status)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	htmlData, _ := GetRawChapterData(string(data), start)
	fmt.Println(htmlData)
	return htmlData, nil
}

func TestGetHTML(url string, c *Client) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "export-novels-to-eReader-app")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("Error reaching %v: %v\n", url, res.Status)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	start := `<div style="page-break-after: always;"></div>
	`
	htmlData, _ := GetRawChapterData(string(data), start)
	fmt.Println(htmlData)
	return htmlData, nil
}

func GetRawChapterData(data, start string) (string, error) {
	topCroped := strings.Split(data, `<h1 class="p-novel__title p-novel__title--rensai">`)
	novel := strings.Split(topCroped[1], `</div><!-- /.p-novel__body -->`)
	step := `
	`
	return start + `<h1 class="p-novel__title p-novel__title--rensai">` + novel[0] + step, nil
}
