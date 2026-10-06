package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type urlInfo struct {
	site       string
	url        string
	startIndex int
}

type Chapter struct {
	title    string
	contents []string
}

func (c *Client) GetAddresses(startAddress string) ([]Chapter, error) {
	var found []Chapter
	if startAddress == "" {
		return []Chapter{}, errors.New("Please introduce a url")
	}
	urlData, err := getSiteInfo(startAddress)
	if err != nil {
		return []Chapter{}, err
	}
	for i := urlData.startIndex; i < urlData.startIndex+1; i++ {
		fullURL := fmt.Sprintf("%s/%d/", urlData.url, i)
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return []Chapter{}, err
		}
		req.Header.Set("User-Agent", "export-novels-to-eReader-app")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return []Chapter{}, err
		}
		defer res.Body.Close()
		if res.StatusCode > 400 {
			break
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return []Chapter{}, errors.New("Bad data")
		}
		newChapter, err := getChapterData(string(data))
		fmt.Printf("Title: %s\n", newChapter.title)
		for _, line := range newChapter.contents {
			fmt.Println(line)
		}
		found = append(found, newChapter)
	}
	return found, nil
}

func getSiteInfo(baseURL string) (urlInfo, error) {
	url, removed := strings.CutPrefix(baseURL, "https://")
	urlParts := strings.Split(url, "/")
	fullURL := strings.Join(urlParts[:2], "/")
	startIndex := 1
	if len(urlParts) > 2 && urlParts[2] != "" {
		index, err := strconv.Atoi(urlParts[2])
		if err != nil {
			return urlInfo{}, err
		}
		startIndex = index
	}
	if removed {
		fullURL = fmt.Sprintf("%s%s", "https://", fullURL)
	}
	return urlInfo{
		site:       urlParts[0],
		url:        fullURL,
		startIndex: startIndex,
	}, nil
}

func getChapterData(data string) (Chapter, error) {
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
		title:    title[0][1],
		contents: lines,
	}, nil

}
