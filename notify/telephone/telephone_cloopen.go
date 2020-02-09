package telephone

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/types"
)

var (
	baseURL    = "https://app.cloopen.com:8883"
	bersion    = "2013-12-26"
	timeFormat = "20060102150405"
)

type Telephone struct {
	conf   *config.TelephoneConfig
	logger log.Logger
}

func NewCloopen(c *config.TelephoneConfig, l log.Logger) (*Telephone, error) {
	return &Telephone{conf: c, logger: l}, nil
}

// Notify implements the Notifier interface.
func (t *Telephone) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	cloopen := &Cloopen{
		AccountSid:   t.conf.AccountSid,
		AppID:        t.conf.AppID,
		AccountToken: t.conf.AccountToken,

		logger: t.logger,
	}

	for _, to := range t.conf.Operators {
		req := &Request{
			To:         to,
			MediaTxt:   t.conf.MediaTxt,
			DisplayNum: t.conf.DisplayNum,
		}
		_, err := cloopen.Send(req)
		if err != nil {
			level.Debug(t.logger).Log("send telephone failure to", to)
			level.Debug(t.logger).Log("send telephone failure err", err.Error())
			return true, err
		}
	}

	return false, nil
}

type Cloopen struct {
	AccountSid   string
	AccountToken string
	AppID        string

	BaseURL string
	Version string

	logger log.Logger
}

type Request struct {
	To         string // 被叫号码，被叫为座机时需要添加区号，如：01052823298；被叫为分机时分机号由‘-’隔开，如：01052823298-3627
	MediaTxt   string
	DisplayNum string
}

// Send 获取所有消息事件信息
func (srv *Cloopen) Send(req *Request) (valid bool, err error) {
	// 请求参数构建
	url := srv.URL()
	body := srv.Body(req)
	headers := srv.Headers()
	httpContentString, err := srv.Request(url, body, headers)
	if err != nil {
		return false, err
	}
	// 返回数据处理
	valid, err = srv.Response(httpContentString)
	if err != nil {
		return false, err
	}
	return valid, err
}

// Response 返回数据处理
func (srv *Cloopen) Response(httpContentString string) (valid bool, err error) {
	// res 返回请求
	res := map[string]interface{}{}
	err = json.Unmarshal([]byte(httpContentString), &res)
	if err != nil {
		return false, err
	}
	if res["statusCode"].(string) != "000000" {
		return false, errors.New(res["statusMsg"].(string))
	}
	return true, err
}

// Request Request 请求
func (srv *Cloopen) Request(url, body string, headers map[string]string) (httpContentString string, err error) {
	// http-Client
	client := &http.Client{}
	// request
	request, _ := http.NewRequest("POST", url, strings.NewReader(body))

	// add headers
	//request.Header.Set("Accept", headers["Accept"])
	//request.Header.Set("Content-Type", headers["Content-Type"])
	//request.Header.Set("Authorization", headers["Authorization"])

	for k, v := range headers {
		request.Header.Set(k, v)
	}
	// post-request
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	httpContent, err := ioutil.ReadAll(resp.Body)
	level.Debug(srv.logger).Log("resp content", string(httpContent))
	return string(httpContent), err
}

// Headers Headers 构建
func (srv *Cloopen) Headers() (headers map[string]string) {
	// format timestamp
	batch := time.Now().Format(timeFormat)
	// auth
	src := srv.AccountSid + ":" + batch
	auth := base64.StdEncoding.EncodeToString([]byte(src))
	return map[string]string{"Accept": "application/json", "Content-Type": "application/json;charset=utf-8", "Authorization": auth}
}

// Body Body 构建
func (srv *Cloopen) Body(req *Request) (body string) {
	s := `{"to": "%s", "displayNum": "%s", "mediaTxt": "%s", "appId": "%s","playTimes": "3"}`
	return fmt.Sprintf(s, req.To, req.DisplayNum, req.MediaTxt, srv.AppID)
}

// URL url 构建
func (srv *Cloopen) URL() (url string) {
	if srv.BaseURL == "" {
		srv.BaseURL = baseURL
	}
	if srv.Version == "" {
		srv.Version = bersion
	}
	// format timestamp
	batch := time.Now().Format(timeFormat)

	// sign
	sign := srv.AccountSid + srv.AccountToken + batch

	// md5
	MD5 := md5.New()
	MD5.Write([]byte(sign))
	lowerSign := hex.EncodeToString(MD5.Sum(nil))

	// lowerSign to upperSign
	upperSign := strings.ToUpper(lowerSign)
	// combine url
	return strings.Join([]string{srv.BaseURL, "/", srv.Version, "/Accounts/", srv.AccountSid, "/Calls/LandingCalls?sig=", upperSign}, "")
}
