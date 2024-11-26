package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CommunityEnterprise represents the structure for extracted data.
type CommunityEnterprise struct {
	EnterpriseName string `json:"ชื่อ"`
	BusinessGroup  string `json:"กลุ่มกิจการ"`
	BusinessType   string `json:"ประเภทกิจการ"`
	ProductName    string `json:"ชื่อผลิตภัณฑ์/บริการ"`
	ImageURL       string `json:"image_url"`
}

func main() {
	// Define the URL of the target page (for a single page)
	url := "https://smce2023.doae.go.th/ProductC_Result.php?page_size=5&PAGE=1&business_type_id=1&smce_id=&select_province=&select_region=&select_amphur=&key_word=&startPage=1&endPage=10"

	// Make the HTTP request
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch page, status code: %d", resp.StatusCode)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	var allEnterprises []CommunityEnterprise

	// Extract data for each product (single page)
	doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
		var enterprise CommunityEnterprise

		// Extract image URL
		image := row.Find("td img")
		if image != nil {
			imgSrc, exists := image.Attr("src")
			if exists {
				enterprise.ImageURL = "https://smce2023.doae.go.th/" + strings.TrimSpace(imgSrc) // Add base URL for image
			}
		}

		// Extract enterprise details
		row.Find("td").Each(func(j int, col *goquery.Selection) {
			col.Find(".box-product").Each(func(k int, detail *goquery.Selection) {
				field := strings.TrimSpace(detail.Find(".pro-field").Text())
				value := strings.TrimSpace(detail.Find(".pro-disc").Text())

				switch field {
				case "ชื่อ":
					enterprise.EnterpriseName = value
				case "กลุ่มกิจการ":
					enterprise.BusinessGroup = value
				case "ประเภทกิจการ":
					enterprise.BusinessType = value
				case "ชื่อผลิตภัณฑ์/บริการ":
					enterprise.ProductName = value
				}
			})
		})

		// Append the extracted data
		allEnterprises = append(allEnterprises, enterprise)
	})

	// Save the extracted data to JSON
	outputFile, err := os.Create("Output.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allEnterprises); err != nil {
		log.Fatalf("Failed to encode data to JSON: %v", err)
	}

	log.Println("Data extraction completed. Output saved to Output.json")
}
