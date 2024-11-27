package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

type Product struct {
	OrganizationName     string   `json:"organization_name"`
	RegistrationCode     string   `json:"registration_code"`
	Address              string   `json:"address"`
	Phone                string   `json:"phone"`
	Fax                  string   `json:"fax"`
	Representatives      []string `json:"representatives"`
	ProductName          string   `json:"product_name"`
	Properties           string   `json:"properties"`
	Composition          string   `json:"composition"`
	NutritionInfo        string   `json:"nutrition_info"`
	ProductionCapacity   string   `json:"production_capacity"`
	PricePerTon          string   `json:"price_per_ton"`
	Standards            string   `json:"standards"`
	QualityAssurance     string   `json:"quality_assurance"`
	ProductionPeriod     string   `json:"production_period"`
	SeasonalUse          string   `json:"seasonal_use"`
	DistributionChannels string   `json:"distribution_channels"`
}

func main() {
	// The URL to fetch
	url := "https://smce2023.doae.go.th/product_detail.php?smce_id=270020210002&ps_id=45"

	// Send HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	defer resp.Body.Close()

	// Check if status is OK
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: status code", resp.StatusCode)
		return
	}

	// Use charset.NewReader to handle the character encoding
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("Error creating UTF-8 reader:", err)
		return
	}

	// Parse the document
	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// Initialize the product struct
	product := Product{}

	// Extract organization details
	tables := doc.Find("table.table-striped.table-hover")
	if tables.Length() < 2 {
		fmt.Println("Error: Expected at least two tables")
		return
	}

	organizationTable := tables.Eq(0)
	organizationTable.Find("tr").Each(func(i int, s *goquery.Selection) {
		// Get the label and value from the <td> elements
		tds := s.Find("td")
		if tds.Length() >= 2 {
			label := strings.TrimSpace(tds.First().Text())
			value := strings.TrimSpace(tds.Last().Text())

			switch label {
			case "ชื่อ :":
				product.OrganizationName = value
			case "รหัสทะเบียน :":
				product.RegistrationCode = value
			case "ที่ตั้ง :":
				product.Address = value
			case "โทรศัพท์  :":
				product.Phone = value
			case "โทรสาร :":
				product.Fax = value
			case "ผู้มีอำนาจทำการแทน :":
				// Get HTML to preserve <br /> tags
				htmlContent, _ := tds.Last().Html()
				// Split by <br /> tags
				reps := strings.Split(htmlContent, "<br />")
				for _, rep := range reps {
					rep = strings.TrimSpace(stripTags(rep))
					if rep != "" {
						product.Representatives = append(product.Representatives, rep)
					}
				}
			}
		}
	})

	// Extract product name from h3
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if strings.HasPrefix(text, "ชื่อผลิตภัณฑ์/บริการ :") {
			parts := strings.SplitN(text, ":", 2)
			if len(parts) == 2 {
				product.ProductName = strings.TrimSpace(parts[1])
			}
		}
	})

	// Extract product details
	productTable := tables.Eq(1)
	productTable.Find("tr").Each(func(i int, s *goquery.Selection) {
		td := s.Find("td").First()
		text := strings.TrimSpace(td.Text())
		parts := strings.SplitN(text, ":", 2)
		if len(parts) == 2 {
			label := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch label {
			case "คุณสมบัติ":
				product.Properties = value
			case "องค์ประกอบ":
				product.Composition = value
			case "ข้อมูลโภชนาการ":
				product.NutritionInfo = value
			case "ความสามารถในการผลิต":
				// Handle potential line breaks
				value = strings.ReplaceAll(value, "\n", " ")
				product.ProductionCapacity = value
			case "ราคา ต่อ ตัน":
				product.PricePerTon = value
			case "มาตรฐาน":
				product.Standards = value
			case "การรับรองคุณภาพ":
				product.QualityAssurance = value
			case "ระยะเวลาการผลิต":
				product.ProductionPeriod = value
			case "เทศกาลที่ใช้":
				product.SeasonalUse = value
			case "ช่องทางการจัดจำหน่าย":
				product.DistributionChannels = value
			}
		}
	})

	// Write the product struct to JSON file
	outputFile, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(product); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println("Data extracted and saved to output.json successfully.")
}

// stripTags removes HTML tags from a string
func stripTags(html string) string {
	// Simple implementation to remove HTML tags
	return strings.TrimSpace(html)
}
