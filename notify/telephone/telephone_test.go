package telephone

import (
	"net/http"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/alertmanager/config"
)

var n *Notifier

func init() {
	// please use your authorization information
	n = &Notifier{
		conf: &config.TelephoneConfig{
			AppKey:        "",
			AppSecret:     "",
			UserName:      "",
			Authorization: "",
			BaseURL:       "",
			DisplayNumber: "",
			TemplateId:    "",
			Operators:     []string{""},
		},
		client: &http.Client{},
		logger: log.NewNopLogger(),
	}
}

func TestNotifier_InitialAccessToken(t *testing.T) {
	err := n.InitialAccessToken()
	if err != nil {
		t.Error(err)
	}
}

func TestNotifier_RefreshAccessToken(t *testing.T) {
	err := n.InitialAccessToken()
	if err != nil {
		t.Error(err)
	}

	err = n.RefreshAccessToken()
	if err != nil {
		t.Error(err)
	}
}

func TestNotifier_Send(t *testing.T) {
	err := n.InitialAccessToken()
	if err != nil {
		t.Error(err)
	}

	err = n.Send("110")
	if err != nil {
		t.Error(err)
	}
}
