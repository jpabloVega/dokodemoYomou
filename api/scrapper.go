package api

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/signintech/gopdf"
)

const (
	LINEHEIGHT    = 20.0
	MAXPAGEHEIGHT = 750.0
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

type Credits struct {
	Author string
	Title  string
	URL    string
}

func (c *Client) GetAddresses(startAddress string) (string, error) {
	if startAddress == "" {
		return "", nil
	}
	urlData, err := getSiteInfo(startAddress)
	if err != nil {
		return "", err
	}
	var pdfAddress string
	var chaptersData []Chapter
	for i := urlData.startIndex; i < urlData.startIndex+5; i++ {
		fullURL := fmt.Sprintf("%s/%d/", urlData.url, i)
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "export-novels-to-eReader-app")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		if res.StatusCode > 400 {
			break
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		chapterData, err := getChapterData(string(data))
		if err != nil {
			return "", err
		}
		chaptersData = append(chaptersData, chapterData)
	}
	bookCredits, err := c.getChapterAuthor(urlData.url)
	if err != nil {
		return "", err
	}
	pdfAddress, err = createPDF(chaptersData, bookCredits)
	if err != nil {
		return "", err
	}
	return "Pdf created at: " + pdfAddress, nil
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

func (c *Client) getChapterAuthor(url string) (Credits, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Credits{}, err
	}
	req.Header.Set("User-Agent", "export-novels-to-eReader-app")
	res, err := c.httpClient.Do(req)
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

func createPDF(chapterData []Chapter, credits Credits) (string, error) {
	// Create pdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Set fonts
	err := pdf.AddTTFFontWithOption("NotoSansJP", "ttf/static/NotoSansJP-Regular.ttf", gopdf.TtfOption{Style: gopdf.Regular})
	if err != nil {
		return "", err
	}
	err = pdf.AddTTFFontWithOption("NotoSansJP", "ttf/static/NotoSansJP-Bold.ttf", gopdf.TtfOption{Style: gopdf.Bold})
	if err != nil {
		return "", err
	}
	for _, chapter := range chapterData {

		pdf.AddPage()

		err = pdf.SetFont("NotoSansJP", "B", 16)
		if err != nil {
			return "", err
		}

		// Set title
		pdf.SetX((pdf.GetX() / 2) + 100)
		pdf.CellWithOption(&gopdf.Rect{W: 400, H: 20}, chapter.title,
			gopdf.CellOption{Align: gopdf.Justify | gopdf.Center})
		pdf.Br(40)

		// Set body
		err = pdf.SetFont("NotoSansJP", "", 14)
		if err != nil {
			return "", err
		}
		pdf.SetX(50)
		for _, line := range chapter.contents {
			if line == "" {
				continue
			}

			if pdf.GetY() > MAXPAGEHEIGHT {
				pdf.AddPage()
				pdf.SetX(50)
				_ = pdf.SetFont("NotoSansJP", "", 14)
			}

			bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}
			err = pdf.MultiCell(bodyRect, line)
			if err != nil {
				return "", err
			}
		}
	}
	pdf.AddPage()
	err = pdf.SetFont("NotoSansJP", "B", 16)
	if err != nil {
		return "", err
	}

	chapterTitle := fmt.Sprintf("書名： %s", credits.Title)
	bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}

	pdf.Cell(bodyRect, chapterTitle)
	pdf.Br(40)

	authorCredit := fmt.Sprintf("作者： %s", credits.Author)
	pdf.Cell(bodyRect, authorCredit)
	pdf.Br(40)

	chapterUrl := fmt.Sprintf("URL： %s", credits.URL)
	pdf.Cell(bodyRect, chapterUrl)
	pdf.Br(40)

	pdfPath := "tmpPDF/hello3rd.pdf"
	err = pdf.WritePdf(pdfPath)
	if err != nil {
		return "", err
	}
	return pdfPath, nil
}
