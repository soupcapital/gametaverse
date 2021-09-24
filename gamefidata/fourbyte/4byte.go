package fourbyte

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
)

const (
	ByteAPI = "https://www.4byte.directory/api/v1/signatures/?hex_signature="
)

type FourByteDB struct {
	db map[string]string
}

var (
	DB *FourByteDB
)

func init() {
	DB = &FourByteDB{
		db: make(map[string]string),
	}
}

func (fb *FourByteDB) Get(sig string) (method string, err error) {
	if m, ok := fb.db[sig]; ok {
		return m, nil
	}

	url := fmt.Sprintf("%s%s", ByteAPI, sig)
	log.Info("rpc:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Error("Query method info error:%s", err.Error())
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//TODO: retry
		log.Error("read all error:%s", err.Error())
		return "", err
	}

	type RspJSON struct {
		Count    int         `json:"count"`
		Next     interface{} `json:"next"`
		Previous interface{} `json:"previous"`
		Results  []struct {
			ID             int       `json:"id"`
			CreatedAt      time.Time `json:"created_at"`
			TextSignature  string    `json:"text_signature"`
			HexSignature   string    `json:"hex_signature"`
			BytesSignature string    `json:"bytes_signature"`
		} `json:"results"`
	}

	var info RspJSON
	err = json.Unmarshal(body, &info)
	if err != nil {
		//TODO: retry
		log.Error("Unmarshal JSON error:%s", err.Error())
		return "", err
	}
	if info.Count == 0 {
		return "", nil
	}
	for _, record := range info.Results {
		if record.HexSignature == sig {
			method = record.TextSignature
			methods := strings.Split(method, "(")
			if len(methods) == 0 {
				return "", nil
			}
			fb.db[sig] = methods[0]
			return methods[0], nil
		}
	}

	return
}
