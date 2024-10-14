package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"

	"slices"
	"strings"
	"time"

	"lambdactl/pkg/utils"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type InstanceSpecs struct {
	Bus        string `yaml:"Bus"`
	GPUs       int    `json:"gpus" yaml:"GPUs"`
	MemoryGiB  int    `json:"memory_gib" yaml:"MemoryGiB"`
	Model      string `yaml:"Model"`
	StorageGiB int    `json:"storage_gib" yaml:"StorageGiB"`
	VCPUs      int    `json:"vcpus" yaml:"VCPUs"`
}

type InstanceType struct {
	Description       string        `json:"description" yaml:"Description"`
	GPUDescription    string        `json:"gpu_description" yaml:"GPUDescription"`
	Name              string        `json:"name" yaml:"Name"`
	PriceCentsPerHour int           `json:"price_cents_per_hour" yaml:"PriceCentsPerHour"`
	Specs             InstanceSpecs `json:"specs" yaml:"Specs"`
}

type Region struct {
	Description string `json:"description" yaml:"Description"`
	Name        string `json:"name" yaml:"Name"`
}

type InstanceData struct {
	InstanceType     InstanceType `json:"instance_type" yaml:"InstanceType"`
	RegionsAvailable []Region     `json:"regions_with_capacity_available" yaml:"RegionsAvailable"`
}

type InstanceTypesResponse struct {
	InstanceTypes map[string]InstanceData `json:"data" yaml:"InstanceTypes"`
}

type InstanceLaunchData struct {
	InstanceIDs []string `json:"instance_ids" yaml:"InstanceIDs"`
}

type InstanceLaunchResponse struct {
	InstanceLaunches InstanceLaunchData `json:"data" yaml:"InstanceLaunches"`
}

type InstanceDetails struct {
	Filesystems  []string     `json:"file_system_names" yaml:"Filesystems"`
	Hostname     string       `json:"hostname" yaml:"Hostname"`
	ID           string       `json:"id" yaml:"ID"`
	InstanceType InstanceType `json:"instance_type" yaml:"InstanceType"`
	IP           string       `json:"ip" yaml:"IP"`
	IsReserved   bool         `json:"is_reserved" yaml:"IsReserved"`
	Name         string       `json:"name" yaml:"Name"`
	PrivateIP    string       `json:"private_ip" yaml:"PrivateIP"`
	Region       Region       `json:"region" yaml:"Region"`
	SSHKeys      []string     `json:"ssh_key_names" yaml:"SSHKeys"`
	Status       string       `json:"status" yaml:"Status"`
}

type InstanceListResponse struct {
	InstanceList []InstanceDetails `json:"data" yaml:"InstanceList"`
}

type InstanceOption struct {
	PriceHour int           `yaml:"PriceHour"`
	Region    string        `yaml:"Region"`
	Specs     InstanceSpecs `yaml:"Specs"`
	Type      string        `yaml:"Type"`
}

type APIClient struct {
	BaseURL string
	APIKey  string
}

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
			instanceSpecs, err := ParseInstanceType(data.InstanceType)
			if err != nil {
				continue
			}

			instanceOption := InstanceOption{
				PriceHour: data.InstanceType.PriceCentsPerHour,
				Region:    region.Name,
				Specs:     instanceSpecs,
				Type:      data.InstanceType.Name,
			}
			instanceOptions = append(instanceOptions, instanceOption)
		}
	}

	return instanceOptions, nil
}

func ParseInstanceType(input InstanceType) (InstanceSpecs, error) {
	specs := input.Specs
	specs.Bus = "pcie" // Default

	fields := strings.Split(input.Name, "_")
	for i, field := range fields {
		switch i {
		case 0:
			if field == "cpu" {
				specs.Model = "cpu"
			}
		case 2:
			specs.Model = field
		case 3:
			specs.Bus = field
		}
	}

	return specs, nil
}

func ParseOptionType(input string) (InstanceSpecs, error) {
	specs := InstanceSpecs{
		Bus: "pcie", // Default
	}

	fields := strings.Split(input, "_")
	for i, field := range fields {
		switch i {
		case 0:
			if field == "cpu" {
				specs.Model = "cpu"
			}
		case 2:
			specs.Model = field
		case 3:
			specs.Bus = field
		}
	}

	return specs, nil
}

// func ParseGPUString(input string) (GPUSpec, error) {
// 	re := regexp.MustCompile(`(\d+)x\s+(\w+)\s+\((\d+)\s+GB(?:\s+(\w+))?\)`)
// 	matches := re.FindStringSubmatch(input)

// 	if len(matches) < 4 {
// 		return GPUSpec{}, fmt.Errorf("invalid GPU string format")
// 	}

// 	ram, err := strconv.Atoi(matches[3])
// 	if err != nil {
// 		return GPUSpec{}, fmt.Errorf("invalid RAM: %v", err)
// 	}

// 	bus := "PCIe" // Default to PCIe if not specified
// 	if len(matches) > 4 && matches[4] != "" {
// 		bus = matches[4]
// 	}

// 	return GPUSpec{
// 		Model: matches[2],
// 		RAM:   ram,
// 		Bus:   bus,
// 	}, nil
// }

func SelectBestInstanceOption(options []InstanceOption, requested InstanceOption) (InstanceOption, error) {
	var bestOption InstanceOption
	lowestCost := math.MaxInt

	for _, option := range options {
		// Check if the option meets the minimum requirements
		// if option.Spec.RAM < requested.Spec.RAM || option.Count < requested.Count {
		// 	continue
		// }

		// Check model and bus only if they are specified
		if requested.Specs.Model != "" && option.Specs.Model != requested.Specs.Model {
			continue
		}

		if requested.Specs.Bus != "" && option.Specs.Bus != requested.Specs.Bus {
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

	if bestOption.Type == "" {
		return InstanceOption{}, errors.New("no suitable GPU option found")
	}

	return bestOption, nil
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
		"instance_type_name": instanceOption.Type,
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

	var allInstances InstanceListResponse
	err = json.Unmarshal(resp, &allInstances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	return allInstances.InstanceList, nil
}

func (c *APIClient) SSHIntoMachine(instance InstanceDetails) error {
	privateKeyFile := os.ExpandEnv(viper.GetString("privateKey"))
	if privateKeyFile == "" {
		privateKeyFile = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}

	privateKey, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	// TODO: Get username from config or use default
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", instance.IP), config)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %v", err)
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}

	if err = session.RequestPty(term, h, w, ssh.TerminalModes{}); err != nil {
		return fmt.Errorf("failed to request PTY on remote session: %v", err)
	}

	if err = session.Shell(); err != nil {
		return fmt.Errorf("failed to launch remote shell: %v", err)
	}

	return session.Wait()
}
