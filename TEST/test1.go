package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CommunityEnterprise represents the structure for extracted data.
type CommunityEnterprise struct {
	Serial        int    `json:"serial"`
	Registration  string `json:"registration"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
}

func main() {
	// Open the HTML file or URL (replace with URL if fetching online)
	file, err := os.Open("input.html") // Replace with a URL if fetching online
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	var enterprises []CommunityEnterprise

	// Locate the table and process its rows
	doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
		cols := row.Find("td")

		// Extract data from each column
		registration := strings.TrimSpace(cols.Eq(1).Text())
		name := strings.TrimSpace(cols.Eq(2).Find("a").Text())
		address := strings.TrimSpace(cols.Eq(3).Text())

		// Extract phone number from the address
		phone := ""
		if strings.Contains(address, "โทรศัพท์") {
			parts := strings.Split(address, "โทรศัพท์")
			if len(parts) > 1 {
				phone = strings.TrimSpace(parts[1])
				address = strings.TrimSpace(parts[0])
			}
		}

		// Assign serial based on the loop index (i + 1)
		serialInt := i + 1

		// Append to the results
		enterprises = append(enterprises, CommunityEnterprise{
			Serial:       serialInt,
			Registration: registration,
			Name:         name,
			Address:      address,
			Phone:        phone,
		})
	})

	// Save results to JSON
	outputFile, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(enterprises); err != nil {
		log.Fatalf("Failed to encode data to JSON: %v", err)
	}

	log.Println("Data extraction completed. Output saved to output.json")
}
