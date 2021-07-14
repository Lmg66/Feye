package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"gopkg.in/yaml.v2"
)

/*
配置文件初始化
*/
type conf struct {
	Authorization string `yaml:"Authorization"`
	Timesleep     int    `yaml:"Timesleep"`
	Proxy         string `yaml:"Proxy"`
	Output        string `yaml:"Output"`
	Email         string `yaml:"Email"`
	Password      string `yaml:"Password"`
}

func (c *conf) getConf() *conf {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		fmt.Println(err.Error())
	}
	return c
}

/*
fofa配置检测
*/
func Fofa_Config() *conf {
	var config conf
	conf := config.getConf()
	if conf.Authorization == "" {
		fmt.Printf("请配置Authorization")
		os.Exit(0)
	} else {
		fmt.Println("检测到Authorization,请确保正确")
	}
	return conf
}

/*
Zoomeye配置检测
*/
func Zoomeye_config() *conf {
	var config conf
	conf := config.getConf()
	if config.Email == "" || config.Password == "" {
		fmt.Println("请配置Email和Password")
		os.Exit(0)
	} else {
		fmt.Println("检测到Email和Password,正在尝试获取access_token请确保正确")
	}
	return conf
}

/*
base64编码
*/
func StdEncoding(SearchKEY string) string {
	b := []byte(SearchKEY)
	searchbs64 := base64.StdEncoding.EncodeToString(b)
	return searchbs64
}

/*
保存爬取内容文件
*/
func Savetotxt(out [][]string, Output string) {
	if Output == "" {
		Output = "output.txt"
	}
	fp, err := os.OpenFile(Output, os.O_CREATE|os.O_APPEND, 6)
	if err != nil {
		fmt.Println("文件打开失败")
	}
	defer fp.Close()
	for i := 0; i < len(out); i++ {
		if out[i][1] != "" {
			_, err := fp.WriteString(out[i][1] + "\n")
			if err != nil {
				fmt.Println("写入文件失败。")
			}
		}
	}
}

/*
fofa爬虫
*/
func Fofa_Requests(config *conf) {
	c := colly.NewCollector()
	if config.Proxy != "" {
		proxy, err := proxy.RoundRobinProxySwitcher(config.Proxy)
		if err != nil {
			log.Fatal(err)
		}
		c.SetProxyFunc(proxy)
	}
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
		os.Exit(0)
	})
	c.OnRequest(func(e *colly.Request) {
		e.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0")
		e.Headers.Set("Authorization", config.Authorization)
		fmt.Println("visiting:", e.URL)
	})
	var number []string = make([]string, 0)
	c.OnHTML("ul[class='el-pager']", func(e *colly.HTMLElement) {
		e.ForEach("li", func(i int, item *colly.HTMLElement) {
			if item.Text != "" {
				number = append(number, item.Text)
			}
			i++
		})
	})
	c.OnResponse(func(r *colly.Response) {
		match, _ := regexp.MatchString("\"message\":\"资源访问权限不足\",", string(r.Body))
		if match {
			fmt.Println("资源访问权限不足")
		} else {
			match, _ := regexp.MatchString("\"link\":\"(.*?)\",", string(r.Body))
			if match {
				ret := regexp.MustCompile(`"link":"(.*?)"`)
				link := ret.FindAllStringSubmatch(string(r.Body), -1)
				Savetotxt(link, config.Output)
				for i := 0; i < len(link); i++ {
					if link[i][1] != "" {
						fmt.Println(link[i][1])
					}
				}
			}
		}
	})
	var SearchKEY string
	fmt.Printf("请输入fofa搜索关键字:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	SearchKEY = scanner.Text()
	searchbs64 := StdEncoding(SearchKEY)
	searchbs64 = url.QueryEscape(searchbs64) //url编码
	url := "http://fofa.so/result?&qbase64=" + searchbs64
	c.Visit(url)
	var num string
	if len(number) == 0 {
		fmt.Println("关键字存在页面为1页")
		num = "1"
	} else {
		num = number[len(number)-1]
		fmt.Println("关键字存在页面:", num)
	}
	num_int, err := strconv.Atoi(num)
	if err != nil {
		fmt.Println(err.Error())
	}
	var StartPage, StopPage int
	fmt.Printf("请输入开始页码:")
	fmt.Scanln(&StartPage)
	fmt.Printf("请输入终止页码:")
	fmt.Scanln(&StopPage)
	if StopPage < StartPage {
		fmt.Println("StopPage < StartPage")
		os.Exit(0)
	}
	if StartPage <= 0 || StopPage > num_int {
		fmt.Println("StartPage < 0 || StopPage > max(pagenumber)")
		os.Exit(0)
	}
	for i := StartPage; i <= StopPage; i++ {
		fmt.Printf("正在爬取第%d页\n", i)
		url = "https://api.fofa.so/v1/search?qbase64=" + searchbs64 + "&full=false&pn=" + strconv.Itoa(i) + "&ps=10"
		c.Visit(url)
		if i != StopPage {
			time.Sleep(time.Duration(config.Timesleep)*time.Second + time.Duration(rand.Int31n(int32(config.Timesleep*100)))*time.Millisecond)
		}
	}

}

/*
Zoomeye爬虫
*/
/*
传递出Zoomeye获得的access_token 增加给OnRequest
*/
var Authorization string

func get_Authorization(access_token string) {
	Authorization = "JWT " + access_token
}

/*
Zoomeye爬虫
*/
func Zoomeye_Requests(config *conf) {
	c := colly.NewCollector()
	if config.Proxy != "" {
		proxy, err := proxy.RoundRobinProxySwitcher(config.Proxy)
		if err != nil {
			log.Fatal(err)
		}
		c.SetProxyFunc(proxy)
	}
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", string(r.Body), "\nError:", err)
		if Authorization == "" {
			fmt.Println("未获得access_token,请检查账号密码|代理")
		}
		os.Exit(0)
	})
	c.OnRequest(func(e *colly.Request) {
		e.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0")
		if Authorization != "" {
			e.Headers.Set("Authorization", Authorization)
		}
		fmt.Println("visiting:", e.URL)
	})
	c.OnResponse(func(r *colly.Response) {
		match, _ := regexp.MatchString("\"access_token\": \"(.*?)\"", string(r.Body))
		if match {
			ret := regexp.MustCompile(`"access_token": "(.*?)"`)
			link := ret.FindAllStringSubmatch(string(r.Body), -1)
			if link[0][1] != "" {
				get_Authorization(link[0][1]) //传递出access_token 增加Headers
				fmt.Println("access_token successful!")
			}
		}
		match_num, _ := regexp.MatchString("\"search\": (.*?),", string(r.Body))
		if match_num {
			ret := regexp.MustCompile(`"search": (.*?),`)
			link := ret.FindAllStringSubmatch(string(r.Body), -1)
			fmt.Println("API数据剩余额度:" + link[0][1])
		}
		match_total, _ := regexp.MatchString("\"total\": (.*?),", string(r.Body))
		if match_total {
			ret := regexp.MustCompile(`"total": (.*?),`)
			link := ret.FindAllStringSubmatch(string(r.Body), -1)
			fmt.Println("总计:", link[0][1])
		}
		match_ip, _ := regexp.MatchString("\"ip\": \"(.*?)\",.*?\"port\": (.*?),", string(r.Body))
		if match_ip {
			ret := regexp.MustCompile(`"ip": "(.*?)",.*?"port": (.*?),`)
			link := ret.FindAllStringSubmatch(string(r.Body), -1)
			for i := 0; i < len(link); i++ {
				if link[i][1] != "" {
					Lmg := link[i][1] + ":" + link[i][2]
					fmt.Println(Lmg)
					link[i][1] = Lmg
				}
				Savetotxt(link, config.Output)
			}
		}

	})
	type date struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	dt := &date{
		Username: config.Email,
		Password: config.Password,
	}
	da, err := json.Marshal(dt)
	if err != nil {
		fmt.Println(err)
	}
	c.PostRaw("https://api.zoomeye.org/user/login", da)
	c.Visit("https://api.zoomeye.org/resources-info")
	var SearchKEY string
	fmt.Printf("请输入Zoomeye搜索关键字:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	SearchKEY = scanner.Text()
	SearchKEY = url.QueryEscape(SearchKEY)
	var StartPage, StopPage int
	fmt.Printf("请输入开始页码:")
	fmt.Scanln(&StartPage)
	fmt.Printf("请输入终止页码:")
	fmt.Scanln(&StopPage)
	if StopPage < StartPage || StartPage < 0 {
		fmt.Println("StopPage < StartPage or StartPage < 0")
		os.Exit(0)
	}
	for i := StartPage; i <= StopPage; i++ {
		fmt.Printf("正在爬取第%d页\n", i)
		url_request := "https://api.zoomeye.org/host/search?query=" + SearchKEY + "&page=" + strconv.Itoa(i)
		c.Visit(url_request)
		if i != StopPage {
			time.Sleep(time.Duration(config.Timesleep)*time.Second + time.Duration(rand.Int31n(int32(config.Timesleep*100)))*time.Millisecond)
		}
	}
}

func main() {

	fofa := flag.Bool("fofa", false, "fofa数据采集")
	zoom := flag.Bool("zoom", false, "zoomeye数据采集")
	flag.Parse()
	if *fofa {
		lnng := Fofa_Config()
		Fofa_Requests(lnng)
	}
	if *zoom {
		lnng := Zoomeye_config()
		Zoomeye_Requests(lnng)
	}
	if !*fofa && !*zoom {
		fmt.Printf("-h 查看命令")
	}
}
