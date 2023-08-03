package registry

import (
	corev1 "k8s.io/api/core/v1"
	"strings"
)

// import (
//
//	"encoding/json"
//	"go.uber.org/zap/zapcore"
//	"io"
//	corev1 "k8s.io/api/core/v1"
//	"strings"
//
//	"wkm/api/types/options"
//	"wkm/storage/types"
//
// )
//
//	type ListRequest struct {
//		Namespace string `form:"namespace"`
//		// pod名称(模糊匹配)
//		Pod string `form:"pod"`
//		options.PageOption
//	}
//
//	func (l ListRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//		enc.AddString("namespace", l.Namespace)
//		enc.AddInt("page", l.Page)
//		enc.AddInt("limit", l.Limit)
//		return nil
//	}
//
//	type Pod struct {
//		Name     string `json:"name"`
//		Node     string `json:"node"`
//		State    string `json:"state"`
//		Message  string `json:"message"`
//		Uptime   int    `json:"uptime"`
//		Restarts string `json:"restarts"`
//	}
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

//	func (e Env) MarshalJSON() ([]byte, error) {
//		ej := envJson{Name: e.Name, Value: e.Value}
//		if e.ValueFrom != nil {
//			data, err := json.Marshal(e.ValueFrom)
//			if err != nil {
//				return nil, err
//			}
//			ej.Value = string(data)
//		}
//		return json.Marshal(ej)
//	}
//
//	func (e *Env) UnmarshalJSON(data []byte) error {
//		var ej envJson
//		if err := json.Unmarshal(data, &ej); err != nil {
//			return err
//		}
//		e.Name = ej.Name
//		if strings.HasPrefix(ej.Value, "{") && strings.HasSuffix(ej.Value, "}") {
//			var evs corev1.EnvVarSource
//			if err := json.Unmarshal([]byte(ej.Value), &evs); err != nil {
//				return err
//			}
//			e.ValueFrom = &evs
//		} else {
//			e.Value = ej.Value
//		}
//		return nil
//	}
//
//	func ParseEnvs(envs []corev1.EnvVar) []Env {
//		if len(envs) == 0 {
//			return []Env{}
//		}
//
//		results := make([]Env, len(envs))
//		for i, e := range envs {
//			results[i] = Env(e)
//		}
//		return results
//	}
//
// type ServiceInfos []ServiceInfo
//
//	func (infos ServiceInfos) Get(name string) *ServiceInfo {
//		for i, info := range infos {
//			if info.Name == name {
//				return &infos[i]
//			}
//		}
//		return nil
//	}
//
//	type ServiceInfo struct {
//		Name         string          `json:"name"`
//		ResourceType K8sResourceType `json:"resourceType"`
//		Containers   []Container     `json:"containers"`
//		Ready        int             `json:"ready"`
//		Total        int             `json:"total"`
//		Host         string          `json:"host"`
//		Ports        []string        `json:"ports"`
//		Pods         []Pod           `json:"pods"`
//		Link         string          `json:"link"`
//	}
//
//	type ListResponse struct {
//		Total    int          `json:"total"`
//		Services ServiceInfos `json:"services"`
//	}
//
//	type GetRequest struct {
//		Namespace string `form:"namespace" json:"namespace"`
//		// 服务名称
//		Name string `uri:"name" binding:"required" json:"name"`
//	}
//
//	func (r GetRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//		enc.AddString("namespace", r.Namespace)
//		enc.AddString("name", r.Name)
//		return nil
//	}
//
//	type GetResponse struct {
//		ServiceInfo
//	}
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

//
//func (u UpdateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", u.Namespace)
//	enc.AddString("name", u.Name)
//	enc.AddString("resourceType", u.ResourceType.String())
//	for _, i := range u.Containers {
//		enc.AddString("image."+i.Name, i.Image)
//	}
//	return nil
//}
//
//type ScaleRequest struct {
//	Ssid string `binding:"required" json:"-"`
//	BasicParams
//	// 资源类型
//	ResourceType K8sResourceType `json:"resourceType" binding:"gt=0"`
//	// 实例数
//	Replicas int `json:"replicas" binding:"gt=0"`
//}
//
//func (r ScaleRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("name", r.Name)
//	enc.AddInt("replicas", r.Replicas)
//	enc.AddString("resourceType", r.ResourceType.String())
//	return nil
//}
//
//type RevisionListRequest struct {
//	// 服务名称
//	Name string `uri:"name" binding:"required" json:"name"`
//	options.PageOption
//}
//
//func (r RevisionListRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("name", r.Name)
//	enc.AddInt("page", r.Page)
//	enc.AddInt("limit", r.Limit)
//	return nil
//}
//
//type Change struct {
//	Name string `json:"name"`
//	From string `json:"from"`
//	To   string `json:"to"`
//}
//
//type Revision struct {
//	// 修订版本号
//	Number int `json:"number"`
//	// 操作人
//	Operator string `json:"operator"`
//	// 创建时间
//	CreateAt types.DateTime `json:"createAt"`
//	// 变更内容
//	Changes []Change `json:"changes"`
//}
//
//type RevisionListResponse struct {
//	Total     int        `json:"total"`
//	Revisions []Revision `json:"revisions"`
//}
//
//type YamlGetRequest struct {
//	Namespace string `form:"namespace" json:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required" json:"name"`
//	// 修订版本号
//	// 如果未指定则查询当前的yaml
//	Revision int `form:"revision"`
//	// 资源类型
//	ResourceType K8sResourceType `form:"resourceType" binding:"gt=0"`
//}
//
//func (r YamlGetRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("name", r.Name)
//	enc.AddInt("revision", r.Revision)
//	enc.AddString("resourceType", r.ResourceType.String())
//	return nil
//}
//
//type YamlGetResponse struct {
//	Data []byte `json:"-"`
//}
//
//type RawImportRequest struct {
//	Ssid      string `binding:"required"`
//	Namespace string `form:"namespace"`
//	// 如果存在是否覆盖
//	Overwrite bool `form:"overwrite"`
//	// 导入的文件
//	Reader io.Reader
//}
//
//func (r RawImportRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	return nil
//}
//
//type YamlUpdateRequest struct {
//	Ssid string `binding:"required" json:"-"`
//	BasicParams
//
//	// raw json
//	Data map[string]interface{} `json:"data" binding:"required"`
//}
//
//func (r YamlUpdateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("name", r.Name)
//	return nil
//}
//
//type PodRestartRequest struct {
//	Ssid      string `binding:"required" json:"-"`
//	Namespace string `form:"namespace" json:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required" json:"name"`
//	// pod
//	Pod string `uri:"pod" binding:"required" json:"pod"`
//}
//
//func (r PodRestartRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("name", r.Name)
//	enc.AddString("pod", r.Pod)
//	return nil
//}
//
//type LogListRequest struct {
//	Namespace string `form:"namespace" json:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required" json:"name"`
//	// pod
//	Pod string `uri:"pod" binding:"required" json:"pod"`
//	options.PageOption
//}
//
//func (r LogListRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("name", r.Name)
//	enc.AddString("pod", r.Pod)
//	enc.AddInt("page", r.Page)
//	enc.AddInt("limit", r.Limit)
//	return nil
//}
//
//type LogListResponse struct {
//	logv1.ListResponse
//}
//
//type LogDownloadRequest struct {
//	Namespace string `form:"namespace" json:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required" json:"name"`
//	// pod
//	Pod string `uri:"pod" binding:"required" json:"pod"`
//
//	// 日志名和时间范围下载日志, 只能二选一
//	// 如果都指定日志名优先
//	// 下载的日志文件名
//	Logs []string `json:"logs"`
//	// 关键词
//	Keyword string `json:"keyword"`
//	// 开始和结束时间
//	TimeStart string `json:"timeStart"`
//	TimeEnd   string `json:"timeEnd"`
//
//	W io.Writer `json:"-"`
//}
//
//func (r LogDownloadRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("name", r.Name)
//	enc.AddString("pod", r.Pod)
//	enc.AddString("logs", strings.Join(r.Logs, ","))
//	return nil
//}
//
//type LogViewRequest struct {
//	Namespace string `form:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required"`
//	// pod
//	Pod string `uri:"pod" binding:"required"`
//	// 下载的日志文件名
//	Log string `form:"log" binding:"required"`
//}
//
//func (r LogViewRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("name", r.Name)
//	enc.AddString("pod", r.Pod)
//	enc.AddString("log", r.Log)
//	return nil
//}
//
//type LogViewResponse struct {
//	// 每行日志内容
//	Data []string `json:"data"`
//}
//
//type DeleteRequest struct {
//	Ssid      string `binding:"required" json:"-"`
//	Namespace string `form:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required"`
//	// 资源类型
//	ResourceType K8sResourceType `form:"resourceType" binding:"gt=0"`
//}
//
//func (d DeleteRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("name", d.Name)
//	enc.AddString("resourceType", d.ResourceType.String())
//	return nil
//}
//
//type GraphRequest struct {
//	Namespace string `form:"namespace"`
//	// 导出格式
//	Format string `form:"format"`
//}
//
//func (g GraphRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", g.Namespace)
//	enc.AddString("format", g.Format)
//	return nil
//}
//
//type DescribeRequest struct {
//	Namespace string `form:"namespace"`
//	// 资源类型
//	ResourceType K8sResourceType `form:"resourceType" binding:"gt=0"`
//	// 服务名称
//	Name string `uri:"name" binding:"required"`
//	// pod名称
//	Pod string `uri:"pod"`
//}
//
//func (r DescribeRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("krt", r.ResourceType.String())
//	enc.AddString("name", r.Name)
//	return nil
//}
//
//type DescribeResponse struct {
//	Output string
//}
//
//type RollbackRequest struct {
//	BasicParams
//	Ssid string `binding:"required" json:"-"`
//	// 资源类型
//	ResourceType K8sResourceType `form:"resourceType" binding:"gt=0"`
//	// 修订版本号
//	Revision int `form:"revision"`
//}
//
//func (r RollbackRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("krt", r.ResourceType.String())
//	enc.AddString("name", r.Name)
//	enc.AddInt("revision", r.Revision)
//	return nil
//}
//
//type BackupRequest struct {
//	Namespace string `form:"namespace"`
//}
//
//func (r BackupRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	return nil
//}
//
//type BackupResponse struct {
//	Success int `json:"success"`
//	Total   int `json:"total"`
//}
//
//func (r BackupResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddInt("success", r.Success)
//	enc.AddInt("total", r.Total)
//	return nil
//}
//
//type GetLogLevelRequest struct {
//	Namespace string `form:"namespace"`
//	// 服务名称
//	Name string `uri:"name" binding:"required"`
//}
//
//func (r GetLogLevelRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("name", r.Name)
//	return nil
//}
//
//type GetLogLevelResponse struct {
//	Main         map[string]string `json:"main"`
//	Dependencies map[string]string `json:"dependencies"`
//}
//
//func (r GetLogLevelResponse) GetLevel(name string) string {
//	if l, ok := r.Main[name]; ok {
//		return l
//	}
//	return r.Dependencies[name]
//}
//
//type SetLogLevelRequest struct {
//	Ssid string `binding:"required" json:"-"`
//	BasicParams
//
//	GetLogLevelResponse
//}
//
//func (r SetLogLevelRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("namespace", r.Namespace)
//	enc.AddString("name", r.Name)
//	for k, v := range r.Main {
//		enc.AddString("main."+k, v)
//	}
//	for k, v := range r.Dependencies {
//		enc.AddString("deps."+k, v)
//	}
//	return nil
//}
//
//type SetLogLevelResponse struct{}
