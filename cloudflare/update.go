package cloudflare

import "time"

type UpdateRequestBody struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Proxied bool   `json:"proxied"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
	TTL     int    `json:"ttl"`
}

type UpdateResponse struct {
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   struct {
		Content   string    `json:"content"`
		Name      string    `json:"name"`
		Proxied   bool      `json:"proxied"`
		Type      string    `json:"type"`
		Comment   string    `json:"comment"`
		CreatedOn time.Time `json:"created_on"`
		Id        string    `json:"id"`
		Locked    bool      `json:"locked"`
		Meta      struct {
			AutoAdded bool   `json:"auto_added"`
			Source    string `json:"source"`
		} `json:"meta"`
		ModifiedOn time.Time `json:"modified_on"`
		Proxiable  bool      `json:"proxiable"`
		Tags       []string  `json:"tags"`
		Ttl        int       `json:"ttl"`
		ZoneId     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
	} `json:"result"`
	Success bool `json:"success"`
}
