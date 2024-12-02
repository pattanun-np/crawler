package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Data struct for the fields you want to extract
type WebsiteData struct {
	No            string `json:"no"`
	WebsiteName   string `json:"website_name"`
	OperatorName  string `json:"operator_name"`
	DBDRegister   string `json:"dbd_register"`
	DBDVerified   string `json:"dbd_verified"`
}

func main() {
	// Open the webpage URL
	url := "https://trustmarkthai.com/en/search"
	// Make a request to fetch the page content
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}

	var data []WebsiteData

	// Locate the rows in the table containing the data you want
	doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
		// Extract the data for each field from the columns of the row
		no := strings.TrimSpace(row.Find("td").Eq(0).Text())
		websiteName := strings.TrimSpace(row.Find("td").Eq(1).Text())
		operatorName := strings.TrimSpace(row.Find("td").Eq(2).Text())
		dbdRegister := strings.TrimSpace(row.Find("td").Eq(3).Text())
		dbdVerified := strings.TrimSpace(row.Find("td").Eq(4).Text())

		// Append the data to the slice
		data = append(data, WebsiteData{
			No:            no,
			WebsiteName:   websiteName,
			OperatorName:  operatorName,
			DBDRegister:   dbdRegister,
			DBDVerified:   dbdVerified,
		})
	})

	// Create or open the output.json file
	outputFile, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Create a JSON encoder with pretty print
	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode data to JSON: %v", err)
	}

	fmt.Println("Data extraction completed. Output saved to output.json")
}
