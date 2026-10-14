package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
)

const (
	LINEHEIGHT    = 20.0
	MAXPAGEHEIGHT = 750.0
)

var (
	bodySize        int
	titleSize       int
	chapterAmount   int
	destinationPath string
	fileName        string
	oPDF            bool
	oHTML           bool
	oEPub           bool
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntVarP(&bodySize, "bodysize", "b", 14, "Size for the font of the body")
	getCmd.Flags().IntVarP(&titleSize, "titlesize", "t", 16, "Size for the font of the titles")
	getCmd.Flags().IntVarP(&chapterAmount, "volumesize", "v", 3, "Amount of chapter to download")
	getCmd.Flags().StringVarP(&destinationPath, "saveto", "s", "pfd", "Path to the folder where pdf's are stored")
	getCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name of the newly created pdf")
	getCmd.Flags().BoolVar(&oPDF, "pdf", false, "Output in PDF")
	getCmd.Flags().BoolVar(&oHTML, "html", false, "Output in HTML")
	getCmd.Flags().BoolVar(&oEPub, "epub", false, "Output in EPUB")
	getCmd.MarkFlagsOneRequired("pdf", "html", "epub")
	getCmd.MarkFlagsMutuallyExclusive("pdf", "html", "epub")

}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the urls",
	Long:  "Implement me",
	RunE:  GetContents,
}

func GetContents(cmd *cobra.Command, args []string) error {
	c := api.NewClient(5 * time.Second)

	for _, address := range args {
		err := GetCommand(cmd, address, &c)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCommand(cmd *cobra.Command, address string, c *api.Client) error {
	if oPDF {
		return GetPDF(cmd, address, c)
	}
	if oHTML {
		return GetHTML(cmd, address, c)
	}
	if oEPub {
		fmt.Println("Not implemented")
		return nil
	}
	return nil
}

func GetHTML(cmd *cobra.Command, address string, c *api.Client) error {
	if address == "" {
		return errors.New("No argument given")
	}

	var urlData api.UrlInfo
	urlData, err := api.GetSiteInfo(address)
	if err != nil {
		return err
	}

	var bookCredits api.Credits
	bookCredits, err = c.GetChapterAuthor(urlData.Url)
	if err != nil {
		return err
	}
	start := `<!DOCTYPE html>
<html lang="jp">
<head>
    <meta charset="UTF-8">
    <title>` + bookCredits.Title + `</title>
    <style>
        body {
            font-family: serif;
            padding: 0;
            margin: 0;
        }
        p {
            text-indent: 1.5em; /* Standard book paragraph indent */
            margin-top: 0;
            margin-bottom: 0;
            text-align: justify; /* Kindle readers expect justified text */
        }
        h1, h2, h3 {
            text-align: center;
            text-indent: 0; /* Don't indent headings */
            margin-top: 1em;
            margin-bottom: 0.5em;
        }
        .first-para {
            text-indent: 0; /* Optional: First paragraph of a chapter usually isn't indented */
        }
    </style>
	</head>
	<body>`
	for i := urlData.StartIndex; i < urlData.StartIndex+chapterAmount; i++ {
		fullURL := fmt.Sprintf("%s/%d/", urlData.Url, i)
		newStart, err := c.GetHTML(fullURL, start)
		if err != nil {
			return err
		}
		start = newStart
	}

	filename := destinationPath + "/" + fileName + ".html"

	err = os.WriteFile(filename, []byte(start), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully saved HTML to %s\n", filename)
	return nil
}

func GetPDF(cmd *cobra.Command, address string, c *api.Client) error {
	if address == "" {
		return errors.New("No argument given")
	}

	var urlData api.UrlInfo
	urlData, err := api.GetSiteInfo(address)
	if err != nil {
		return err
	}

	var chaptersData []api.Chapter
	for i := urlData.StartIndex; i < urlData.StartIndex+chapterAmount; i++ {
		fullURL := fmt.Sprintf("%s/%d/", urlData.Url, i)
		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "export-novels-to-eReader-app")
		res, err := c.HttpClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return fmt.Errorf("Error reaching %v: %v\n", fullURL, res.Status)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		chapterData, err := api.GetChapterData(string(data))
		if err != nil {
			return err
		}
		chaptersData = append(chaptersData, chapterData)
	}
	bookCredits, err := c.GetChapterAuthor(urlData.Url)
	if err != nil {
		return err
	}
	pdfAddress, err := CreatePDF(chaptersData, bookCredits, fileName)
	if err != nil {
		return err
	}
	fmt.Printf("PDF created successfully at: %v", pdfAddress)
	return nil
}

func CreatePDF(chapterData []api.Chapter, credits api.Credits, fileName string) (string, error) {
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

		err = pdf.SetFont("NotoSansJP", "B", titleSize)
		if err != nil {
			return "", err
		}

		// Set title
		pdf.SetX((pdf.GetX() / 2) + 100)
		pdf.CellWithOption(&gopdf.Rect{W: 400, H: 20}, chapter.Title,
			gopdf.CellOption{Align: gopdf.Justify | gopdf.Center})
		pdf.Br(40)

		// Set body
		err = pdf.SetFont("NotoSansJP", "", bodySize)
		if err != nil {
			return "", err
		}
		pdf.SetX(50)
		for _, line := range chapter.Contents {
			if line == "" {
				continue
			}

			if pdf.GetY() > MAXPAGEHEIGHT {
				pdf.AddPage()
				pdf.SetX(50)
				_ = pdf.SetFont("NotoSansJP", "", bodySize)
			}

			bodyRect := &gopdf.Rect{W: 500, H: LINEHEIGHT}
			err = pdf.MultiCell(bodyRect, line)
			if err != nil {
				return "", err
			}
		}
	}
	pdf.AddPage()
	err = pdf.SetFont("NotoSansJP", "B", titleSize)
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

	fullPath := fmt.Sprintf("%s/%s.pdf", destinationPath, fileName)
	if fileName == "" {
		fileName = credits.Title
	}
	err = pdf.WritePdf(fullPath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}
