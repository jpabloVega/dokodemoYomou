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
	"unicode/utf8"

	vocabtrie "github.com/jpabloVega/dokodemoYomou/dictionary/vocab_trie"
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

func ScanTest() error {
	t := vocabtrie.NewTrie()
	err := t.LoadFromFile("dictionary/vocab_trie/vocabulary.json")
	if err != nil {
		return err
	}
	nCount := make(map[string]int)
	contentsLine := "ある日、オリヴィアは夢を見た。婚約者のデイルが、義妹のグレースを好きだと言い、グレースも、デイルが好きだったと打ち明けられる夢"
	splitLine := []rune(contentsLine)
	var foundWords []string
	start := 0
	end := len(splitLine)
	for true {
		currWord := string(splitLine[start:end])
		lvl, found := t.Search(currWord)
		if found {
			nCount[lvl] += 1
			start += utf8.RuneCountInString(currWord)
			end = len(splitLine)
			foundWords = append(foundWords, currWord)
			continue
		}
		end -= 1
		if end == start {
			start += 1
			end = len(splitLine)
		}
		if start >= len(splitLine)-1 {
			break
		}
	}
	fmt.Println("Words found: ")
	for _, word := range foundWords {
		fmt.Printf("- %v\n", word)
	}
	fmt.Println("Words from every level")
	for level, amount := range nCount {
		fmt.Printf("%s: %d\n", level, amount)
	}
	return nil
}

func (c *Client) ScanBook(address string) error {
	urlData, err := GetSiteInfo(address)
	if err != nil {
		return err
	}
	chapterData, err := c.getSingleAddress(urlData)
	if err != nil {
		return err
	}
	t := vocabtrie.NewTrie()
	err = t.LoadFromFile("dictionary/vocab_trie/vocabulary.json")
	if err != nil {
		return err
	}
	nCount := make(map[string]int)
	var foundWords []string
	for _, line := range chapterData.Contents {
		splitLine := []rune(line)
		start := 0
		end := len(splitLine)
		for true {
			currWord := string(splitLine[start:end])
			lvl, found := t.Search(currWord)
			if found {
				nCount[lvl] += 1
				start += utf8.RuneCountInString(currWord)
				end = len(splitLine)
				foundWords = append(foundWords, currWord)
				continue
			}
			end -= 1
			if end == start {
				start += 1
				end = len(splitLine)
			}
			if start >= len(splitLine)-1 {
				break
			}
		}
	}
	fmt.Println("Words found: ")
	for _, word := range foundWords {
		fmt.Printf("- %v\n", word)
	}
	fmt.Println("Words from every level")
	for level, amount := range nCount {
		fmt.Printf("%s: %d\n", level, amount)
	}
	return nil
}

func (c *Client) getSingleAddress(address UrlInfo) (Chapter, error) {
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
	return Chapter{
		Title:    title[0][1],
		Contents: lines,
	}, nil

}
