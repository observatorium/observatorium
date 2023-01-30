package api

import (
	"encoding/json"
	"time"

	"github.com/efficientgo/core/errors"

	obsrbac "github.com/observatorium/api/rbac"
)

// Struct is unexported from observatorium/api, so adding needed fields here.
type Tenants struct {
	Tenants []*Tenant `json:"tenants"`
}

type Tenant struct {
	Name       string        `json:"name"`
	ID         string        `json:"id"`
	OIDC       *OIDC         `json:"oidc"`
	RateLimits []*Ratelimits `json:"rateLimits"`
}

type OIDC struct {
	ClientID  string `json:"clientID"`
	IssuerURL string `json:"issuerURL"`
}

type Ratelimits struct {
	Endpoint string   `json:"endpoint"`
	Limit    int      `json:"limit"`
	Window   Duration `json:"window"`
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))

		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}

		*d = Duration(tmp)

		return nil
	default:
		return errors.New("invalid duration")
	}
}

type RBAC struct {
	Roles        []obsrbac.Role        `json:"roles"`
	RoleBindings []obsrbac.RoleBinding `json:"roleBindings"`
}
