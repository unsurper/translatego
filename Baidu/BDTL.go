package Baidu

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func BaiduTranslate(appid string, appkey string, fr string, to string, query string) string {
	client := &http.Client{Timeout: 5 * time.Second}

	rand.Seed(int64(time.Now().UnixNano()))
	salt := strconv.Itoa(rand.Intn(32768) + (65536 - 32768))
	sign := MD5(appid + query + salt + appkey)

	payload := url.Values{"appid": {appid}, "q": {query}, "from": {fr}, "to": {to}, "salt": {salt}, "sign": {sign}}
	apiURL := "https://fanyi-api.baidu.com/api/trans/vip/translate"
	resp, err := client.Post(apiURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(payload.Encode()))

	if err == nil {
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		JO := gjson.ParseBytes(data)
		if JO.Exists() {
			if JO.Get("error_code").Int() > 0 { // 如果存在 这个字段肯定不会是零的咯
				return JO.Get("error_msg").String()
			}

			if JO.Get("trans_result").IsArray() {
				return JO.Get("trans_result").Array()[0].Get("dst").String()
			}
		}
	}

	return ""
}
func MD5(str string) string {
	_md5 := md5.New()
	_md5.Write([]byte(str))
	return hex.EncodeToString(_md5.Sum([]byte("")))
}
