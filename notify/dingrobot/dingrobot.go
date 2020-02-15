package dingrobot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"net/http"

	"github.com/go-kit/kit/log"
	commoncfg "github.com/prometheus/common/config"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

type dingTalkNotification struct {
	MessageType string                        `json:"msgtype"`
	Markdown    *dingTalkNotificationMarkdown `json:"markdown,omitempty"`
}

type dingTalkNotificationMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type DingRobot struct {
	conf   *config.DingRobotConfig
	tmpl   *template.Template
	logger log.Logger
}

// New returns a new Dingtalk robot notifier.
func New(c *config.DingRobotConfig, t *template.Template, l log.Logger) (*DingRobot, error) {
	return &DingRobot{conf: c, tmpl: t, logger: l}, nil
}

// Notify implements the Notifier interface.
func (d *DingRobot) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	level.Debug(d.logger).Log("start to send ding robot")
	var (
		tmplErr error
		data    = notify.GetTemplateData(ctx, d.tmpl, as, d.logger)
		tmpl    = notify.TmplText(d.tmpl, data, &tmplErr)
		title   = tmpl(d.conf.Title)
		content = tmpl(d.conf.Content)
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

	c, err := commoncfg.NewClientFromConfig(*d.conf.HTTPConfig, "dingRobot", false)
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

func (d *DingRobot) retry(statusCode int) (bool, error) {
	// Webhooks are assumed to respond with 2xx response codes on a successful
	// request and 5xx response codes are assumed to be recoverable.
	if statusCode/100 != 2 {
		return (statusCode/100 == 5), fmt.Errorf("unexpected status code %v from %s", statusCode, d.conf.WebhookURL)
	}

	return false, nil
}
