package rkeinstaller

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Configs struct {
	Nodes []*Node

	Subnet string
}

type Node struct {
	Name     string
	Host     string
	User     string
	Password string
	KeyPath  string
	Port     int
	Role     string
}

func LoadConfigFromFile(path string) (*Configs, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Configs{}
	if err = yaml.Unmarshal(f, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Configs) Check() error {
	if len(c.Nodes) == 0 {
		return errors.New("节点数目为零, 无需部署:)")
	}

	for _, node := range c.Nodes {
		if node.Name == "" {
			return errors.New("节点名称不能为空")
		}
		if node.Role == "" {
			return errors.New("节点角色不能为空")
		}
	}
	if c.GetMaster() == nil {
		return errors.New("没有 master 节点")
	}

	return nil
}

func (c *Configs) GetMaster() *Node {
	for _, node := range c.Nodes {
		if node.Role == "master" {
			return node
		}
	}
	return nil
}
