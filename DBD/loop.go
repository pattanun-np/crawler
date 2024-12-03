package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// BusinessInfo represents the structure for the extracted data
type BusinessInfo struct {
	No                  int    `json:"no"`
	OwnerName           string `json:"owner_name"`
	BusinessName        string `json:"business_name"`
	NationalID          string `json:"national_id"`
	OnlineStoreName     string `json:"online_store_name"`
	Platform            string `json:"platform"`
	BusinessTypeTH      string `json:"business_type_th"`
	BusinessTypeEN      string `json:"business_type_en"`
	AddressTH           string `json:"address_th"`
	AddressEN           string `json:"address_en"`
	TrustmarkStatus     string `json:"trustmark_status"`
	RegistrationDate    string `json:"registration_date"`
	DBDRegisteredDate   string `json:"dbd_registered_date"`
	DBDRenewalDate      string `json:"dbd_renewal_date"`
	DBDExpirationDate   string `json:"dbd_expiration_date"`
	RegisteredDateEN    string `json:"registered_date_en"`
	ExpirationDateEN    string `json:"expiration_date_en"`
}

// cleanField removes extra spaces and formats text properly
func cleanField(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

// fetchData retrieves the data from a specific URL and assigns No.
func fetchData(url string, no int) (BusinessInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return BusinessInfo{}, fmt.Errorf("ไม่สามารถเข้าถึงหน้าเว็บได้: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return BusinessInfo{}, fmt.Errorf("การเข้าถึงหน้าเว็บล้มเหลว: HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return BusinessInfo{}, fmt.Errorf("ไม่สามารถแปลง HTML เป็น Document: %v", err)
	}

	info := BusinessInfo{No: no}

	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		header := cleanField(s.Find("td.text-right").Text())
		value := cleanField(s.Find("td:not(.text-right)").Text())

		switch header {
		case "เลขประจำตัวประชาชน/เลขทะเบียนนิติบุคคล (Thai national Id/Juristic person Id) :":
			info.NationalID = value
		case "ชื่อผู้ประกอบการ :":
			info.OwnerName = value
		case "ชื่อที่ใช้ในการประกอบพาณิชยกิจ :":
			info.BusinessName = value
		case "ชื่อร้านค้าออนไลน์ (Online store) :":
			info.OnlineStoreName = value
			if strings.Contains(strings.ToLower(value), "facebook") {
				info.Platform = "Facebook"
			} else if strings.Contains(strings.ToLower(value), "shopee") {
				info.Platform = "Shopee"
			} else if strings.Contains(strings.ToLower(value), "lazada") {
				info.Platform = "Lazada"
			} else if strings.Contains(strings.ToLower(value), "line") {
				info.Platform = "LINE"
			} else if strings.Contains(strings.ToLower(value), "instagram") {
				info.Platform = "Instagram"
			} else if strings.Contains(strings.ToLower(value), "twitter") {
				info.Platform = "Twitter"
			} else if strings.Contains(strings.ToLower(value), "tiktok") {
				info.Platform = "TikTok"
			} else {
				info.Platform = "Website"
			}
		case "ประเภทธุรกิจ :":
			info.BusinessTypeTH = value
		case "(Type of business) :":
			info.BusinessTypeEN = value
		case "สถานที่ติดต่อได้ :":
			info.AddressTH = value
		case "(Address) :":
			info.AddressEN = value
		case "สถานะเครื่องหมาย :":
			info.TrustmarkStatus = value
		case "วันที่เริ่มต้นประกอบพาณิชยกิจ :":
			info.RegistrationDate = value
		case "ได้รับ DBD Registered ครั้งแรก :":
			info.DBDRegisteredDate = value
		case "วันที่ได้รับ DBD Registered / ต่ออายุ :":
			info.DBDRenewalDate = value
		case "วันที่หมดอายุ DBD Registered :":
			info.DBDExpirationDate = value
		case "Registered date :":
			info.RegisteredDateEN = value
		case "Expire date :":
			info.ExpirationDateEN = value
		}
	})

	return info, nil
}

func main() {
	// Base URL ของหน้าแรก
	baseURL := "https://trustmarkthai.com/th/search?page=%d"

	// Slice สำหรับเก็บผลลัพธ์ทั้งหมด
	var allData []BusinessInfo

	// กำหนดจำนวนหน้าที่จะดึงข้อมูล
	totalPages := 106150 // ปรับตามจำนวนหน้าที่ต้องการ

	// ลำดับรายการเริ่มต้น
	no := 1

	// วนลูปดึงข้อมูลจากแต่ละหน้า
	for page := 1; page <= totalPages; page++ {
		// สร้าง URL สำหรับแต่ละหน้า
		pageURL := fmt.Sprintf(baseURL, page)

		// ทำ HTTP GET request
		resp, err := http.Get(pageURL)
		if err != nil {
			log.Printf("ไม่สามารถเข้าถึงหน้าเว็บหน้า %d ได้: %v", page, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("การเข้าถึงหน้าเว็บล้มเหลวในหน้า %d: HTTP %d", page, resp.StatusCode)
			continue
		}

		// สร้าง goquery document จากเนื้อหา HTML
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("ไม่สามารถแปลง HTML เป็น Document ในหน้า %d: %v", page, err)
			continue
		}

		// Slice สำหรับเก็บลิงก์ที่ตรงตามเงื่อนไข
		var dataURLs []string

		// ดึงลิงก์ที่มี href="https://trustmarkthai.com/callbackData/popup.php?data="
		doc.Find("a[href^='https://trustmarkthai.com/callbackData/popup.php?data=']").Each(func(i int, s *goquery.Selection) {
			link, exists := s.Attr("href")
			if exists {
				dataURLs = append(dataURLs, link)
			}
		})

		log.Printf("หน้า %d พบลิงก์ทั้งหมด %d รายการ", page, len(dataURLs))

		// วนลูปดึงข้อมูลจากแต่ละลิงก์
		for _, url := range dataURLs {
			info, err := fetchData(url, no)
			if err != nil {
				log.Printf("เกิดข้อผิดพลาดในการดึงข้อมูลจาก %s: %v", url, err)
				continue
			}
			allData = append(allData, info)
			no++
		}
	}

	// บันทึกผลลัพธ์ทั้งหมดเป็น JSON
	jsonData, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		log.Fatalf("ไม่สามารถแปลงข้อมูลทั้งหมดเป็น JSON: %v", err)
	}

	err = os.WriteFile("output.json", jsonData, 0644)
	if err != nil {
		log.Fatalf("ไม่สามารถบันทึก JSON ลงไฟล์: %v", err)
	}

	log.Println("บันทึกข้อมูลทั้งหมดลงในไฟล์ output.json สำเร็จ")
}
