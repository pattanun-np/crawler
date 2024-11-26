package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// CommunityEnterprise represents the structure for extracted data.
type CommunityEnterprise struct {
	Serial       int    `json:"serial"`
	Registration string `json:"registration"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
}

func main() {
	baseURL := "https://smce2023.doae.go.th/ProductCategory/SmceCategory.php?page_size=10&PAGE=%d&province_id=&region_id=&amphur_id=&key_word=&startPage=1&endPage=10"

	headers := map[string]string{
		"accept":                  "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"accept-language":         "en-US,en;q=0.9,th;q=0.8",
		"cache-control":           "no-cache",
		"cookie":                  "PHPSESSID=00qdjvaig0hbvtpksf7l2lhqj3",
		"pragma":                  "no-cache",
		"referer":                 "https://smce2023.doae.go.th/ProductCategory/SmceCategory.php?region_id=&province_id=&amphur_id=&key_word=",
		"sec-ch-ua":               `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":        "?0",
		"sec-ch-ua-platform":      `"Windows"`, // Change to Windows instead of macOS
		"sec-fetch-dest":          "document",
		"sec-fetch-mode":          "navigate",
		"sec-fetch-site":          "same-origin",
		"sec-fetch-user":          "?1",
		"upgrade-insecure-requests": "1",
		"user-agent":              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	}

	var allEnterprises []CommunityEnterprise

	// Loop through pages
	for page := 1; page <= 10; page++ {
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
			var (
				serial       int
				registration string
				name         string
				address      string
				phone        string
			)

			row.Find("td").Each(func(j int, col *goquery.Selection) {
				text := strings.TrimSpace(col.Text())

				switch j {
				case 0: // Serial
					serialInt, err := strconv.Atoi(text)
					if err != nil {
						log.Printf("Failed to parse serial number: %v", err)
						serialInt = 0
					}
					serial = serialInt

				case 1: // Registration
					registration = text

				case 2: // Name
					name = strings.TrimSpace(col.Find("a").Text())
					if name == "" {
						name = text
					}

				case 3: // Address
					address = text

					// Extract phone number if present
					if strings.Contains(address, "โทรศัพท์") {
						parts := strings.Split(address, "โทรศัพท์")
						if len(parts) > 1 {
							phone = strings.TrimSpace(parts[1])
							address = strings.TrimSpace(parts[0])
						}
					}
				}
			})

			allEnterprises = append(allEnterprises, CommunityEnterprise{
				Serial:       serial,
				Registration: registration,
				Name:         name,
				Address:      address,
				Phone:        phone,
			})
		})

		time.Sleep(1 * time.Second) // Avoid overwhelming the server
	}

	// Save all results to JSON
	outputFile, err := os.Create("output.json") // Import os package for file creation
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
