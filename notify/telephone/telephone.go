package telephone

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"
	commoncfg "github.com/prometheus/common/config"
)

// Notifier implements a Notifier for voice notifications current.
type Notifier struct {
	conf   *config.HWCConfig
	logger log.Logger
	client *http.Client

	accessToken   string
	refreshToken  string
	expiresIn     int64
	accessTokenAt time.Time
}

type TokenResult struct {
	Resultcode   string `json:"resultcode,omitempty"`
	Resultdesc   string `json:"resultdesc,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

// New returns a new HuaWeiCloud notifier.
func New(c *config.HWCConfig, l log.Logger) (*Notifier, error) {
	client, err := commoncfg.NewClientFromConfig(*c.HTTPConfig, "telephone", false)
	if err != nil {
		return nil, err
	}

	// initial hwc access token
	url := fmt.Sprintf("%s/rest/fastlogin/v1.0?app_key=%s&username=%s", c.BaseURL, c.AppKey, c.UserName)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Authorization", c.Authorization)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode/100 != 2 {
			return nil, fmt.Errorf("unexpected status code %v from %s", resp.StatusCode, url)
		}
	}
	var tokenResult TokenResult
	if err := json.NewDecoder(resp.Body).Decode(&tokenResult); err != nil {
		return nil, err
	}

	n := Notifier{conf: c, logger: l, client: client}
	n.accessToken = tokenResult.AccessToken
	n.expiresIn, _ = strconv.ParseInt(tokenResult.ExpiresIn, 10, 64)
	n.refreshToken = tokenResult.RefreshToken
	n.accessTokenAt = time.Now()

	return &n, nil
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	// Refresh AccessToken over 47 hours
	if n.accessToken == "" || time.Since(n.accessTokenAt) > 47*time.Hour {
		url := fmt.Sprintf("%s/omp/oauth/refresh?app_key=%s&app_secret=%s&grant_type=refresh_token&refresh_token=%s", n.conf.BaseURL, n.conf.AppKey, n.conf.AppSecret, n.refreshToken)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return true, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

		resp, err := n.client.Do(req.WithContext(ctx))
		if err != nil {
			return true, notify.RedactURL(err)
		}
		defer notify.Drain(resp)

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode/100 != 2 {
				return true, fmt.Errorf("unexpected status code %v from %s", resp.StatusCode, url)
			}
		}
		var tokenResult TokenResult
		if err := json.NewDecoder(resp.Body).Decode(&tokenResult); err != nil {
			return true, err
		}

		// Cache accessToken
		n.accessToken = tokenResult.AccessToken
		n.expiresIn, _ = strconv.ParseInt(tokenResult.ExpiresIn, 10, 64)
		n.refreshToken = tokenResult.RefreshToken
		n.accessTokenAt = time.Now()
	}

	// send voice notify
	url := fmt.Sprintf("%s/rest/httpsessions/callnotify/%s?app_key=%s&access_token=%s", n.conf.BaseURL, "v2.0", n.conf.AppKey, n.accessToken)
	for _, operator := range n.conf.Operators {
		s := `{
			"displayNbr": "%s", 
			"calleeNbr": "+86%s", 
			"playInfoList": [{"templateId":"%s", "templateParas":["1"], "collectInd": 0}]
		}`
		body := fmt.Sprintf(s, n.conf.DisplayNumber, operator, n.conf.TemplateId)
		req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
		if err != nil {
			return true, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := n.client.Do(req.WithContext(ctx))
		if err != nil {
			return true, err
		}
		defer notify.Drain(resp)
	}

	return false, nil
}
