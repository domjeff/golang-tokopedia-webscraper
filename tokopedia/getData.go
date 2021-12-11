package tokopedia

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"go-scrapping/model"

	"github.com/PuerkitoBio/goquery"
)

func setHeaders(req *http.Request) {
	req.Header.Set(
		"User-Agent",
		// "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Mobile Safari/537.36",
		"Chrome/96.0.4664.45",
		// "Chrome",
	)

	// return req
}

func setDetailedHeaders(req *http.Request, detailedUrl string, referer string) {
	req.Header.Set(
		"authority",
		"ta.tokopedia.com",
	)
	req.Header.Set(
		"method",
		"GET",
	)
	req.Header.Set(
		"path",
		strings.Replace(detailedUrl, "https://ta.tokopedia.com", "", -1),
	)
	req.Header.Set(
		"scheme",
		"https",
	)
	req.Header.Set(
		"User-Agent",
		// "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Mobile Safari/537.36",
		"Chrome/96.0.4664.93",
	)
	req.Header.Set(
		"accept",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	)
	req.Header.Set(
		"accept-encoding",
		"gzip, deflate, br",
	)
	req.Header.Set(
		"accept-language",
		"en-GB,en;q=0.9,en-US;q=0.8,id;q=0.7,ml;q=0.6,zh-CN;q=0.5,zh;q=0.4",
	)
	req.Header.Set(
		"upgrade-insecure-requests",
		"1",
	)
	req.Header.Set(
		"Referer",
		referer,
	)
}

func printToTxt(in string) {
	data := fmt.Sprintf(in)

	err := ioutil.WriteFile("Survey.txt", []byte(data), 0644)
	if err != nil {
		log.Fatalf("error writing Survey.txt: %s", err)
	}
}

func getDetailedData(item *model.Item, detailedUrl string, shopName string) *model.Item {
	result := &model.Item{}
	result = item
	client := &http.Client{}

	referer := ""
	if shopName == "" {
		referer = item.DetailedLink
	} else {
		referer = fmt.Sprintf(
			"https://www.tokopedia.com/%v/%v?src=topads",
			// standardizeNameForReferer(item.Store),
			shopName,
			standardizeNameForReferer(result.Name),
		)
	}

	req, _ := http.NewRequest(
		"GET",
		// fmt.Sprintf(detailedUrl),
		referer,
		nil,
	)
	// Referer: https://www.tokopedia.com/iphone-store-2/oppo-f1s-4-32gb-garansi-1-tahun-gold
	// fmt.Println(referer)
	setDetailedHeaders(req, detailedUrl, referer)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		//it will be solved by tommorow
		if res.StatusCode != 404 {
			fmt.Println(result)
			fmt.Println(referer)
			log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		}
	}

	// encoding := res.Header.Get("content-encoding")
	// fmt.Println(encoding)
	zr, err := gzip.NewReader(res.Body)
	zr.Name = "This is the name"
	if err != nil {
		log.Fatalf(err.Error())
	}

	if err := zr.Close(); err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(zr)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(doc.Text())
	text := doc.Text()
	result.Rating = text[strings.Index(text, "\"rating\":")+9 : strings.Index(text, "\"rating\":")+12]
	reg, err := regexp.Compile("[^\\w.]+")
	if err != nil {
		log.Fatalf(err.Error())
	}
	result.Rating = reg.ReplaceAllString(result.Rating, "")
	descTemp := text[strings.Index(text, "\"title\":\"Deskripsi\",\"subtitle\":")+32:]
	result.Description = strings.Split(descTemp, "\",\"applink\"")[0]

	// printToTxt(doc.Text())

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		img, ok := s.Attr("src")
		if ok && strings.Index(img, "images.tokopedia.net") != -1 {
			result.ImageURL = img
		}
	})
	// fmt.Println(result)
	item = result
	return result
}

func standardizeNameForReferer(in string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(in, "-")
	processedString = strings.ToLower(processedString)
	for {
		processedString = strings.Replace(processedString, "--", "-", -1)
		if strings.Index(processedString, "--") == -1 {
			break
		}
	}
	for {
		if strings.Index(processedString, "-") == 0 {
			processedString = processedString[1:]
		} else {
			break
		}
	}
	for {
		if strings.LastIndex(processedString, "-") == len(processedString)-1 {
			processedString = processedString[:len(processedString)-1]
		} else {
			break
		}
	}

	return processedString
}

func GetDataPerPage(items []*model.Item, page int) []*model.Item {
	client := &http.Client{}
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://www.tokopedia.com/p/handphone-tablet/handphone?ob=23&page=%v", page),
		nil,
	)
	setHeaders(req)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// items := make([]model.Item, 0)
	doc.Find("div[data-testid]").
		FilterFunction(func(i int, s *goquery.Selection) bool {
			str, _ := s.Attr("data-testid")
			return str == "divProductWrapper"
		}).Each(func(i int, s *goquery.Selection) {
		item := new(model.Item)
		detailedLink, _ := s.Parent().Attr("href")
		item.DetailedLink = detailedLink

		s.Children().Each(func(j int, sc *goquery.Selection) {
			// if sc.HasClass("css-79elbk") {
			// 	sc.Children().Children().Children().Each(func(_ int, scc *goquery.Selection) {
			// 		// fmt.Println(j, sc.Text())
			// 		v, e := scc.Attr("src")
			// 		if e {
			// 			item.ImageURL = v
			// 			// fmt.Println(v)
			// 		}
			// 	})
			// } else
			if sc.HasClass("css-11s9vse") {
				sc.Children().Each(func(k int, scc *goquery.Selection) {
					// fmt.Println(i, k, scc.Text())
					if scc.HasClass("css-1bjwylw") {
						item.Name = scc.Text()
					} else if i == 0 && k == 2 {
						scc.Children().Children().Each(func(l int, sccc *goquery.Selection) {
							if sccc.HasClass("css-o5uqvq") {
								price := sccc.Text()
								item.Price, _ = strconv.Atoi(strings.Replace(strings.Replace(price, ".", "", -1), "Rp", "", -1))
							}
						})
					} else if i == 0 && k == 4 {
						scc.Children().Children().Each(func(l int, sccc *goquery.Selection) {
							// fmt.Println(l, sccc.Text())
							item.Store = sccc.Text()
						})
					} else if k == 1 {
						price := ""
						scc.Children().Each(func(l int, sccc *goquery.Selection) {
							if l == 0 || l == 1 {
								price = sccc.Text()
							}
						})
						price = strings.Replace(strings.Replace(price, ".", "", -1), "Rp", "", -1)
						item.Price, _ = strconv.Atoi(price)
					} else if k == 2 {
						scc.Children().Children().Each(func(l int, sccc *goquery.Selection) {
							item.Store = sccc.Text()
						})
					}
				})
			}
		})

		items = append(items, item)
		// fmt.Println(len(items))
		// if len(items) >= 100 {
		// 	return
		// }
	})
	return items
}
func GetAllData(total int) []*model.Item {
	items := make([]*model.Item, 0)
	i := 1
	for {
		items = append(items, GetDataPerPage(items, i)...)
		if len(items) >= total {
			items = items[0:total]
			break
		}
		i = i + 1
	}
	res := []*model.Item{}
	for _, item := range items {
		splitter1 := regexp.MustCompile(`https%3A%2F%2Fwww.tokopedia.com%2F`)
		splitter2 := regexp.MustCompile(`%2F`)
		shopName := ""

		afterSplit1 := splitter1.Split(item.DetailedLink, -1)
		if len(afterSplit1) > 1 {
			shopName = splitter2.Split(afterSplit1[1], -1)[0]
		}

		res = append(res, getDetailedData(item, item.DetailedLink, shopName))
	}
	return items
}
