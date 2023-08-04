package registry

import (
	"context"
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"os"
	"strings"
)

type Interface interface {
	// GetTags 根据服务名称查询标签
	// 比如镜像: harbor:5000/wecloud/wmc:1.5.1
	// repo: wecloud/wmc
	// tag: 1.5.1
	GetTags(ctx context.Context, repo string) ([]string, error)
	// GetImageDependenceRaw 获取镜像的依赖约束(未解析)
	GetImageDependenceRaw(ctx context.Context, image string) (map[string]string, error)
	// GetImageDependence 获取镜像的依赖约束
	GetImageDependence(ctx context.Context, image string) (map[string]*semver.Constraints, error)

	getVersionAndDependenceByUpdateRequest(ctx context.Context, req UpdateRequest) (string, map[string]string, error)
}

type Registry struct {
	// 镜像仓库地址和端口
	// 必须是可以访问到的地址或域名
	Address string `json:"address"`
	// 镜像仓库用户名
	User string `json:"user"`
	// 镜像仓库密码
	Password string `json:"password"`
	// 忽略不安全的HTTPS
	Insecure bool `json:"insecure" default:"true"`
	// 镜像拉取密钥
	PullSecret string `json:"pull_secret"`
}

func (reg *Registry) getAuth() (authn.Authenticator, error) {
	if reg.User != "" && reg.Password != "" {
		return authn.FromConfig(authn.AuthConfig{Username: reg.User, Password: reg.Password}), nil
	}

	if _, err := os.Stat(".docker/config.json"); err == nil {
		reg, err := name.NewRegistry(reg.Address)
		if err != nil {
			return nil, err
		}
		return authn.DefaultKeychain.Resolve(reg)
	}
	return nil, nil
}

func (reg *Registry) GetTags(ctx context.Context, service string) ([]string, error) {
	repo, err := name.NewRepository(reg.Address+"/"+service, reg.getNameOptions()...)
	if err != nil {
		return nil, err
	}
	auth, err := reg.getAuth()
	if err != nil {
		return nil, err
	}
	got, err := remote.List(repo, remote.WithAuth(auth))
	if err != nil {
		return nil, err
	}
	return got, err
}

func (reg *Registry) GetImageDependenceRaw(image string) (map[string]string, error) {
	ref, err := name.ParseReference(reg.Address+"/"+image, reg.getNameOptions()...)
	if err != nil {
		return nil, err
	}
	auth, err := reg.getAuth()
	if err != nil {
		return nil, err
	}
	desc, err := remote.Get(ref, remote.WithAuth(auth))
	if err != nil {
		return nil, err
	}

	images, err := desc.Image()
	if err != nil {
		return nil, err
	}
	cfg, err := images.ConfigFile()
	if err != nil {
		return nil, err
	}

	results := make(map[string]string, len(cfg.Config.Labels))
	for k, v := range cfg.Config.Labels {
		if len(k) <= 4 || !strings.HasPrefix(k, "ver_") {
			continue
		}
		results[k[4:]] = v
	}
	return results, nil
}

func (reg *Registry) getNameOptions() []name.Option {
	var opts []name.Option
	if reg.Insecure {
		opts = append(opts, name.Insecure)
	}
	return opts
}
