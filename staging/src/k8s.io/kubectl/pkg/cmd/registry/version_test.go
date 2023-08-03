package registry

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"reflect"
	"strings"
	"testing"
)

type fakeRegistry struct {
	wantDependenceRaw map[string]map[string]string
}

func (f fakeRegistry) GetTags(ctx context.Context, repo string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (f fakeRegistry) GetImageDependenceRaw(_ context.Context, name string) (map[string]string, error) {
	return f.wantDependenceRaw[name], nil
}

func (f fakeRegistry) GetImageDependence(ctx context.Context, image string) (map[string]*semver.Constraints, error) {
	//TODO implement me
	panic("implement me")
}

type fakeDynamicInformerInterface struct {
	wantGet map[string]*unstructured.Unstructured
}

func (f fakeDynamicInformerInterface) List(selector labels.Selector) (ret []runtime.Object, err error) {
	//TODO implement me
	panic("implement me")
}

func (f fakeDynamicInformerInterface) Get(name string) (runtime.Object, error) {
	return f.wantGet[name], nil
}

func (f fakeDynamicInformerInterface) ByNamespace(_ string) cache.GenericNamespaceLister { return f }

func (f fakeDynamicInformerInterface) DynamicInformer(gvr schema.GroupVersionResource) cache.GenericLister {
	return f
}

func (reg fakeRegistry) getVersionAndDependenceByUpdateRequest(ctx context.Context, req UpdateRequest) (string, map[string]string, error) {
	var version string
	deps := make(map[string]string)
	for _, c := range req.Containers {
		dependence, err := reg.GetImageDependenceRaw(ctx, c.GetImage())
		if err != nil {
			return "", nil, err
		}
		for k, v := range dependence {
			if got, ok := deps[k]; ok {
				// TODO: 重复的约束可以排重
				deps[k] = got + "," + v
			} else {
				deps[k] = v
			}
		}

		if version == "" {
			i := strings.LastIndexByte(c.Image, ':')
			if i == -1 {
				continue
			}
			v, err := semver.NewVersion(c.Image[i+1:])
			if err != nil {
				continue
			}
			version = fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
		}
	}
	return version, deps, nil
}

func TestService_getVersionAndDependenceByUpdateRequest(t *testing.T) {
	type fields struct {
		reg    Interface
		logger *zap.Logger
	}
	type args struct {
		ctx context.Context
		req UpdateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   map[string]string
		wantErr bool
	}{
		{
			name: "get",
			fields: fields{
				reg: fakeRegistry{wantDependenceRaw: map[string]map[string]string{
					"wecloud/wmc:notValid": {
						"cms": "~1.0.0",
						"rcs": "~1.1.0",
					},
					"wecloud/oam:1.2.1": {
						"security": "~1.2.0",
						"rcs":      "~1.3.0",
					},
					"wecloud/ext:1.2.2-beta": {
						"security": "~1.2.0",
					},
				}},
				logger: zap.NewNop(),
			},
			args: args{
				ctx: context.Background(),
				req: UpdateRequest{
					Containers: []Container{
						{Image: "wecloud/wmc:notValid"},
						{Image: "harbor:5000/wecloud/oam:1.2.1"},
						{Image: "harbor:5000/wecloud/ext:1.2.2-beta"},
					},
				},
			},
			want: "1.2.1",
			want1: map[string]string{
				"cms":      "~1.0.0",
				"rcs":      "~1.1.0,~1.3.0",
				"security": "~1.2.0,~1.2.0",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got1, err := tt.fields.reg.getVersionAndDependenceByUpdateRequest(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVersionAndDependenceByUpdateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVersionAndDependenceByUpdateRequest() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getVersionAndDependenceByUpdateRequest() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestService_checkForwardDependence(t *testing.T) {
	type fields struct {
		logger *zap.Logger
	}
	type args struct {
		ctx  context.Context
		objs map[string]*unstructured.Unstructured
		dep  map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				logger: zap.NewNop(),
			},
			args: args{
				ctx: context.Background(),
				objs: map[string]*unstructured.Unstructured{
					"oam": {Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								K8sLabelVersion: "1.1.2",
							},
						},
					}},
					"ext": {Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								K8sLabelVersion: "1.2.2",
							},
						},
					}},
				},
				dep: map[string]string{
					"oam": "~1.1.0",
					"ext": "=1.2.2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Registry{}
			if err := s.checkForwardDependence(tt.args.ctx, tt.args.objs, tt.args.dep); (err != nil) != tt.wantErr {
				t.Errorf("checkForwardDependence() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_checkReverseDependence(t *testing.T) {
	type fields struct{}
	type args struct {
		ctx     context.Context
		objs    map[string]*unstructured.Unstructured
		svc     string
		version string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				objs: map[string]*unstructured.Unstructured{
					"oam": {Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								K8sLabelVersion: "1.1.2",
							},
							"annotations": map[string]interface{}{
								"security" + K8sAnnotationDependence: "=1.2.3",
							},
						},
					}},
					"ext": {Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								K8sLabelVersion: "1.2.2",
							},
							"annotations": map[string]interface{}{
								"security" + K8sAnnotationDependence: "~1.2.1,~1.2.0",
							},
						},
					}},
				},
				svc:     "security",
				version: "1.2.3",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Registry{}
			if err := s.checkReverseDependence(tt.args.ctx, tt.args.objs, tt.args.svc, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("checkReverseDependence() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
