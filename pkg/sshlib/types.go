package sshlib

import (
	"io"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
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

type SSHExecCommand struct {
	Target SSHTarget
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type SSHTarget struct {
	Host    string // IP or Hostname
	KeyName string // Default id_rsa
	Port    int    // Default 22
	User    string // Default ubuntu
}

type SSHStreams struct {
	Stdin int
}
