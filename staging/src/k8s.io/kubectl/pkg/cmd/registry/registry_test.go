package registry

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestRegistry_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("skip registry query")
	}

	type fields struct {
		Address    string
		User       string
		Password   string
		Insecure   bool
		PullSecret string
	}
	type args struct {
		service string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "registry",
			fields: fields{
				Address:  "192.168.40.96:5000",
				User:     "admin",
				Password: "Harbor12345",
			},
			args:    args{service: "wecloud/wmc"},
			want:    nil,
			wantErr: false,
		},
		{
			name: "jdc",
			fields: fields{
				Address: "wellcloud-cn-east-2.jcr.service.jdcloud.com",
			},
			args:    args{service: "wmc"},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Registry{
				Address:    tt.fields.Address,
				User:       tt.fields.User,
				Password:   tt.fields.Password,
				Insecure:   tt.fields.Insecure,
				PullSecret: tt.fields.PullSecret,
			}
			got, err := r.GetTags(context.Background(), tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Query() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestRegistry_GetImageDependenceRaw(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skip get image dependence raw")
//	}
//
//	type fields struct {
//		Address    string
//		User       string
//		Password   string
//		Insecure   bool
//		PullSecret string
//	}
//	type args struct {
//		ctx   context.Context
//		image string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    map[string]string
//		wantErr bool
//	}{
//		{
//			name: "jdc",
//			fields: fields{
//				Address: "wellcloud-cn-east-2.jcr.service.jdcloud.com",
//			},
//			args: args{
//				ctx:   context.Background(),
//				image: "wmc:1.6.2",
//			},
//			want:    nil,
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := &Registry{
//				Address:    tt.fields.Address,
//				User:       tt.fields.User,
//				Password:   tt.fields.Password,
//				Insecure:   tt.fields.Insecure,
//				PullSecret: tt.fields.PullSecret,
//			}
//			got, err := r.GetImageDependenceRaw(tt.args.ctx, tt.args.image)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetImageDependenceRaw() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetImageDependenceRaw() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func Test_ParseConstraints(t *testing.T) {
	type args struct {
		ctx    context.Context
		labels map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "parse",
			args: args{
				ctx: context.Background(),
				labels: map[string]string{
					"oam": "~1.2.3",
					"ocm": "~1.4.3 || 1.1.0",
				},
			},
			want:    "map[oam:~1.2.3 ocm:~1.4.3 || 1.1.0]",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConstraints(tt.args.ctx, tt.args.labels)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConstraints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if fmt.Sprint(got) != tt.want {
				t.Errorf("parseConstraints() got = %s, want %s", fmt.Sprint(got), tt.want)
			}
		})
	}
}
