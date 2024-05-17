package rkeinstaller

import (
	"fmt"
	"log"
	"strings"
)

type Installer interface {
	Start() error
}

type installer struct {
	conf *Configs

	token   string
	master  string
	servers []string
}

func NewInstaller(conf *Configs) Installer {
	i := &installer{
		conf:    conf,
		servers: []string{},
	}
	for _, node := range conf.Nodes {
		if node.Role == "master" || node.Role == "server" {
			i.servers = append(i.servers, node.Host)
		}
	}
	return i
}

func (i *installer) Start() error {
	log.Println("start deploy...")
	log.Printf("start deploy master: %v\n", i.conf.GetMaster().Host)
	if err := i.DeployNode(i.conf.GetMaster()); err != nil {
		return err
	}
	log.Printf("deploy master success: %v, Got token: %v\n", i.conf.GetMaster().Host, i.token)

	for _, node := range i.conf.Nodes {
		if node.Role == "master" {
			continue
		}
		log.Printf("start deploy %s: %v\n", node.Role, node.Host)
		if err := i.DeployNode(node); err != nil {
			return err
		}
	}

	fmt.Println(i.servers)
	return nil
}

func (i *installer) tlsSan() string {
	servers := []string{}
	for _, server := range i.servers {
		servers = append(servers, fmt.Sprintf("- %v", server))
	}
	return strings.Join(servers, "\n")
}

func (i *installer) DeployNode(node *Node) error {
	options := []Option{}
	if node.Password != "" {
		options = append(options, WithPassword(node.Password))
	}
	if node.KeyPath != "" {
		options = append(options, WithKey(node.KeyPath))
	}
	runner, err := NewRunner(*node, options...)
	if err != nil {
		return err
	}

	switch node.Role {
	case "master":
		outs, err := runner.BatchCommandRemote([]string{
			"curl -sfL https://rancher-mirror.rancher.cn/rke2/install.sh | INSTALL_RKE2_MIRROR=cn sh -",
			"systemctl enable rke2-server.service",
			"mkdir -p /etc/rancher/rke2",
			fmt.Sprintf("echo 'node-ip: %s\ntls-san:\n%s' > /etc/rancher/rke2/config.yaml", node.Host, i.tlsSan()),
			"systemctl start rke2-server.service",
			"cat /var/lib/rancher/rke2/server/node-token",
		}, nil)
		if err != nil {
			return err
		}
		i.token = outs[len(outs)-1]
		i.master = node.Host
	case "server":
		fallthrough
	case "agent":
		_, err := runner.BatchCommandRemote([]string{
			"curl -sfL https://rancher-mirror.rancher.cn/rke2/install.sh | INSTALL_RKE2_MIRROR=cn sh -",
			"mkdir -p /etc/rancher/rke2/",
			fmt.Sprintf(
				"echo 'server: https://%s:9345\ntoken: %s\ntls-san:\n%s' > /etc/rancher/rke2/config.yaml",
				i.master, i.token, i.tlsSan()),
			fmt.Sprintf("systemctl enable rke2-%s.service", node.Role),
			fmt.Sprintf("systemctl start rke2-%s.service", node.Role),
		}, nil)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s role not support", node.Role)
	}

	return nil
}
