package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"uuid"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/jpabloVega/dokodemoYomou/database"
	"github.com/jpabloVega/dokodemoYomou/email"
	"github.com/raitucarp/epub"
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
	destinationPath string
	fileName        string
	oPDF            bool
	oHTML           bool
	oEPub           bool
	vertical        bool
	mailToKindle    bool
)

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntVar(&bodySize, "bodysize", 14, "Size for the font of the body (only for pdf)")
	getCmd.Flags().IntVar(&titleSize, "titlesize", 16, "Size for the font of the titles (only for pdf)")
	getCmd.Flags().IntVarP(&chapterAmount, "volumeSize", "", 10, "Amount of chapters to download")
	getCmd.Flags().IntVar(&startFrom, "start", 0, "Start from chapter")
	getCmd.Flags().StringVarP(&destinationPath, "saveto", "s", "pfd", "Path to the folder where pdf's are stored")
	getCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name of the newly created pdf")
	getCmd.Flags().BoolVarP(&vertical, "vertical", "v", false, "Export text in vertical(only availible to html format)")
	getCmd.Flags().BoolVarP(&mailToKindle, "mailToKindle", "m", false, "Mail to kindle")
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
	Args:  cobra.ExactArgs(1),
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
	if address == "" {
		return errors.New("No argument given")
	}

	var urlData api.UrlInfo
	urlData, err := api.GetSiteInfo(address)
	if err != nil {
		return err
	}

	db, err := database.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	bookExists, dbBook, err := database.GetBook(db, urlData.Url)
	if bookExists {
		urlData.StartIndex = dbBook.LastDownload + 1
	}

	if startFrom != 0 {
		urlData.StartIndex = startFrom
	}

	chaptersChan := make(chan api.Chapter, chapterAmount)
	var wg sync.WaitGroup

	c.GetFromRange(urlData, chapterAmount, chaptersChan, &wg)

	go func() {
		wg.Wait()
		close(chaptersChan)
	}()

	chaptersData := make([]api.Chapter, chapterAmount)
	var lastDownload = urlData.StartIndex
	for res := range chaptersChan {
		if res.Err != nil {
			fmt.Printf("Error in chapter %s: %v", res.Title, res.Err)
			continue
		}
		lastDownload++
		chaptersData[res.Index] = res
	}

	if len(chaptersData) == 0 {
		return errors.New("No chapters to create document")
	}

	var bookCredits api.Credits
	bookCredits, err = c.GetChapterAuthor(urlData.Url)
	if err != nil {
		return err
	}

	var lineAmount int
	for _, chap := range chaptersData {
		lineAmount += len(chap.Contents)
	}

	var docs []string

	if oPDF {
		docs, err = GetPDF(chaptersData, bookCredits, lineAmount)
	}
	if oHTML {
		docs, err = GetHTML(chaptersData, bookCredits, lineAmount)
	}
	if oEPub {
		docs, err = GetEPub(chaptersData, bookCredits, lineAmount)
	}
	if err != nil {
		return err
	}

	newBook := database.DBBook{
		Name:         bookCredits.Title,
		Link:         urlData.Url,
		LastDownload: lastDownload,
		Like:         false,
		DownloadedAt: time.Now(),
	}
	err = database.SaveBook(db, newBook)
	if err != nil {
		return err
	}

	if mailToKindle {
		err = email.MailToKindle(docs)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetHTML(chaptersData []api.Chapter, bookCredits api.Credits, lineAmount int) ([]string, error) {
	htmlProgress := progressbar.Default(int64(lineAmount), "Creating html...")

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

	for _, chapter := range chaptersData {
		fmt.Fprintf(&html, `	<h1>%s</h1>

`, chapter.Title)
		for _, body := range chapter.Contents {
			htmlProgress.Add(1)
			if body == "" {
				continue
			}
			fmt.Fprintf(&html, `    <p>%s</p>
`, body)

		}
		html.WriteString(`

    <mbp:pagebreak/>

`)
		htmlProgress.Add(1)
	}

	chapterTitle := fmt.Sprintf("書名： %s", bookCredits.Title)
	authorCredit := fmt.Sprintf("作者： %s", bookCredits.Author)
	chapterUrl := fmt.Sprintf("URL： %s", bookCredits.URL)

	fmt.Fprintf(&html, `    <h1>%s</h1>
	<h1>%s</h1>
    <h1>%s</h1>
</body>
</html>`, chapterTitle, authorCredit, chapterUrl)

	fileNameExt := fmt.Sprintf("%v.html", fileName)
	filePath := filepath.Join(destinationPath, fileNameExt)

	err := os.WriteFile(filePath, []byte(html.String()), 0644)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Successfully created HTML at %s\n", filePath)
	return []string{filePath}, nil
}

func GetPDF(chapterData []api.Chapter, credits api.Credits, lineAmount int) ([]string, error) {
	pdfProgress := progressbar.Default(int64(lineAmount), "Creating PDF...")

	// Create pdf
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	// Set fonts
	err := pdf.AddTTFFontWithOption("NotoSansJP", "ttf/static/NotoSansJP-Regular.ttf", gopdf.TtfOption{Style: gopdf.Regular})
	if err != nil {
		return nil, err
	}
	err = pdf.AddTTFFontWithOption("NotoSansJP", "ttf/static/NotoSansJP-Bold.ttf", gopdf.TtfOption{Style: gopdf.Bold})
	if err != nil {
		return nil, err
	}

	for _, chapter := range chapterData {

		if chapter.Err != nil {
			continue
		}

		pdf.AddPage()

		err = pdf.SetFont("NotoSansJP", "B", titleSize)
		if err != nil {
			return nil, err
		}

		// Set title
		pdf.SetX((pdf.GetX() / 2) + 100)
		pdf.CellWithOption(&gopdf.Rect{W: 400, H: 20}, chapter.Title,
			gopdf.CellOption{Align: gopdf.Justify | gopdf.Center})
		pdf.Br(40)

		// Set body
		err = pdf.SetFont("NotoSansJP", "", bodySize)
		if err != nil {
			return nil, err
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
				return nil, err
			}
		}
	}
	pdf.AddPage()
	err = pdf.SetFont("NotoSansJP", "B", titleSize)
	if err != nil {
		return nil, err
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

	if fileName == "" {
		fileName = credits.Title
	}
	fileNameExt := fmt.Sprintf("%v.pdf", fileName)
	filePath := filepath.Join(destinationPath, fileNameExt)

	err = pdf.WritePdf(filePath)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Pdf successfully created at: %s\n", filePath)
	return []string{filePath}, nil
}

func GetEPub(chapterData []api.Chapter, credits api.Credits, lineAmount int) ([]string, error) {
	ePubProgress := progressbar.Default(int64(lineAmount), "Creating Epub...")
	id := uuid.New()

	w := epub.New(fmt.Sprintf("urn:uuid:%s", id.String()))

	w.Title(credits.Title)
	w.Author(credits.Author)
	w.Languages("ja")

	var cssStyling string
	var vertHeader string

	if vertical {
		w.Direction("rtl")
		vertHeader = ` dir="rtl"`
		cssStyling = `
	<style>
		html {
			writing-mode: vertical-rl;
			-webkit-writing-mode: vertical-rl;
			-epub-writing-mode: vertical-rl;
			font-family: "serif", "Yu Mincho", "MS Mincho", serif;
		}
		p {
			text-indent: 1em;
			margin: 0;
			line-height: 1.8;
		}
		.tate-chu-yoko {
			-webkit-text-combine: horizontal;
			-epub-text-combine: horizontal;
			text-combine-upright: all;
		}
	</style>`
	} else {
		cssStyling = `
	<style>
		html {
			font-family: "serif", "Yu Mincho", "MS Mincho", serif;
		}
		p {
			text-indent: 1em;
			margin: 0;
			line-height: 1.8;
		}
	</style>`
	}

	w.MetaContent(map[string]string{
		"rendition:layout":      "reflowable",
		"rendition:orientation": "auto",
	})

	var tocItems []epub.TOC

	for i, chapter := range chapterData {
		currChapter := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://w3.org"%s xml:lang="ja" lang="ja">
<head>
    <title>%s</title>
%s
</head>
<body>
	<h1>%s</h1>
`, vertHeader, chapter.Title, cssStyling, chapter.Title)
		var xhtml strings.Builder
		xhtml.WriteString(currChapter)
		for _, line := range chapter.Contents {
			ePubProgress.Add(1)
			if line == "" {
				continue
			}
			currLine := fmt.Sprintf(`	<p>%s</p>
`, line)
			xhtml.WriteString(currLine)
		}

		endLines := `</body>
</html>`
		xhtml.WriteString(endLines)
		chapterName := fmt.Sprintf("Chapter-%d.xhtml", i)

		w.AddContent(chapterName, []byte(xhtml.String()))

		tocItems = append(tocItems, epub.TOC{
			Title: chapter.Title,
			Href:  chapterName,
		})
	}

	mainTOC := epub.TOC{
		Title: "目次",
		Items: tocItems,
	}

	err := w.TableOfContents("toc", mainTOC)
	if err != nil {
		return nil, err
	}

	fileNameExt := fmt.Sprintf("%s.epub", fileName)
	filePath := filepath.Join(destinationPath, fileNameExt)

	err = w.Write(filePath)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Successfully created epub at: %s\n", filePath)
	return []string{filePath}, nil
}
