package wise

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func jsonDecode(resp *http.Response, target any) error {
	return json.NewDecoder(resp.Body).Decode(target)
}

func jsonUnmarshal(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func readBody(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)

	return string(body), err
}

func parseWiseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
