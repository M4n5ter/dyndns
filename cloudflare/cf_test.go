package cloudflare_test

import (
	"testing"

	"github.com/m4n5ter/dyndns/cloudflare"
)

func TestNewApiUrl(t *testing.T) {
	api := cloudflare.NewApiUrl("123", "456")
	if api != "https://api.cloudflare.com/client/v4/zones/123/dns_records/456" {
		t.Errorf("api url error: %v", api)
	}

	api = cloudflare.NewApiUrl("123", "")
	if api != "https://api.cloudflare.com/client/v4/zones/123/dns_records" {
		t.Errorf("api url error: %v", api)
	}

}
