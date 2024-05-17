package rkeinstaller

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	type args struct {
		configs string
		path    string
	}
	tests := []struct {
		name    string
		args    args
		want    *Configs
		wantErr bool
	}{
		{
			name: "1_ok",
			args: args{
				configs: "nodes:\n  - name: k8s-server-01\n    host: 172.21.1.1\n    user: root\n    password: 123123\n    port: 22\n    role: master",
			},
			want: &Configs{
				Nodes: []*Node{
					{
						Name:     "k8s-server-01",
						Host:     "172.21.1.1",
						User:     "root",
						Password: "123123",
						Port:     22,
						Role:     "master",
					},
				},
			},
			wantErr: false,
		}, {
			name: "2_unmarshal_err",
			args: args{
				configs: "nodes",
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "3_reader_err",
			args: args{
				configs: "node",
				path:    "test.yaml",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := os.CreateTemp("", "*.yaml")
			defer os.Remove(f.Name())
			_, _ = f.Write([]byte(tt.args.configs))
			if tt.args.path == "" {
				tt.args.path = f.Name()
			}
			got, err := LoadConfigFromFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfigFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigs_Check(t *testing.T) {
	type fields struct {
		Nodes  []*Node
		Subnet string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "1_ok",
			fields: fields{
				Nodes: []*Node{
					{
						Name:     "k8s-server-01",
						Host:     "172.21.1.1",
						User:     "root",
						Password: "123123",
						Port:     22,
						Role:     "master",
					},
				},
				Subnet: "172.21.1.0/24",
			},
			wantErr: false,
		}, {
			name: "2_zero_nodes_err",
			fields: fields{
				Nodes: []*Node{},
			},
			wantErr: true,
		}, {
			name: "3_empety_node_name_err",
			fields: fields{
				Nodes: []*Node{
					{
						Name: "",
					},
				},
			},
			wantErr: true,
		}, {
			name: "4_empety_node_role_err",
			fields: fields{
				Nodes: []*Node{
					{
						Name: "xxx",
					},
				},
			},
			wantErr: true,
		}, {
			name: "5_no_master_err",
			fields: fields{
				Nodes: []*Node{
					{
						Name: "xxx",
						Role: "node",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Configs{
				Nodes:  tt.fields.Nodes,
				Subnet: tt.fields.Subnet,
			}
			if err := c.Check(); (err != nil) != tt.wantErr {
				t.Errorf("Configs.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
