package registry

import (
	corev1 "k8s.io/api/core/v1"
	"strings"
)

type ImageType int

const (
	ImageTypeInit   ImageType = 1 + iota // 初始化容器镜像
	ImageTypeNormal                      // 普通容器镜像
)

type Container struct {
	Name string    `json:"name" binding:"required"`
	Type ImageType `json:"type" binding:"gt=0"`
	// 镜像版本
	Image string `json:"image" binding:"required"`
	// 环境变量
	Env []Env `json:"env"`
}

//	func (m Container) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//		enc.AddString("name", m.Name)
//		enc.AddString("image", m.Image)
//		return nil
//	}
func (m Container) GetImage() string {
	base := m.Image
	var tag string
	parts := strings.Split(m.Image, ":")
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
		base = strings.Join(parts[:len(parts)-1], ":")
		tag = parts[len(parts)-1]
	}

	repo := base
	parts = strings.SplitN(base, "/", 2)
	if len(parts) == 2 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
		repo = parts[1]
	}
	return repo + ":" + tag
}

func (m Container) K8sEnv() []corev1.EnvVar {
	if len(m.Env) == 0 {
		return []corev1.EnvVar{}
	}

	results := make([]corev1.EnvVar, len(m.Env))
	for i, e := range m.Env {
		results[i] = corev1.EnvVar(e)
	}
	return results
}

//	type envJson struct {
//		Name  string `json:"name"`
//		Value string `json:"value"`
//	}
type Env corev1.EnvVar

type BasicParams struct {
	Namespace string `form:"namespace" json:"namespace"`
	// 服务名称
	Name string `uri:"name" binding:"required" json:"name"`

	// 修订描述
	Comment string `json:"comment,omitempty"`
}

type UpdateRequest struct {
	Ssid string `binding:"required" json:"-"`
	BasicParams
	// 资源类型
	ResourceType K8sResourceType `json:"resourceType" binding:"gt=0"`
	// 镜像
	Containers []Container `json:"containers" binding:"required"`
}
