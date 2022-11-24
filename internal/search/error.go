package search

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func responseError(resp *esapi.Response) error {
	var em map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&em); err != nil {
		return err
	}
	return fmt.Errorf("[%s] %s: %s",
		resp.Status(),
		em["error"].(map[string]interface{})["type"],
		em["error"].(map[string]interface{})["reason"])
}
