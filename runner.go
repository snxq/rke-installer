package rkeinstaller

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

type Runner interface {
	CommandLocal(command string, env []string) (string, error)
	CommandRemote(command string, env []string) (string, error)
	BatchCommandRemote(commands []string, env []string) ([]string, error)
	CopyRemote(localPath, remotePath string) error
}

type Option func(sshconf *ssh.ClientConfig) error

func WithPassword(password string) Option {
	return func(sshconf *ssh.ClientConfig) error {
		sshconf.Auth = append(sshconf.Auth, ssh.Password(password))
		return nil
	}
}

func WithKey(keyPath string) Option {
	return func(sshconf *ssh.ClientConfig) error {
		sshconf.Auth = append(sshconf.Auth, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			pem, err := os.ReadFile(keyPath)
			if err != nil {
				return nil, err
			}
			signer, err := ssh.ParsePrivateKey(pem)
			if err != nil {
				return nil, err
			}
			return []ssh.Signer{signer}, nil
		}))
		return nil
	}
}

func NewRunner(c Node, options ...Option) (Runner, error) {
	sshconf := &ssh.ClientConfig{
		User:            c.User,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	for _, o := range options {
		if err := o(sshconf); err != nil {
			return nil, err
		}
	}
	sshcli, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), sshconf)
	if err != nil {
		return nil, err
	}

	scpcli, err := scp.NewClientBySSH(sshcli)
	if err != nil {
		return nil, err
	}

	return &runner{
		sshcli: sshcli,
		scpcli: &scpcli,
	}, nil
}

type runner struct {
	sshcli *ssh.Client
	scpcli *scp.Client
}

func (r *runner) CommandLocal(command string, env []string) (string, error) {
	cmds := strings.Split(command, " ")
	if len(cmds) == 0 {
		return "", errors.New("empty command")
	}

	cmd := exec.Command(cmds[0], cmds[1:]...)
	if len(env) != 0 {
		cmd.Env = append(cmd.Env, env...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (r *runner) CommandRemote(command string, env []string) (string, error) {
	session, err := r.sshcli.NewSession()
	if err != nil {
		return "", err
	}

	for _, e := range env {
		kv := strings.Split(e, "=")
		if len(kv) != 2 {
			continue
		}
		if err := session.Setenv(kv[0], kv[1]); err != nil {
			return "", err
		}
	}

	out, err := session.CombinedOutput(command)
	fmt.Println(string(out))
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (r *runner) BatchCommandRemote(commands []string, env []string) ([]string, error) {
	result := make([]string, len(commands))
	for idx, command := range commands {
		out, err := r.CommandRemote(command, env)
		if err != nil {
			return result, err
		}

		result[idx] = out
	}
	return result, nil
}

func (r *runner) CopyRemote(localPath, remotePath string) error {
	return nil
}
