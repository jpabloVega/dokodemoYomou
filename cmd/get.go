package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jpabloVega/dokodemoYomou/api"
	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
)

const (
	LINEHEIGHT    = 20.0
	MAXPAGEHEIGHT = 750.0
)

var bodySize int
var titleSize int
var chapterAmount int
var destinationPath string
var fileName string

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntVarP(&bodySize, "bodysize", "b", 14, "Size for the font of the body")
	getCmd.Flags().IntVarP(&titleSize, "titlesize", "t", 16, "Size for the font of the titles")
	getCmd.Flags().IntVarP(&chapterAmount, "volumesize", "v", 3, "Amount of chapter to download")
	getCmd.Flags().StringVarP(&destinationPath, "destination", "d", "pfd", "Path to the folder where pdf's are stored")
	getCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name of the newly created pdf")

}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the urls",
	Long:  "Implement me",
	Run:   GetContents,
}

func GetContents(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("No arguments given")
		return
	}
	c := api.NewClient(5 * time.Second)

	for _, address := range args {
		var urlData api.UrlInfo
		urlData, err := api.GetSiteInfo(address)
		if err != nil {
			fmt.Println(err)
			return
		}
		var pdfAddress string
		var chaptersData []api.Chapter
		for i := urlData.StartIndex; i < urlData.StartIndex+chapterAmount; i++ {
			fullURL := fmt.Sprintf("%s/%d/", urlData.Url, i)
			req, err := http.NewRequest("GET", fullURL, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			req.Header.Set("User-Agent", "export-novels-to-eReader-app")
			res, err := c.HttpClient.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				fmt.Printf("Error reaching link: %v\n", res.Status)
				return
			}
			data, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			chapterData, err := api.GetChapterData(string(data))
			if err != nil {
				fmt.Println(err)
				return
			}
			chaptersData = append(chaptersData, chapterData)
		}
		bookCredits, err := c.GetChapterAuthor(urlData.Url)
		if err != nil {
			fmt.Println(err)
			return
		}
		pdfAddress, err = CreatePDF(chaptersData, bookCredits, fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("PDF created successfully at: %v", pdfAddress)
	}

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

	if fileName == "" {
		fileName = credits.Title
	}
	err = pdf.WritePdf(destinationPath + "/" + fileName + ".pdf")
	if err != nil {
		return "", err
	}
	return destinationPath, nil
}
