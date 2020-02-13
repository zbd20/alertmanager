package telephone

import (
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/alertmanager/config"
	commoncfg "github.com/prometheus/common/config"
)

var n *Notifier

func init() {
	n, _ = New(
		&config.TelephoneConfig{
			AppKey:        "",
			AppSecret:     "",
			UserName:      "",
			Authorization: "",
			BaseURL:       "",
			DisplayNumber: "",
			TemplateId:    "",
			Operators:     []string{""},
			HTTPConfig:    &commoncfg.HTTPClientConfig{},
		},
		log.NewNopLogger(),
	)
}

func TestNotifier_InitialAccessToken(t *testing.T) {
	err := n.InitialAccessToken()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(n.String())
	}
}

func TestNotifier_RefreshAccessToken(t *testing.T) {
	err := n.RefreshAccessToken()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(n.String())
	}
}

func TestNotifier_Send(t *testing.T) {
	err := n.Send("110")
	if err != nil {
		t.Error(err)
	} else {
		t.Log()
	}
}
