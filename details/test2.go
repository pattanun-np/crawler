package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

type Product struct {
	SMCEID               string   `json:"smce_id"`
	PSID                 string   `json:"ps_id"`
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
	Latitude             string   `json:"latitude"`
	Longitude            string   `json:"longitude"`
}

func main() {
	// Define the range of IDs to iterate over
	// Adjust the ranges according to your needs
	smceIDStart := 270020210001
	smceIDEnd := 2700202100100 // Adjust the end ID as needed

	psIDStart := 1
	psIDEnd := 100 // Adjust the end ID as needed

	// Create or open the output file
	outputFile, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Write the opening bracket for the JSON array
	outputFile.WriteString("[\n")

	firstRecord := true

	// Loop over the IDs
	for smceID := smceIDStart; smceID <= smceIDEnd; smceID++ {
		for psID := psIDStart; psID <= psIDEnd; psID++ {
			smceIDStr := strconv.Itoa(smceID)
			psIDStr := strconv.Itoa(psID)

			// Fetch and parse the product
			product, err := fetchAndParseProduct(smceIDStr, psIDStr)
			if err != nil {
				fmt.Printf("ID smce_id=%s, ps_id=%s: %v\n", smceIDStr, psIDStr, err)
				continue
			}

			// Assign IDs to the product struct
			product.SMCEID = smceIDStr
			product.PSID = psIDStr

			// Encode the product to JSON
			productJSON, err := json.MarshalIndent(product, "  ", "  ")
			if err != nil {
				fmt.Printf("Error encoding product ID smce_id=%s, ps_id=%s: %v\n", smceIDStr, psIDStr, err)
				continue
			}

			// Write comma if not the first record
			if !firstRecord {
				outputFile.WriteString(",\n")
			} else {
				firstRecord = false
			}

			// Write the product JSON to the file
			outputFile.WriteString("  ")
			outputFile.Write(productJSON)

			fmt.Printf("Saved product ID smce_id=%s, ps_id=%s\n", smceIDStr, psIDStr)

			// Optional: Sleep between requests to be polite
			time.Sleep(200 * time.Millisecond)
		}
	}

	// Write the closing bracket for the JSON array
	outputFile.WriteString("\n]\n")

	fmt.Println("Data extracted and saved to output.json successfully.")
}

func fetchAndParseProduct(smceID, psID string) (Product, error) {
	var product Product

	// Build the URL with query parameters
	baseURL := "https://smce2023.doae.go.th/product_detail.php"
	params := url.Values{}
	params.Add("smce_id", smceID)
	params.Add("ps_id", psID)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create a new HTTP request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return product, fmt.Errorf("Error creating request: %v", err)
	}

	// Set necessary headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,th;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	// Send HTTP GET request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return product, fmt.Errorf("Error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	// Check if status is OK
	if resp.StatusCode != http.StatusOK {
		return product, fmt.Errorf("Status code %d", resp.StatusCode)
	}

	// Use charset.NewReader to handle character encoding
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return product, fmt.Errorf("Error creating UTF-8 reader: %v", err)
	}

	// Limit the reader to prevent large responses
	limitedReader := io.LimitReader(utf8Reader, 10*1024*1024) // 10 MB limit

	// Parse the document
	doc, err := goquery.NewDocumentFromReader(limitedReader)
	if err != nil {
		return product, fmt.Errorf("Error parsing HTML: %v", err)
	}

	// Check for a specific element to verify if the product exists
	if doc.Find("div.col-md-12").Length() == 0 {
		return product, fmt.Errorf("Product not found")
	}

	// Extract organization details
	tables := doc.Find("table.table-striped.table-hover")
	if tables.Length() < 2 {
		return product, fmt.Errorf("Expected at least two tables")
	}

	organizationTable := tables.Eq(0)
	organizationTable.Find("tr").Each(func(i int, s *goquery.Selection) {
		// Get the label and value from the <td> elements
		tds := s.Find("td")
		if tds.Length() >= 2 {
			label := cleanText(tds.First().Text())
			value := cleanText(tds.Last().Text())

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
				// Decode HTML entities
				htmlContent = html.UnescapeString(htmlContent)
				// Remove HTML tags
				htmlContent = stripTags(htmlContent)
				// Split by line breaks or numbers
				reps := splitRepresentatives(htmlContent)
				for _, rep := range reps {
					rep = cleanText(rep)
					if rep != "" {
						product.Representatives = append(product.Representatives, rep)
					}
				}
			}
		}
	})

	// Extract product name from h3
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if strings.HasPrefix(text, "ชื่อผลิตภัณฑ์/บริการ :") {
			parts := strings.SplitN(text, ":", 2)
			if len(parts) == 2 {
				product.ProductName = cleanText(parts[1])
			}
		}
	})

	// Extract product details
	productTable := tables.Eq(1)
	productTable.Find("tr").Each(func(i int, s *goquery.Selection) {
		td := s.Find("td").First()
		text := cleanText(td.Text())
		parts := strings.SplitN(text, ":", 2)
		if len(parts) == 2 {
			label := cleanText(parts[0])
			value := cleanText(parts[1])

			switch label {
			case "คุณสมบัติ":
				product.Properties = value
			case "องค์ประกอบ":
				product.Composition = value
			case "ข้อมูลโภชนาการ":
				product.NutritionInfo = value
			case "ความสามารถในการผลิต":
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

	// Extract the latitude and longitude
	extractCoordinates(doc, &product)

	return product, nil
}

// Function to extract coordinates from the document
func extractCoordinates(doc *goquery.Document, product *Product) {
	// Find the iframe containing the map
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && strings.Contains(src, "maps.google.com") {
			lat, lng := parseCoordinatesFromURL(src)
			if lat != "" && lng != "" {
				product.Latitude = lat
				product.Longitude = lng
			}
		}
	})
}

// Function to parse coordinates from the iframe src URL
func parseCoordinatesFromURL(urlStr string) (string, string) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", ""
	}

	// The URL might contain coordinates in the "q" parameter
	// For example: https://maps.google.com/maps?q=13.7563,100.5018&...
	queryParams := u.Query()
	if q := queryParams.Get("q"); q != "" {
		coords := strings.Split(q, ",")
		if len(coords) == 2 {
			lat := strings.TrimSpace(coords[0])
			lng := strings.TrimSpace(coords[1])
			return lat, lng
		}
	}

	// Alternatively, extract from the path using regex
	path := u.Path
	re := regexp.MustCompile(`@([\d.-]+),([\d.-]+)`)
	matches := re.FindStringSubmatch(path)
	if len(matches) == 3 {
		lat := matches[1]
		lng := matches[2]
		return lat, lng
	}

	// If coordinates are embedded in the "pb" parameter
	if pb := queryParams.Get("pb"); pb != "" {
		re := regexp.MustCompile(`!1m14.*!3d([\d.-]+)!4d([\d.-]+)`)
		matches := re.FindStringSubmatch(pb)
		if len(matches) == 3 {
			lat := matches[1]
			lng := matches[2]
			return lat, lng
		}
	}

	return "", ""
}

// stripTags removes HTML tags from a string
func stripTags(htmlStr string) string {
	// Use regex to remove HTML tags
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(htmlStr, "")
}

// cleanText trims spaces, replaces multiple spaces and newlines
func cleanText(text string) string {
	text = strings.TrimSpace(text)
	// Replace multiple spaces with a single space
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")
	return text
}

// splitRepresentatives splits the representatives field into individual names
func splitRepresentatives(text string) []string {
	// Split by numbers followed by periods (e.g., "1.", "2.", "3.")
	re := regexp.MustCompile(`\d+\.\s*`)
	// Find all matches
	indices := re.FindAllStringIndex(text, -1)
	var reps []string
	for i := 0; i < len(indices); i++ {
		start := indices[i][1]
		var end int
		if i+1 < len(indices) {
			end = indices[i+1][0]
		} else {
			end = len(text)
		}
		rep := text[start:end]
		rep = strings.TrimSpace(rep)
		reps = append(reps, rep)
	}
	return reps
}
