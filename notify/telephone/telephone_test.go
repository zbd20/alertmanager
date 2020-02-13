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
			AppKey:        "4MaDaD3NH55UCcE7VrzzIicAgtP3",
			AppSecret:     "8fx3qLQysC5N62K0sgLf08xY3ox2",
			UserName:      "KuaiLeQie",
			Authorization: "",
			BaseURL:       "https://rtcvc.cn-north-1.myhuaweicloud.com:10643",
			DisplayNumber: "",
			TemplateId:    "0ea04e1119104871944958272442d32f",
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
	err := n.Send("18392505264")
	if err != nil {
		t.Error(err)
	} else {
		t.Log()
	}
}
