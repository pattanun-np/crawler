package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Enterprise represents the structure for extracted data
type Enterprise struct {
	Serial           string  `json:"serial,omitempty"`        // เพิ่มฟิลด์ Serial
	OrganizationName string  `json:"organization_name,omitempty"`
	RegistrationCode string  `json:"registration_code,omitempty"`
	Address          string  `json:"address,omitempty"`
	Phone            string  `json:"phone,omitempty"`
	Representatives  string  `json:"representatives,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`  // เพิ่มฟิลด์ Latitude
	Longitude        float64 `json:"longitude,omitempty"` // เพิ่มฟิลด์ Longitude
}

func main() {
	// ตั้งค่า page size และ จำนวนหน้า
	pageSize := 10       // จำนวนข้อมูลต่อหน้า
	totalPages := 50     // จำนวนหน้าที่ต้องการดึงข้อมูล
	var allEnterprises []Enterprise

	// ดึงข้อมูลจากทุกหน้า
	for pageNumber := 1; pageNumber <= totalPages; pageNumber++ {
		// URL สำหรับดึงข้อมูลจากแต่ละหน้า
		url := fmt.Sprintf("https://smce2023.doae.go.th/ProductCategory/SmceCategory.php?page_size=%d&PAGE=%d&province_id=&region_id=&amphur_id=&key_word=&startPage=%d&endPage=%d", pageSize, pageNumber, pageSize*(pageNumber-1)+1, pageSize*pageNumber)

		// ดึงข้อมูลจากหน้าแรกที่มีการจัด Pagination
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to fetch URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Failed to fetch page: HTTP %d", resp.StatusCode)
		}

		// แปลงจาก windows-874 เป็น UTF-8
		reader := transform.NewReader(resp.Body, charmap.Windows874.NewDecoder())

		// แปลง HTML
		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			log.Fatalf("Failed to parse HTML: %v", err)
		}

		// ดึงข้อมูลจากแต่ละแถวในตาราง
		doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
			// ดึง serial จากคอลัมน์แรก
			serial := extractSerial(row)

			// ค้นหาลิงก์ที่มี smce_id
			row.Find("td a").Each(func(j int, link *goquery.Selection) {
				smceID, exists := link.Attr("href")
				if exists && strings.Contains(smceID, "smce_id=") {
					// ดึง smce_id จาก href
					smceID = extractSmceID(smceID)

					// Log ข้อมูลหน้า
					log.Printf("Fetching page %d, smce_id: %s, serial: %s", pageNumber, smceID, serial)

					// ดึงข้อมูลจาก smce_id นี้
					fetchEnterpriseData(smceID, serial, &allEnterprises)
				}
			})
		})

		// แสดงผลใน terminal ว่ากำลังดึงข้อมูลจากหน้าไหน
		log.Printf("Fetching page %d...", pageNumber)
	}

	// เขียนข้อมูลที่ดึงได้ไปยังไฟล์ JSON
	outputFile, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allEnterprises); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println("Data extraction completed. Output saved to output.json")
}

// Function to extract smce_id from the href attribute
func extractSmceID(href string) string {
	parts := strings.Split(href, "=")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// Function to extract serial number from the row
func extractSerial(row *goquery.Selection) string {
	// ดึงข้อมูลจากคอลัมน์แรก (ซึ่งมักจะเป็น serial)
	serial := row.Find("td").Eq(0).Text()
	return strings.TrimSpace(serial)
}

// Fetch data for each smce_id
func fetchEnterpriseData(smceID string, serial string, allEnterprises *[]Enterprise) {
	enterpriseURL := fmt.Sprintf("https://smce2023.doae.go.th/ProductCategory/managecontent.php?smce_id=%s", smceID)

	// ดึงข้อมูลจากหน้าเดียว
	resp, err := http.Get(enterpriseURL)
	if err != nil {
		log.Printf("Error fetching URL: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching page: HTTP %d", resp.StatusCode)
		return
	}

	// แปลงเอกสาร HTML จาก charset
	utf8Reader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("Error creating UTF-8 reader: %v", err)
		return
	}

	// แปลง HTML
	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		return
	}

	// ค้นหาตารางในเอกสาร HTML
	tables := doc.Find("table.table-striped.table-hover")
	if tables.Length() < 1 {
		log.Printf("Error: Expected at least one table")
		return
	}

	// ดึงข้อมูลจากแถวของตาราง
	organizationTable := tables.Eq(0)
	enterprise := Enterprise{
		Serial: serial, // ตั้งค่า serial
	} // สร้าง object enterprise ใหม่

	organizationTable.Find("tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() >= 2 {
			label := strings.TrimSpace(tds.First().Text())
			value := strings.TrimSpace(tds.Last().Text())

			// แมปข้อมูลจากตารางไปยัง fields ใน struct
			switch label {
			case "ชื่อ :":
				enterprise.OrganizationName = value
			case "รหัสทะเบียน :":
				enterprise.RegistrationCode = value
			case "ที่อยู่ :":
				enterprise.Address = value
			case "โทรศัพท์  :":
				enterprise.Phone = value
			case "ผู้มีอำนาจทำการแทน :":
				// การดึง HTML content ของผู้มีอำนาจทำการแทน
				htmlContent, _ := tds.Last().Html()

				// กำจัด <br /> และ \u003cbr/\u003e ออกจากข้อความ
				cleanContent := strings.ReplaceAll(htmlContent, "<br />", " ")  // แทนที่ <br /> ด้วยช่องว่าง
				cleanContent = strings.ReplaceAll(cleanContent, "\u003cbr/\u003e", " ") // แทนที่ \u003cbr/\u003e ด้วยช่องว่าง

				// ลบ HTML tags และช่องว่างที่เกินออก
				cleanedText := stripTags(cleanContent)

				// กำหนดค่าให้กับ Representatives
				enterprise.Representatives = strings.TrimSpace(cleanedText)  // แสดงข้อมูลโดยตรง
			}
		}
	})

	// ค้นหาพิกัดจาก URL ของ Google Maps ในส่วนของที่อยู่
	address := enterprise.Address
	latitude, longitude := extractCoordinates(address)

	// เพิ่มพิกัดลงใน enterprise
	enterprise.Latitude = latitude
	enterprise.Longitude = longitude

	// เพิ่มข้อมูลที่ดึงได้ลงใน allEnterprises
	*allEnterprises = append(*allEnterprises, enterprise)
}

// stripTags removes HTML tags from a string
func stripTags(html string) string {
	// ลบ HTML tags ออก
	return strings.TrimSpace(html)
}

// Function to extract coordinates (latitude, longitude) from Google Maps URL
func extractCoordinates(address string) (float64, float64) {
	// Regular expression to match Google Maps link and extract latitude and longitude
	re := regexp.MustCompile(`https://maps\.google\.com/maps\?q=([0-9.-]+),([0-9.-]+)`)
	matches := re.FindStringSubmatch(address)

	if len(matches) == 3 {
		latitude := matches[1]
		longitude := matches[2]

		// Convert latitude and longitude to float64
		lat, _ := strconv.ParseFloat(latitude, 64)
		lng, _ := strconv.ParseFloat(longitude, 64)

		return lat, lng
	}

	// Default if coordinates are not found
	return 0.0, 0.0
}
