package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/seelentov/atlassian-client/models"
)

var (
	ErrRequest = errors.New("request failed")
)

type AtlassianClient struct {
	username string
	token    string
	url      string
}

func NewAtlassianClient(company string, username string, token string) *AtlassianClient {
	return &AtlassianClient{
		username,
		token,
		"https://" + company + ".atlassian.net/wiki",
	}
}

func (c *AtlassianClient) doReq(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", c.url+url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, fmt.Errorf("%w:%d %s", ErrRequest, resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *AtlassianClient) GetPage(id int, withContent bool) (*models.Page, error) {
	url := fmt.Sprintf("/rest/api/content/%d", id)

	if withContent {
		url += "?expand=body.storage"
	}

	body, err := c.doReq(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Title string `json:"title"`
		Body  struct {
			Storage struct {
				Value string `json:"value"`
			} `json:"storage"`
		} `json:"body"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	page := &models.Page{
		Id:      id,
		Title:   result.Title,
		Content: "",
	}

	if withContent {
		page.Content = result.Body.Storage.Value
	}

	return page, nil
}

func (c *AtlassianClient) GetChildrenIds(id int) ([]int, error) {
	url := fmt.Sprintf("/api/v2/pages/%d/children", id)

	body, err := c.doReq(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []struct {
			Id string `json:"id"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	childrenIds := []int{}

	for _, el := range result.Results {
		id, err := strconv.Atoi(el.Id)
		if err != nil {
			return nil, err
		}

		childrenIds = append(childrenIds, id)
	}

	return childrenIds, nil
}
