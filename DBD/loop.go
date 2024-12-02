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
)

type WebsiteData struct {
    No           string `json:"no"`
    WebsiteName  string `json:"website_name"`
    OperatorName string `json:"operator_name"`
    DBDRegister  string `json:"dbd_register"`
    DBDVerified  string `json:"dbd_verified"`
}

func main() {
    baseURL := "https://trustmarkthai.com/th/search?page=%d"

    headers := map[string]string{
        "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "th-TH,th;q=0.9",
        "User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
    }

    var allData []WebsiteData

    for page := 1; page <= 1000; page++ {
        url := fmt.Sprintf(baseURL, page)
        log.Printf("กำลังดึงข้อมูลจากหน้า %d...\n", page)

        client := &http.Client{
            Timeout: 10 * time.Second,
        }
        req, err := http.NewRequest("GET", url, nil)
        if err != nil {
            log.Fatalf("ไม่สามารถสร้างคำขอได้: %v", err)
        }

        for key, value := range headers {
            req.Header.Set(key, value)
        }

        resp, err := client.Do(req)
        if err != nil {
            log.Fatalf("ไม่สามารถดึงข้อมูลจาก URL ได้: %v", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            log.Printf("เกิดข้อผิดพลาดในการดึงข้อมูลจากหน้า %d: HTTP %d", page, resp.StatusCode)
            continue
        }

        // ตรวจสอบการเข้ารหัสของเนื้อหา
        contentType := resp.Header.Get("Content-Type")
        fmt.Println("Content-Type:", contentType)
        // ตัวอย่าง: Content-Type: text/html; charset=UTF-8

        // แยกวิเคราะห์ HTML โดยตรงโดยไม่ต้องแปลงการเข้ารหัส
        doc, err := goquery.NewDocumentFromReader(resp.Body)
        if err != nil {
            log.Fatalf("ไม่สามารถแยกวิเคราะห์ HTML ได้: %v", err)
        }

        doc.Find("table.table tbody tr").Each(func(i int, row *goquery.Selection) {
            var (
                no           string
                websiteName  string
                operatorName string
                dbdRegister  string
                dbdVerified  string
            )

            row.Find("td").Each(func(j int, col *goquery.Selection) {
                text := strings.TrimSpace(col.Text())

                switch j {
                case 0: // No.
                    no = text
                case 1: // Website Name
                    websiteName = text
                case 2: // Operator Name
                    operatorName = text
                case 3: // DBD Register
                    imgTag := col.Find("img")
                    if imgTag.Length() > 0 {
                        imgSrc, exists := imgTag.Attr("src")
                        if exists {
                            dbdRegister = imgSrc
                        }
                    } else {
                        dbdRegister = text
                    }
                case 4: // DBD Verified
                    dbdVerified = text
                }
            })

            allData = append(allData, WebsiteData{
                No:           no,
                WebsiteName:  websiteName,
                OperatorName: operatorName,
                DBDRegister:  dbdRegister,
                DBDVerified:  dbdVerified,
            })
        })

        time.Sleep(1 * time.Second)
    }

    outputFile, err := os.Create("output.json")
    if err != nil {
        log.Fatalf("ไม่สามารถสร้างไฟล์ผลลัพธ์ได้: %v", err)
    }
    defer outputFile.Close()

    encoder := json.NewEncoder(outputFile)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(allData); err != nil {
        log.Fatalf("ไม่สามารถเข้ารหัสข้อมูลเป็น JSON ได้: %v", err)
    }

    log.Println("การดึงข้อมูลเสร็จสมบูรณ์ ผลลัพธ์ถูกบันทึกลงใน output.json")
}
