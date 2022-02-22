package creator

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log"

	"github.com/PuerkitoBio/goquery"
)

type ContractInfo struct {
	Height   string `json:"height"`
	Creator  string `json:"creator"`
	Datetime string `json:"datetime"`
	Contract string `json:"contract"`
}

var _browser browser

func process(addr string, chain string, proxy string) {
	_browser.process(addr, chain, proxy)
}

type browser struct {
	proxy string
}

func (hdl *browser) process(addr string, chain string, proxy string) {
	hdl.proxy = proxy
	rpcBase := fmt.Sprintf("https://etherscan.io/txs?a=%s&f=5", addr)
	switch chain {
	case "bsc":
		rpcBase = fmt.Sprintf("https://bscscan.com/txs?a=%s&f=5", addr)
	case "polygon":
		rpcBase = fmt.Sprintf("https://polygonscan.com/txs?a=%s&f=5", addr)
	case "eth":
		rpcBase = fmt.Sprintf("https://etherscan.io/txs?a=%s&f=5", addr)
	default:
		log.Printf("chain should be bsc/polygon/eth")
		return
	}
	var contracts []ContractInfo
	i := 0
	for {
		cs, err := hdl.queryPage(rpcBase, i)
		if err != nil {
			log.Printf("query Page error:%s", err.Error())
			break
		}
		if len(cs) == 0 {
			break
		}
		contracts = append(contracts, cs...)

		i += 1
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("========================================================================================")
	for _, c := range contracts {
		//fmt.Printf("[%s@%s] create:%s\n", c.Creator, c.Datetime, c.Contract)
		fmt.Printf("%s\n", c.Contract)
	}
	fmt.Println("========================================================================================")
}

func (hdl *browser) queryPage(addr string, page int) (contracts []ContractInfo, err error) {
	requrl := fmt.Sprintf("%s&p=%d", addr, page)
	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		log.Printf("new request error:%s", err.Error())
		return
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if len(hdl.proxy) != 0 {
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(hdl.proxy)
		}
		tr = &http.Transport{
			Proxy:           proxy,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	cli := &http.Client{
		Transport: tr,
	}
	resp, err := cli.Do(req)
	if err != nil {
		log.Printf("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("status code error: %d %s", resp.StatusCode, resp.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("create goquery err:%s", err.Error())
		return
	}

	doc.Find("table tbody").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "There are no matching entries") {
			return
		}
		s.Find("tr").Each(func(i int, s *goquery.Selection) {
			blockHeight := ""
			who := ""
			datetime := ""
			contract := ""
			s.Find("td").Each(func(i int, s *goquery.Selection) {
				switch i {
				case 4:
					datetime = s.Text()
				case 6:
					who = s.Text()
				case 3:
					blockHeight = s.Text()
				case 8:
					s.Find("a").Each(func(i int, s *goquery.Selection) {
						if title, ok := s.Attr("title"); ok {
							words := strings.Split(title, " ")
							if len(words) >= 2 {
								contract = words[len(words)-1]
							}
						}
					})
				}

			})
			// log.Info("blockHeight[%s] who[%s] datetime[%s] contract[%s]",
			// 	blockHeight,
			// 	who,
			// 	datetime,
			// 	contract)
			c := ContractInfo{
				Height:   blockHeight,
				Creator:  who,
				Datetime: datetime,
				Contract: contract,
			}
			contracts = append(contracts, c)
		})
	})
	return
}
