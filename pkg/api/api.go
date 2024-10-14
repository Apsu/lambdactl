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
	"regexp"
	"slices"
	"strconv"
	"time"

	"lambdactl/pkg/utils"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type InstanceSpecs struct {
	VCPUs      int `json:"vcpus" yaml:"VCPUs"`
	MemoryGiB  int `json:"memory_gib" yaml:"MemoryGiB"`
	StorageGiB int `json:"storage_gib" yaml:"StorageGiB"`
	GPUs       int `json:"gpus" yaml:"GPUs"`
}

type InstanceType struct {
	Name              string        `json:"name" yaml:"Name"`
	Description       string        `json:"description" yaml:"Description"`
	GPUDescription    string        `json:"gpu_description" yaml:"GPUDescription"`
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
	Name         string       `json:"name" yaml:"Name"`
	Hostname     string       `json:"hostname" yaml:"Hostname"`
	ID           string       `json:"id" yaml:"ID"`
	InstanceType InstanceType `json:"instance_type" yaml:"InstanceType"`
	IP           string       `json:"ip" yaml:"IP"`
	IsReserved   bool         `json:"is_reserved" yaml:"IsReserved"`
	PrivateIP    string       `json:"private_ip" yaml:"PrivateIP"`
	Region       Region       `json:"region" yaml:"Region"`
	SSHKeys      []string     `json:"ssh_key_names" yaml:"SSHKeys"`
	Status       string       `json:"status" yaml:"Status"`
}

type InstanceListResponse struct {
	InstanceList []InstanceDetails `json:"data" yaml:"InstanceList"`
}

type GPUSpec struct {
	Model string `yaml:"Model"`
	RAM   int    `yaml:"RAM"`
	Bus   string `yaml:"Bus"`
}

type GPUOption struct {
	Spec      GPUSpec `yaml:"Spec"`
	Type      string  `yaml:"Type"`
	Count     int     `yaml:"Count"`
	PriceHour int     `yaml:"PriceHour"`
	Region    string  `yaml:"Region"`
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

func (c *APIClient) FetchGPUOptions() ([]GPUOption, error) {
	resp, err := c.MakeRequest("GET", "instance-types", nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving instance types: %v", err)
	}

	var instanceTypes InstanceTypesResponse
	err = json.Unmarshal(resp, &instanceTypes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response data: %v", err)
	}

	gpuOptions := []GPUOption{}
	for _, data := range instanceTypes.InstanceTypes {
		for _, region := range data.RegionsAvailable {
			gpuSpec, err := parseGPUString(data.InstanceType.Description)
			if err != nil {
				continue
			}
			gpuOptions = append(gpuOptions, GPUOption{
				Spec:      gpuSpec,
				Type:      data.InstanceType.Name,
				Count:     data.InstanceType.Specs.GPUs,
				PriceHour: data.InstanceType.PriceCentsPerHour,
				Region:    region.Name,
			})
		}
	}

	return gpuOptions, nil
}

func parseGPUString(input string) (GPUSpec, error) {
	re := regexp.MustCompile(`(\d+)x\s+(\w+)\s+\((\d+)\s+GB(?:\s+(\w+))?\)`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 4 {
		return GPUSpec{}, fmt.Errorf("invalid GPU string format")
	}

	ram, err := strconv.Atoi(matches[3])
	if err != nil {
		return GPUSpec{}, fmt.Errorf("invalid RAM: %v", err)
	}

	bus := "PCIe" // Default to PCIe if not specified
	if len(matches) > 4 && matches[4] != "" {
		bus = matches[4]
	}

	return GPUSpec{
		Model: matches[2],
		RAM:   ram,
		Bus:   bus,
	}, nil
}

func SelectBestGPUOption(options []GPUOption, requested GPUOption) (GPUOption, error) {
	var bestOption GPUOption
	lowestCost := math.MaxInt

	for _, option := range options {
		// Check if the option meets the minimum requirements
		if option.Spec.RAM < requested.Spec.RAM || option.Count < requested.Count {
			continue
		}

		// Check model and bus only if they are specified
		if requested.Spec.Model != "" && option.Spec.Model != requested.Spec.Model {
			continue
		}

		if requested.Spec.Bus != "" && option.Spec.Bus != requested.Spec.Bus {
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
		return GPUOption{}, errors.New("no suitable GPU option found")
	}

	return bestOption, nil
}

func (c *APIClient) LaunchInstances(gpuOption GPUOption, quantity int) (InstanceLaunchData, error) {
	var sshKeyNames []string
	if err := viper.UnmarshalKey("sshKeyNames", &sshKeyNames); err != nil || sshKeyNames == nil {
		sshKeyNames = []string{"AAP"} // Fallback key
	}
	if quantity < 1 {
		quantity = 1
	}

	data := map[string]interface{}{
		"region_name":        gpuOption.Region,
		"instance_type_name": gpuOption.Type,
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

// Poll the API to check if the instance(s) are ready (active)
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

		time.Sleep(10 * time.Second) // Poll every 10 seconds
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
	// Get private key from viper config or fallback to default
	privateKeyFile := os.ExpandEnv(viper.GetString("privateKey"))
	if privateKeyFile == "" {
		privateKeyFile = os.ExpandEnv("$HOME/.ssh/id_rsa")
	}

	// Read private key file
	privateKey, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	// Parse the private key
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	// Build SSH client configuration
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

	// Handle SSH session interaction
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	if err = session.RequestPty(term, h, w, modes); err != nil {
		return fmt.Errorf("failed to request PTY on remote session: %v", err)
	}

	if err = session.Shell(); err != nil {
		return fmt.Errorf("failed to launch remote shell: %v", err)
	}

	return session.Wait()
}
