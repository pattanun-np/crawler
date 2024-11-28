package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type CommunityEnterprise struct {
	EnterpriseName       string `json:"enterprise_name"`
	BusinessGroup        string `json:"business_group"`
	BusinessType         string `json:"business_type"`
	ProductName          string `json:"product_name"`
	ImageURL             string `json:"image_url"`
	SMCEID               string `json:"smce_id"`
	PSID                 string `json:"ps_id"`
	RegistrationCode     string `json:"registration_code"`
	Address              string `json:"address"`
	Phone                string `json:"phone"`
	Fax                  string `json:"fax"`
	Properties           string `json:"properties"`
	Composition          string `json:"composition"`
	NutritionInfo        string `json:"nutrition_info"`
	ProductionPeriod     string `json:"production_period"`
	ProductionCapacity   string `json:"production_capacity"`
	PricePerTon          string `json:"price_per_ton"`
	Standards            string `json:"standards"`
	QualityAssurance     string `json:"quality_assurance"`
	SeasonalUse          string `json:"seasonal_use"`
	DistributionChannels string `json:"distribution_channels"`
	Latitude             string `json:"latitude"`
	Longitude            string `json:"longitude"`
}

func main() {
	var allData []CommunityEnterprise

	// Fetch community enterprises
	fetchCommunityEnterprises(&allData)

	// Save to output.json
	saveToJSON(allData)
}

func fetchCommunityEnterprises(allData *[]CommunityEnterprise) {
	baseURL := "https://smce2023.doae.go.th/ProductC_Result.php?page_size=5&PAGE=%d&business_type_id=1&smce_id=&select_province=&select_region=&select_amphur=&key_word=&startPage=1&endPage=20"
	headers := map[string]string{
		"accept":        "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"user-agent":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"cookie":        "PHPSESSID=00qdjvaig0hbvtpksf7l2lhqj3",
		"referer":       "https://smce2023.doae.go.th/ProductC_Result.php?business_type_id=1&smce_id=&select_province=&select_region=&select_amphur=&key_word=",
	}

	// Loop to fetch multiple pages
	for page := 1; page <= 5; page++ {
		url := fmt.Sprintf(baseURL, page)
		// Fetching community enterprise page
		log.Printf("Fetching community enterprise page %d...\n", page)

		// Make HTTP request
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Failed to create request: %v", err)
		}
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

		// Convert encoding and parse HTML
		reader := transform.NewReader(resp.Body, charmap.Windows874.NewDecoder()) // TIS-620 -> UTF-8
		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			log.Fatalf("Failed to parse HTML: %v", err)
		}

		// Extract community enterprise data
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

				// Clean up value and store the information
				value = strings.Join(strings.Fields(value), " ")

				// Handle each field
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

			// Extract smce_id and ps_id
			row.Find("a").Each(func(idx int, a *goquery.Selection) {
				href, exists := a.Attr("href")
				if exists && strings.Contains(href, "product_detail.php") {
					// Extract smce_id and ps_id from URL
					re := regexp.MustCompile(`smce_id=(\d+)&ps_id=(\d+)`)
					matches := re.FindStringSubmatch(href)
					if len(matches) == 3 {
						enterprise.SMCEID = matches[1]
						enterprise.PSID = matches[2]
					}
				}
			})

			// Fetch product details using smce_id and ps_id
			if enterprise.SMCEID != "" && enterprise.PSID != "" {
				fetchProductDetails(&enterprise)
			}

			// Add to allData array
			*allData = append(*allData, enterprise)
		})

		time.Sleep(1 * time.Second)
	}
}

func fetchProductDetails(enterprise *CommunityEnterprise) {
	// Build URL for product details page
	baseURL := "https://smce2023.doae.go.th/product_detail.php"
	params := url.Values{}
	params.Add("smce_id", enterprise.SMCEID)
	params.Add("ps_id", enterprise.PSID)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Printf("Error creating request for product details: %v", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching product details: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching product details: HTTP %d", resp.StatusCode)
		return
	}

	// Parse the product details page (convert from Windows-874 encoding to UTF-8)
	reader := transform.NewReader(resp.Body, charmap.Windows874.NewDecoder()) // TIS-620 -> UTF-8
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Printf("Error parsing product details HTML: %v", err)
		return
	}

	// Extract additional product details
	doc.Find("table tr").Each(func(i int, row *goquery.Selection) {
		// Extract only the value without the prefix (e.g., "รหัสผลิตภัณฑ์")
		value := extractDataFromRow(row)

		// Store the value in the corresponding field
		switch {
		case strings.Contains(value, "รหัสผลิตภัณฑ์"):
			enterprise.PSID = cleanField(value)
		case strings.Contains(value, "รหัสสินค้า"):
			enterprise.SMCEID = cleanField(value)
		case strings.Contains(value, "รหัสการจดทะเบียน"):
			enterprise.RegistrationCode = cleanField(value)
		case strings.Contains(value, "ที่อยู่"):
			enterprise.Address = cleanField(value)
		case strings.Contains(value, "โทรศัพท์"):
			enterprise.Phone = cleanField(value)
		case strings.Contains(value, "โทรสาร"):
			enterprise.Fax = cleanField(value)
		case strings.Contains(value, "คุณสมบัติ"):
			enterprise.Properties = cleanField(value)
		case strings.Contains(value, "ส่วนประกอบ"):
			enterprise.Composition = cleanField(value)
		case strings.Contains(value, "ข้อมูลทางโภชนาการ"):
			enterprise.NutritionInfo = cleanField(value)
		case strings.Contains(value, "ระยะเวลาในการผลิต"):
			enterprise.ProductionPeriod = cleanField(value)
		case strings.Contains(value, "กำลังการผลิต"):
			enterprise.ProductionCapacity = cleanField(value)
		case strings.Contains(value, "ราคาต่อตัน"):
			enterprise.PricePerTon = cleanField(value)
		case strings.Contains(value, "มาตรฐาน"):
			enterprise.Standards = cleanField(value)
		case strings.Contains(value, "การรับรองคุณภาพ"):
			enterprise.QualityAssurance = cleanField(value)
		case strings.Contains(value, "ฤดูกาลที่ใช้"):
			enterprise.SeasonalUse = cleanField(value)
		case strings.Contains(value, "ช่องทางการจัดจำหน่าย"):
			enterprise.DistributionChannels = cleanField(value)
		case strings.Contains(value, "พิกัดละติจูด"):
			enterprise.Latitude = cleanField(value)
		case strings.Contains(value, "พิกัดลองจิจูด"):
			enterprise.Longitude = cleanField(value)
		}
	})
}

func extractDataFromRow(row *goquery.Selection) string {
	// Clean and return the value from the row
	value := strings.TrimSpace(row.Text())
	return strings.Join(strings.Fields(value), " ")
}

func cleanField(value string) string {
	// List of prefixes that we want to remove
	prefixes := []string{
		"โทรศัพท์ :", "โทรสาร :", "คุณสมบัติ :", "พิกัดละติจูด :", "พิกัดลองจิจูด :",
		"รหัสผลิตภัณฑ์ :", "รหัสสินค้า :", "รหัสการจดทะเบียน :", "ที่อยู่ :", "ส่วนประกอบ :",
		"ข้อมูลทางโภชนาการ :", "ระยะเวลาในการผลิต :", "กำลังการผลิต :", "ราคาต่อตัน :", 
		"มาตรฐาน :", "การรับรองคุณภาพ :", "ฤดูกาลที่ใช้ :", "ช่องทางการจัดจำหน่าย :",
	}

	// Iterate through all prefixes and remove them if they match
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			// If the value starts with the prefix, remove it but keep what's after ":"
			value = strings.TrimSpace(strings.TrimPrefix(value, prefix))
			break
		}
	}

	// Return cleaned value
	return value
}

func saveToJSON(data []CommunityEnterprise) {
	// Open file to write the JSON data
	file, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("Error creating JSON file: %v", err)
	}
	defer file.Close()

	// Marshal the data to JSON format
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Error encoding JSON data: %v", err)
	}

	log.Println("Data saved to output.json")
}
