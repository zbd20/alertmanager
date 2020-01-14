package dingapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	dingtalk "github.com/iyacontrol/godingtalk"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

type DingApp struct {
	conf *config.DingAppConfig
	tmpl *template.Template
	logger log.Logger
}

// New returns a new Dingtalk app message notifier.
func New(c *config.DingAppConfig, t *template.Template, l log.Logger)  (*DingApp, error) {
	return &DingApp{conf: c, tmpl: t, logger: l}, nil
}

// Notify implements the Notifier interface.
func (d *DingApp) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {

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