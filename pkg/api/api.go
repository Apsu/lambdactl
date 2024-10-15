package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"

	"slices"
	"time"

	"lambdactl/pkg/utils"

	"github.com/spf13/viper"
)

func NewAPIClient(baseURL, apiKey string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
}

func (c *APIClient) MakeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *APIClient) FetchInstanceOptions() ([]InstanceOption, error) {
	resp, err := c.MakeRequest("GET", "instance-types", nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving instance types: %v", err)
	}

	var instanceTypes InstanceTypesResponse
	err = json.Unmarshal(resp, &instanceTypes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response data: %v", err)
	}

	instanceOptions := []InstanceOption{}
	for _, data := range instanceTypes.InstanceTypes {
		for _, region := range data.RegionsAvailable {
			instanceOption := InstanceOption{
				PriceHour: data.InstanceType.PriceCentsPerHour,
				Region:    region.Name,
				Type:      data.InstanceType,
			}
			instanceOptions = append(instanceOptions, instanceOption)
		}
	}

	return instanceOptions, nil
}

func (c *APIClient) LaunchInstances(instanceOption InstanceOption, quantity int) (InstanceLaunchData, error) {
	var sshKeyNames []string
	if err := viper.UnmarshalKey("sshKeyNames", &sshKeyNames); err != nil || sshKeyNames == nil {
		sshKeyNames = []string{"AAP"} // Fallback key
	}
	if quantity < 1 {
		quantity = 1
	}

	data := map[string]interface{}{
		"region_name":        instanceOption.Region,
		"instance_type_name": instanceOption.Type.Name,
		"ssh_key_names":      sshKeyNames,
		"quantity":           quantity,
	}

	resp, err := c.MakeRequest("POST", "instance-operations/launch", data)
	if err != nil {
		return InstanceLaunchData{}, fmt.Errorf("error launching instance(s): %v", err)
	}

	var launchResponse InstanceLaunchResponse
	err = json.Unmarshal(resp, &launchResponse)
	if err != nil {
		return InstanceLaunchData{}, fmt.Errorf("error unmarshaling response data: %v", err)
	}

	return launchResponse.InstanceLaunches, nil
}

func (c *APIClient) WaitForInstances(instancesLaunched InstanceLaunchData) (map[string]InstanceDetails, error) {
	var myInstances = map[string]InstanceDetails{}
	for {
		resp, err := c.MakeRequest("GET", "instances", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get instance details: %v", err)
		}

		var allInstances InstanceListResponse
		err = json.Unmarshal(resp, &allInstances)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal API response: %v", err)
		}

		for _, instance := range allInstances.InstanceList {
			if slices.Contains(instancesLaunched.InstanceIDs, instance.ID) {
				// One of mine
				myInstances[instance.ID] = instance
			}
		}

		if utils.All(myInstances, func(v InstanceDetails) bool { return v.Status == "active" && v.IP != "" }) {
			return myInstances, nil
		}

		time.Sleep(10 * time.Second)
	}
}

func (c *APIClient) ListInstances() ([]InstanceDetails, error) {
	resp, err := c.MakeRequest("GET", "instances", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance details: %v", err)
	}

	var listResponse InstanceListResponse
	err = json.Unmarshal(resp, &listResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	return listResponse.InstanceList, nil
}

func SelectBestInstanceOption(options []InstanceOption, requested InstanceOption) (InstanceOption, error) {
	var bestOption InstanceOption
	lowestCost := math.MaxInt

	// Check if the options meet the minimum requirements
	for _, option := range options {
		// Check model and bus only if they are specified
		if requested.Type.Name != "" && option.Type.Name != requested.Type.Name {
			continue
		}

		// Check region if specified
		if requested.Region != "" && option.Region != requested.Region {
			continue
		}

		// If we've made it here, the option is valid. Check if it's the best so far.
		if option.PriceHour < lowestCost {
			lowestCost = option.PriceHour
			bestOption = option
		}
	}

	if bestOption.Type.Name == "" {
		return InstanceOption{}, errors.New("no suitable GPU option found")
	}

	return bestOption, nil
}
