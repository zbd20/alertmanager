package dingapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	dingtalk "github.com/iyacontrol/godingtalk"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
	commoncfg "github.com/prometheus/common/config"
)

type DingApp struct {
	conf   *config.DingAppConfig
	tmpl   *template.Template
	logger log.Logger
}

type dingTalkNotification struct {
	MessageType string                        `json:"msgtype"`
	Markdown    *dingTalkNotificationMarkdown `json:"markdown,omitempty"`

	Alerts []*types.Alert `json:"alerts"`
}

type dingTalkNotificationMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// New returns a new Dingtalk app message notifier.
func New(c *config.DingAppConfig, t *template.Template, l log.Logger) (*DingApp, error) {
	return &DingApp{conf: c, tmpl: t, logger: l}, nil
}

// Notify implements the Notifier interface.
func (d *DingApp) oldNotify(ctx context.Context, as ...*types.Alert) (bool, error) {

	var (
		tmplErr error
		data    = notify.GetTemplateData(ctx, d.tmpl, as, d.logger)
		tmpl    = notify.TmplText(d.tmpl, data, &tmplErr)
		title   = tmpl(d.conf.Title)
		content = tmpl(d.conf.Content) + "\n" + time.Now().Format("2006-01-02 15:04:05")
	)
	if tmplErr != nil {
		return false, fmt.Errorf("failed to template 'title' or 'content': %v", tmplErr)
	}

	client := dingtalk.NewDingTalkClient(d.conf.CorpID, d.conf.CorpSecret)
	client.RefreshAccessToken()

	toUser := strings.Join(d.conf.Operators, "|")

	err := client.SendAppMarkDownMessage(d.conf.AgentID, toUser, content, title)
	if err != nil {
		level.Debug(d.logger).Log("send failure", toUser)
		return true, err
	}

	return false, nil
}

// Notify by means of webhook
func (d *DingApp) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {

	fmt.Printf("dingApp alert: %v\n", as)
	var (
		tmplErr error
		data    = notify.GetTemplateData(ctx, d.tmpl, as, d.logger)
		tmpl    = notify.TmplText(d.tmpl, data, &tmplErr)
		title   = tmpl(d.conf.Title)
		content = tmpl(d.conf.Content) + "\n" + time.Now().Format("2006-01-02 15:04:05")
	)
	if tmplErr != nil {
		return false, fmt.Errorf("failed to template 'title' or 'content': %v", tmplErr)
	}

	var msg = &dingTalkNotification{
		MessageType: "markdown",
		Markdown: &dingTalkNotificationMarkdown{
			Title: title,
			Text:  content,
		},
		Alerts: as,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", d.conf.WebhookURL, &buf)
	if err != nil {
		return true, err
	}
	req.Header.Set("Content-Type", "application/json")

	c, err := commoncfg.NewClientFromConfig(*d.conf.HTTPConfig, "dingApp", false)
	if err != nil {
		return false, err
	}

	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return true, err
	}
	resp.Body.Close()

	return d.retry(resp.StatusCode)
}

func (d *DingApp) retry(statusCode int) (bool, error) {
	// Webhooks are assumed to respond with 2xx response codes on a successful
	// request and 5xx response codes are assumed to be recoverable.
	if statusCode/100 != 2 {
		return statusCode/100 == 5, fmt.Errorf("unexpected status code %v from %s", statusCode, d.conf.WebhookURL)
	}

	return false, nil
}
