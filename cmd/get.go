package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/schollz/progressbar/v3"
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
	vertical        bool
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntVarP(&bodySize, "bodysize", "b", 14, "Size for the font of the body (only for pdf)")
	getCmd.Flags().IntVarP(&titleSize, "titlesize", "t", 16, "Size for the font of the titles (only for pdf)")
	getCmd.Flags().IntVarP(&chapterAmount, "amount", "a", 3, "Amount of chapters to download (default 3)")
	getCmd.Flags().StringVarP(&destinationPath, "saveto", "s", "pfd", "Path to the folder where pdf's are stored")
	getCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name of the newly created pdf")
	getCmd.Flags().BoolVarP(&vertical, "vertical", "v", false, "Export text in vertical(only availible to html format)")
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

	chaptersChan := make(chan api.Chapter, chapterAmount)
	var wg sync.WaitGroup

	fmt.Println("Started go routine")

	c.GetFromRange(urlData, chapterAmount, chaptersChan, &wg)

	go func() {
		wg.Wait()
		close(chaptersChan)
	}()

	fmt.Println("Finished go routine")

	chaptersData := make([]api.Chapter, chapterAmount)
	for res := range chaptersChan {
		chaptersData[res.Index] = res
	}

	var bookCredits api.Credits
	bookCredits, err = c.GetChapterAuthor(urlData.Url)
	if err != nil {
		return err
	}

	var vertFormat string
	if vertical {
		vertFormat = `
			writing-mode: vertical-rl;
            -webkit-writing-mode: vertical-rl;`
	}

	var html strings.Builder
	header := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <title>` + bookCredits.Title + `</title>
    <style>
        body {
            font-family: serif;
            padding: 0;
            margin: 0;` + vertFormat + `
        }
        p {
            text-indent: 0;
            margin-top: 0;
            margin-bottom: 0;
            text-align: justify;
        }
        h1, h2, h3 {
            text-align: center;
            text-indent: 0;
            margin-top: 1em;
            margin-bottom: 0.5em;
        }
    </style>
</head>
<body>

`

	html.WriteString(header)
	var lineAmount int
	for _, chap := range chaptersData {
		lineAmount += len(chap.Contents)
	}
	chapterDownload := progressbar.Default(int64(lineAmount), "Creating html...")

	for _, chapter := range chaptersData {
		fmt.Fprintf(&html, `	<h1>%s</h1>

`, chapter.Title)
		for _, body := range chapter.Contents {
			chapterDownload.Add(1)
			if body == "" {
				continue
			}
			fmt.Fprintf(&html, `    <p>%s</p>
`, body)

		}
		html.WriteString(`

    <mbp:pagebreak/>

`)
		chapterDownload.Add(1)
	}

	chapterTitle := fmt.Sprintf("書名： %s", bookCredits.Title)
	authorCredit := fmt.Sprintf("作者： %s", bookCredits.Author)
	chapterUrl := fmt.Sprintf("URL： %s", bookCredits.URL)

	fmt.Fprintf(&html, `    <h1>%s</h1>
	<h1>%s</h1>
    <h1>%s</h1>
</body>
</html>`, chapterTitle, authorCredit, chapterUrl)

	filename := destinationPath + "/" + fileName + ".html"

	err = os.WriteFile(filename, []byte(html.String()), 0644)
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

	chaptersChan := make(chan api.Chapter, chapterAmount)
	var wg sync.WaitGroup

	c.GetFromRange(urlData, chapterAmount, chaptersChan, &wg)

	go func() {
		wg.Wait()
		close(chaptersChan)
	}()

	chaptersData := make([]api.Chapter, chapterAmount)
	for res := range chaptersChan {
		chaptersData[res.Index] = res
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

	var lineAmount int
	for _, chap := range chapterData {
		lineAmount += len(chap.Contents)
	}
	pdfProgress := progressbar.Default(int64(lineAmount), "Creating PDF...")

	for _, chapter := range chapterData {

		if chapter.Err != nil {
			continue
		}

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
			pdfProgress.Add(1)
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
