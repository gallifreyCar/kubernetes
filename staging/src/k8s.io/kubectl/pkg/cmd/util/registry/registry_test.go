package registry

import (
	"reflect"
	"testing"
)

func TestGetImageDependenceRaw(t *testing.T) {
	if testing.Short() {
		t.Skip("skip get image dependence")
	}

	type args struct {
		image string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "domain",
			args:    args{image: "harbor:5000/wecloud/wmc:1.8.1"},
			want:    map[string]string{"ocm": "^2.0.0"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetImageDependenceRaw(tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageDependenceRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetImageDependenceRaw() got = %v, want %v", got, tt.want)
			}
		})
	}
}
