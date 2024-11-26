package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// CommunityEnterprise represents the structure for extracted data.
type CommunityEnterprise struct {
	EnterpriseName  string `json:"enterprise_name"`   // เปลี่ยนเป็น snake_case
	BusinessGroup   string `json:"business_group"`    // เปลี่ยนเป็น snake_case
	BusinessType    string `json:"business_type"`     // เปลี่ยนเป็น snake_case
	ProductName     string `json:"product_name"`      // เปลี่ยนเป็น snake_case
	ImageURL        string `json:"image_url"`         // เปลี่ยนเป็น snake_case
}

func main() {
	baseURL := "https://smce2023.doae.go.th/ProductC_Result.php?page_size=5&PAGE=%d&business_type_id=1&smce_id=&select_province=&select_region=&select_amphur=&key_word=&startPage=1&endPage=20"

	headers := map[string]string{
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":           "en-US,en;q=0.9,th;q=0.8",
		"cache-control":             "no-cache",
		"cookie":                    "PHPSESSID=00qdjvaig0hbvtpksf7l2lhqj3",
		"pragma":                    "no-cache",
		"referer":                   "https://smce2023.doae.go.th/ProductC_Result.php?business_type_id=1&smce_id=&select_province=&select_region=&select_amphur=&key_word=",
		"sec-ch-ua":                 `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"Windows"`,
		"sec-fetch-dest":            "document",
		"sec-fetch-mode":            "navigate",
		"sec-fetch-site":            "same-origin",
		"sec-fetch-user":            "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}

	var allEnterprises []CommunityEnterprise

	// Loop through pages
	for page := 1; page <= 20; page++ {
		url := fmt.Sprintf(baseURL, page)
		log.Printf("Fetching page %d...\n", page)

		// Make the HTTP request
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Failed to create request: %v", err)
		}

		// Add headers to the request
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to fetch URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error fetching page %d: HTTP %d", page, resp.StatusCode)
			continue
		}

		// Convert from windows-874 to UTF-8
		reader := transform.NewReader(resp.Body, charmap.Windows874.NewDecoder())

		// Parse the HTML
		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			log.Fatalf("Failed to parse HTML: %v", err)
		}

		// Extract data from the page
		doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
			var enterprise CommunityEnterprise

			// Extract image URL
			row.Find("td img").Each(func(idx int, img *goquery.Selection) {
				imgSrc, exists := img.Attr("src")
				if exists {
					enterprise.ImageURL = "https://smce2023.doae.go.th/" + strings.TrimSpace(imgSrc)
				}
			})

			// Extract enterprise details
			row.Find(".box-product").Each(func(idx int, item *goquery.Selection) {
				field := strings.TrimSpace(item.Find(".pro-field").Text())
				value := strings.TrimSpace(item.Find(".pro-disc").Text())

				// Remove extra spaces and newline characters
				value = strings.Join(strings.Fields(value), " ")

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

			// Append data if at least the EnterpriseName or ProductName is not empty
			if enterprise.EnterpriseName != "" || enterprise.ProductName != "" {
				allEnterprises = append(allEnterprises, enterprise)
			}
		})

		time.Sleep(1 * time.Second) // Avoid overwhelming the server
	}

	// Save all results to JSON
	outputFile, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allEnterprises); err != nil {
		log.Fatalf("Failed to encode data to JSON: %v", err)
	}

	log.Println("Data extraction completed. Output saved to output.json")
}
