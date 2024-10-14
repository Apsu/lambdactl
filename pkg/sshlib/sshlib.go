package sshlib

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type SSHClient struct {
	Client *ssh.Client
	PubKey ssh.PublicKey
}
type SSHSession struct {
	Session *ssh.Session
}

type SFTPClient struct {
	Client *sftp.Client
}

type SSHTarget struct {
	Host    string // IP or Hostname
	KeyName string // Default id_rsa
	Port    int    // Default 22
	User    string // Default ubuntu
}

// Return new client for target
func NewSSHClient(target SSHTarget) (*SSHClient, error) {
	// TODO: Wire in viper for user/port defaults
	if target.Port == 0 {
		target.Port = 22
	}
	if target.User == "" {
		target.User = "ubuntu"
	}

	// TODO: Use viper here -- privateKeyFile := os.ExpandEnv(viper.GetString("privateKey"))
	key, err := os.ReadFile(path.Join(os.ExpandEnv("$HOME/.ssh"), target.KeyName))
	if err != nil {
		return nil, fmt.Errorf("failed to open private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: target.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Maybe warn on this
		Timeout:         5 * time.Second,             // TODO: Make adjustble?
	}

	address := fmt.Sprintf("%s:%d", target.Host, target.Port)
	client, err := ssh.Dial("tcp", address, config)
	return &SSHClient{Client: client, PubKey: signer.PublicKey()}, err
}

// New session for connected client
func (c *SSHClient) NewSession() (*SSHSession, error) {
	session, err := c.Client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error creating new session: %v", err)
	}
	return &SSHSession{session}, nil
}

// Interactive shell on session
func (s *SSHSession) Shell() error {
	// Connect all the pipes
	s.Session.Stdin = os.Stdin
	s.Session.Stdout = os.Stdout
	s.Session.Stderr = os.Stderr

	// Make input raw to remove weirdness
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to make terminal raw: %v", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Grab current size and terminal profile
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %v", err)
	}

	term := os.Getenv("TERM")
	if term == "" {
		term = "xterm-256color"
	}

	// Ask for a matching new PTY
	if err = s.Session.RequestPty(term, h, w, ssh.TerminalModes{}); err != nil {
		return fmt.Errorf("failed to request PTY on remote s: %v", err)
	}

	// Start a login shell
	if err = s.Session.Shell(); err != nil {
		return fmt.Errorf("failed to launch remote shell: %v", err)
	}

	// Block until it returns
	return s.Session.Wait()
}

// All in one interactive shell
func NewShell(host string, port int, user string, keyName string) error {
	target := SSHTarget{
		Host:    host,
		KeyName: keyName,
		Port:    port,
		User:    user,
	}
	s, err := NewSSHClient(target)
	if err != nil {
		return err
	}
	se, err := s.NewSession()
	if err != nil {
		return err
	}
	return se.Shell()
}

// Run command in new session
func (c *SSHClient) Run(command string) error {
	s, err := c.NewSession()
	if err != nil {
		return fmt.Errorf("error creating new session: %v", err)
	}
	defer s.Session.Close()

	// Set up I/O for the command (stdout, stderr)
	s.Session.Stdout = os.Stdout
	s.Session.Stderr = os.Stderr

	// Run the command
	if err := s.Session.Run(command); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	return nil
}

// NewSFTPClient creates and returns an SFTP client from an existing SSH connection
func (c *SSHClient) NewSFTPClient() (*SFTPClient, error) {
	sftpClient, err := sftp.NewClient(c.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %v", err)
	}
	return &SFTPClient{sftpClient}, nil
}

// Mkdir creates a directory with the specified mode
func (s *SFTPClient) Mkdir(dir string, mode os.FileMode) error {
	// Create the directory
	if err := s.Client.MkdirAll(dir); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Set the directory permissions
	if err := s.Client.Chmod(dir, mode); err != nil {
		return fmt.Errorf("failed to set directory permissions: %v", err)
	}

	return nil
}

// Write to remote file and set specified mode
func (s *SFTPClient) WriteFile(source []byte, dest string, mode os.FileMode) error {
	// Create the file on the remote system
	remoteFile, err := s.Client.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Write content to the remote file
	if _, err := remoteFile.Write(source); err != nil {
		return fmt.Errorf("failed to write to remote file: %v", err)
	}

	// Set the file permissions
	if err := s.Client.Chmod(dest, mode); err != nil {
		return fmt.Errorf("failed to set file permissions: %v", err)
	}

	return nil
}
