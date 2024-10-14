package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"lambdactl/pkg/sshlib"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const (
	bootstrapRole = iota
	controllerRole
	workerRole
)

const (
	kubernetesType = "kubernetes"
	slurmType      = "slurm"
)

type RKE2TemplateData struct {
	ClusterIP string
	NodeName  string
	PublicIP  string
	Token     string
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().String("host", "", "Target host")
	deployCmd.Flags().Int("port", 22, "SSH port")
	deployCmd.Flags().String("user", "ubuntu", "SSH user")
	deployCmd.Flags().Bool("root", false, "Switch to root for deployment")
	deployCmd.Flags().String("role", "worker", "Node role")
	deployCmd.Flags().String("version", "", "Deployment version")
	deployCmd.MarkFlagRequired("host")
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy to an instance",
	Args:  cobra.ExactArgs(1), // deployment type
	Run: func(cmd *cobra.Command, args []string) {
		deploymentType := args[0] // e.g., 'kubernetes'

		// Flags
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		root, _ := cmd.Flags().GetBool("root")
		nodeRole, _ := cmd.Flags().GetString("role")
		deployVersion, _ := cmd.Flags().GetString("version")

		// Skip escalation step if already root
		if user == "root" {
			root = false
		}

		// Create SSHTarget based on user input
		target := sshlib.SSHTarget{
			Host:    host,
			KeyName: os.Getenv("SSH_KEY_NAME"), // You might load this from config (e.g., Viper)
			Port:    port,
			User:    user,
		}

		// Create SSH client
		client, err := sshlib.NewSSHClient(target)
		if err != nil {
			log.Fatalf("Failed to create SSH client: %v", err)
		}
		defer client.Client.Close()

		// Enable root access if requested
		if root {
			if err := enableRootAccess(client); err != nil {
				log.Fatalf("Failed to enable root access: %v", err)
			}

			// Reconnect as root
			target.User = "root"
			client, err = sshlib.NewSSHClient(target)
			if err != nil {
				log.Fatalf("Failed to reconnect as root: %v", err)
			}
			defer client.Client.Close()
		}

		// Create SFTP client
		sftpClient, err := client.NewSFTPClient()
		if err != nil {
			log.Fatalf("Failed to create SFTP client: %v", err)
		}
		defer sftpClient.Client.Close()

		switch strings.ToLower(deploymentType) {
		case kubernetesType:
			// Part 1: Prepare the machine
			if err := prepareMachine(client); err != nil {
				log.Fatalf("Failed to prepare machine: %v", err)
			}

			// Part 2: Set up RKE2 (Kubernetes)
			if err := deployKubernetes(client, sftpClient, target.Host, nodeRole, deployVersion); err != nil {
				log.Fatalf("Failed to deploy Kubernetes: %v", err)
			}
		}

		log.Info("Deployment completed successfully!")
	},
}

func enableRootAccess(c *sshlib.SSHClient) error {
	log.Info("Enabling root access by generating authorized_keys...")

	cmd := "sudo mkdir -p /root/.ssh && sudo cp /home/ubuntu/.ssh/authorized_keys /root/.ssh/authorized_keys && sudo chmod 600 /root/.ssh/authorized_keys"
	if err := c.Run(cmd); err != nil {
		return fmt.Errorf("failed to enable root access: %v", err)
	}

	log.Info("Root access enabled.")
	return nil
}

// Part 1: Prepare the machine (remove packages, stop services, install tools)
func prepareMachine(c *sshlib.SSHClient) error {
	log.Info("Preparing the machine...")

	// Steps
	commands := []string{
		"DEBIAN_FRONTEND=noninteractive apt-get autoremove --purge -y '~i !~OUbuntu'", // Remove 3rd-party packages
		// "systemctl stop unwanted-service",           // Stop service
		"DEBIAN_FRONTEND=noninteractive apt-get update",     // Update repos
		"DEBIAN_FRONTEND=noninteractive apt-get upgrade -y", // Upgrade machine
	}

	for _, cmd := range commands {

		log.Debugf("Running: %s\n", cmd)
		if err := c.Run(cmd); err != nil {
			return fmt.Errorf("failed to run command '%s': %v", cmd, err)
		}
	}

	return nil
}

// Part 2: Set up RKE2 (Kubernetes)
func deployKubernetes(c *sshlib.SSHClient, s *sshlib.SFTPClient, publicIP string, nodeRole string, deployVersion string) error {
	log.Info("Setting up RKE2 (Kubernetes)...")

	var rke2Role string
	switch nodeRole {
	case "bootstrap", "controller":
		rke2Role = "server"
	case "worker":
		rke2Role = "agent"
	default:
		return fmt.Errorf("Invalid node role specified: %v", nodeRole)
	}

	// Step 1: Download and install RKE2 installer
	installerCmd := fmt.Sprintf("curl -sfL https://get.rke2.io | INSTALL_RKE2_TYPE=%s INSTALL_RKE2_VERSION=%s sh -", rke2Role, deployVersion)
	if err := c.Run(installerCmd); err != nil {
		return fmt.Errorf("failed to install RKE2: %v", err)
	}

	// Step 2: Create config directories and upload rendered templates
	configDir := "/etc/rancher/rke2/"
	manifestDir := "/var/lib/rancher/rke2/"

	// Ensure the directory exists
	if err := s.Mkdir(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	if err := s.Mkdir(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %v", err)
	}

	templateData := RKE2TemplateData{
		NodeName: "test-" + rke2Role + "-1",
		PublicIP: publicIP,
		Token:    "Test1234",
	}

	// Step 3: Render and upload templates (e.g., config.yaml)
	templates := map[string]string{
		configDir + "config.yaml":                "deploy/configs/" + nodeRole + "-config.yaml",
		manifestDir + "rke2-cilium-values.yaml":  "deploy/manifests/rke2-cilium-values.yaml",
		manifestDir + "rke2-coredns-values.yaml": "deploy/manifests/rke2-coredns-values.yaml",
		// Add more template paths as needed
	}

	for remoteFile, templatePath := range templates {
		// Read the raw template from embedded FS
		templateContent, err := lambdaFS.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %v", templatePath, err)
		}

		// Parse the template
		tmpl, err := template.New(remoteFile).Parse(string(templateContent))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %v", templatePath, err)
		}

		// Render the template with the provided data
		var renderedTemplate bytes.Buffer
		if err := tmpl.Execute(&renderedTemplate, templateData); err != nil {
			return fmt.Errorf("failed to render template %s: %v", templatePath, err)
		}

		// Upload the rendered template to the remote machine
		log.Printf("Rendering and uploading template to %s\n", remoteFile)

		if err := s.WriteFile(renderedTemplate.Bytes(), remoteFile, 0644); err != nil {
			return fmt.Errorf("failed to upload rendered template %s: %v", remoteFile, err)
		}
	}

	// Step 4: Start the RKE2 service
	startServiceCmd := "systemctl enable --now --no-block rke2-server"
	if err := c.Run(startServiceCmd); err != nil {
		return fmt.Errorf("failed to start RKE2: %v", err)
	}

	return nil
}
