package api

// import "time"

type InstanceSpecs struct {
	GPUs       int `json:"gpus" yaml:"GPUs"`
	MemoryGiB  int `json:"memory_gib" yaml:"MemoryGiB"`
	StorageGiB int `json:"storage_gib" yaml:"StorageGiB"`
	VCPUs      int `json:"vcpus" yaml:"VCPUs"`
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

type Instance struct {
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
	InstanceList []Instance `json:"data" yaml:"Instance"`
}

type InstanceOption struct {
	Region string       `yaml:"Region"`
	Type   InstanceType `yaml:"Type"`
}

type CreationData struct {
	Email  string `json:"email"`
	ID     string `json:"id"`
	Status string `json:"status"`
}

type Filesystem struct {
	BytesUsed int `json:"bytes_used"`
	// Created    time.Time    `json:"created"`
	Created    string       `json:"created"`
	CreatedBy  CreationData `json:"created_by"`
	ID         string       `json:"id"`
	IsInUse    bool         `json:"is_in_use"`
	MountPoint string       `json:"mount_point"`
	Name       string       `json:"name"`
	Region     Region       `json:"region"`
}

type FilesystemListResponse struct {
	FilesystemList []Filesystem `json:"data" yaml:"Filesystem"`
}

type SSHKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type SSHKeyListResponse struct {
	SSHKeyList []SSHKey `json:"data"`
}

type APIClient struct {
	BaseURL string
	APIKey  string
}
