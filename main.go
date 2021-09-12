package main


import (
	"fmt"
	"net/http"
	"io/ioutil"
	"time"
	"math/rand"
	"strconv"
	"strings"
	"regexp"
	"sync"
	"os"
	"io"
)

var (
	page_C = make(chan int, 200)
	write_C = make(chan string)
	goNum int
	lock_Book sync.Mutex
	page_Min int
	page_Max int
	book_Cnt int
	book_All int
	GetUrl string
	GetBook string
	url_page string
	reg_Page_max = regexp.MustCompile(`id="maxpage" value=".*?">`)
	reg_Book_id = regexp.MustCompile(`data-url="/d/[\d]{0,}"`)
	reg_Down_url = regexp.MustCompile(`/down\?id=.+?&p=1`)
	reg_Book_name = regexp.MustCompile(`name": ".+?",`)
)

var UA = []string{"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36",
				  "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11",
				  "Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/534.16 (KHTML, like Gecko) Chrome/10.0.648.133 Safari/534.16",
				  "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",
				  "Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
				  "Mozilla/5.0 (Linux; U; Android 2.2.1; zh-cn; HTC_Wildfire_A3333 Build/FRG83D) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
				  "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.71 Safari/537.1 LBBROWSER"}


func get_Html_Url(url string) (string, int) {
	rand.Seed(time.Now().Unix())
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "Http.NewRequest err", 1
	}
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", UA[rand.Intn(len(UA))])
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	resp, err := client.Do(req)
	if err != nil {
		return "client.Do err", 2
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "ioutil.ReadAll err", 3
		} else {
			return string(body), 200
		}
	}
	return "", 4
}

func regexp__Book_Id(html string) [][]string {
	ret:= reg_Book_id.FindAllStringSubmatch(html, -1)
	return ret
}

func regexp__Book_Name(html string) string {
	ret:= reg_Book_name.FindString(html)
	return strings.TrimLeft(strings.TrimRight(ret, `",`), `name": "`)
}

func regexp__Down_Url(html string) string {
	ret:= reg_Down_url.FindString(html)
	return ret
}
func regexp_Page_Max(html string) int {
	ret:= reg_Page_max.FindString(html)
	if ret == "" {
		return 0
	}
	s := strings.TrimLeft(strings.TrimRight(ret, `">`), `id="maxpage" value="`)
	i, _ := strconv.Atoi(s)
	return i
}


func DownloadFile(filepath string, url string) error {
		 // Get the data
		 resp, err := http.Get(url)
		 if err != nil {
			 return err
		 }
		 defer resp.Body.Close()
	 
		 // Create the file
		 out, err := os.Create(filepath)
		 if err != nil {
			 return err
		 }
		 defer out.Close()
	 
		 // Write the body to file
		 _, err = io.Copy(out, resp.Body)
		 return err
}

func get_Book_Url(page int, book_id string) {
	for {
		Down_Html, status := get_Html_Url("https://m.aixdzs.com" + book_id)
		if status == 200 {
			down_Book_Url := regexp__Down_Url(Down_Html)
			down_Book_Name := regexp__Book_Name(Down_Html)
			if GetUrl == "y" {
				f := down_Book_Name + " https://m.aixdzs.com" + down_Book_Url + "\n"
				lock_Book.Lock()
				wirte(url_page, f)
				book_Cnt++
				lock_Book.Unlock()
			}
			if GetBook == "y" {
				for {
					err := DownloadFile(down_Book_Name + ".rar", "https://m.aixdzs.com" + down_Book_Url)
					if err == nil {
						break
					}
				}
				f := down_Book_Name + " https://m.aixdzs.com" + down_Book_Url + "\n"
				lock_Book.Lock()
				wirte(url_page, f)
				book_Cnt++
				lock_Book.Unlock()
			}
			fmt.Printf("位于:%d页-共%d页-下载%d/%d-%s >>https://m.aixdzs.com%s\n", page, page_Max, book_Cnt, book_All, down_Book_Name, down_Book_Url)
			break
		} else {
			fmt.Println(">> 本地网络 或 目标网站出现异常,正在重新访问.")
			time.Sleep(time.Second * 2)
		}
	}
}

func get_Book_Id(url string){
	for {
		select {
		case page := <-page_C:
			for {
				sub_html, status := get_Html_Url(url + strconv.Itoa(page))
				if status == 200 {
					book_Id := regexp__Book_Id(sub_html)
					for _, v := range book_Id {
						bookId := strings.TrimLeft(strings.TrimRight(v[0], `"`), `data-url="`)
						get_Book_Url(page, bookId)
					}
					break
				} else {
					fmt.Println(">> 本地网络 或 目标网站出现异常,正在重新访问.")
					time.Sleep(time.Second * 2)
				}
			}
		default:
			time.Sleep(time.Second * 10)
		}
	}
}

func wirte(folder, data string) {
    txt, err := os.OpenFile(folder + ".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
    if err == nil {
        defer txt.Close()
    }
    txt.Write([]byte(data))
}


func main() {
	var url string
	goNum = 50
	page_Min = 1
	book_Cnt = 0
	book_All = 1
	GetUrl = "y"
	GetBook = "n"
	fmt.Printf("请输入分类页码:")
	fmt.Scanf("%s \r\n", &url_page)
	if url_page == "" {
		fmt.Println("请输入分类页码")
		return
	}
	fmt.Printf("爬取地址y/n:")
	fmt.Scanf("%s \r\n", &GetUrl)
	if GetUrl == "n" {
		fmt.Printf("下载书籍y/n:")
		fmt.Scanf("%s \r\n", &GetBook)
	}
	fmt.Printf("数量:")
	fmt.Scanf("%d \r\n", &book_All)
	if book_All <= 0 {
		fmt.Println("请输入正确数量")
		return
	}
	fmt.Printf("输入异步数量*请勿输入过高，***根据机器性能及网络自行调试***:")
	fmt.Scanf("%d \r\n", goNum)
	//获取总页数
	for {
		url = "https://m.aixdzs.com/sort/" + url_page + "?page="
		sub_html, status := get_Html_Url(url + strconv.Itoa(page_Min))
		page_Max = regexp_Page_Max(sub_html)
		if status == 200 && page_Max != 0 {
			break
		}
		fmt.Println(">> 本地网络 或 目标网站出现异常,正在重新访问.")
		time.Sleep(time.Second * 2)
	}
	for goNum >= 0 {
		go get_Book_Id(url)
		goNum--
	}
	for {
		for page_Min <= page_Max {
			page_C <- page_Min
			page_Min++
		}
		if book_All == book_Cnt {
			fmt.Println("下载完成")
			break
		}
		time.Sleep(time.Second * 2)
	}
}