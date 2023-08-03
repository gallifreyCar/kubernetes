package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"log"
	_ "net/http"
	"strings"
)

// 获取版本和依赖约束
func (reg *Registry) getVersionAndDependenceByUpdateRequest(req UpdateRequest) (string, map[string]string, error) {
	var version string
	deps := make(map[string]string)
	for _, c := range req.Containers {
		dependence, err := reg.GetImageDependenceRaw(c.GetImage())
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

// GetVersionAndDependenceByUpdateRequest 获取版本和依赖约束
func GetVersionAndDependenceByUpdateRequest(image string, reg *Registry) (string, map[string]string, error) {
	var version string
	deps := make(map[string]string)

	dependence, err := reg.GetImageDependenceRawV(image)
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
		i := strings.LastIndexByte(image, ':')
		if i == -1 {
			return "", nil, nil
		}
		v, err := semver.NewVersion(image[i+1:])
		if err != nil {
			return "", nil, nil
		}
		version = fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
	}
	return version, deps, nil
}

// 获取版本
// 从init容器和普通容器中依次遍历, 找到第一个符合语义化版本的镜像tag
func (reg *Registry) getVersionByPodTemplate(podSpec *corev1.PodTemplateSpec) string {
	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		i := strings.LastIndexByte(c.Image, ':')
		if i == -1 {
			continue
		}
		v, err := semver.NewVersion(c.Image[i+1:])
		if err != nil {
			continue
		}
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
	}

	return ""
}

// 获取版本
// 从init容器和普通容器中依次遍历, 找到第一个符合语义化版本的镜像tag
func (reg *Registry) getVersionByPodTemplateV(podSpec *corev1.PodTemplateSpec) string {
	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		i := strings.LastIndexByte(c.Image, ':')
		if i == -1 {
			continue
		}
		v, err := semver.NewVersion(c.Image[i+1:])
		if err != nil {
			continue
		}
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
	}

	return ""
}

// 获取版本
// 从init容器和普通容器中依次遍历, 找到第一个符合语义化版本的镜像tag
func (reg *Registry) getVersionByPodTemplateVV(podSpec *corev1.PodTemplateSpec) string {
	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		if c.Image == "" {
			continue
		}
		i := strings.Index(c.Image, "/")
		return c.Image[i+1:]
	}

	return ""
}

// 隐藏镜像仓库地址
// image: harbor:5000/wecloud/wmc:1.5.1
// return: wecloud/wmc:1.5.1
func hideImageRegistry(image string) string {
	i := strings.IndexByte(image, '/')
	if i == -1 {
		return image
	}
	return image[i+1:]
}

// 获取依赖约束
// 从init容器和普通容器中依次遍历, 获取每个镜像的依赖约束
func (reg *Registry) getDependenceByPodTemplate(podSpec *corev1.PodTemplateSpec) (map[string]string, error) {
	deps := make(map[string]string)

	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		dependence, err := reg.GetImageDependenceRaw(hideImageRegistry(c.Image))
		if err != nil {
			return nil, err
		}
		for k, v := range dependence {
			if got, ok := deps[k]; ok {
				deps[k] = got + "," + v
			} else {
				deps[k] = v
			}
		}
	}

	return deps, nil
}
func (reg *Registry) GetDependenceByPodTemplateV(podSpec *corev1.PodTemplateSpec) (map[string]string, error) {
	deps := make(map[string]string)

	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		dependence, err := reg.GetImageDependenceRawV(hideImageRegistry(c.Image))
		if err != nil {
			return nil, err
		}
		for k, v := range dependence {
			if got, ok := deps[k]; ok {
				deps[k] = got + "," + v
			} else {
				deps[k] = v
			}
		}
	}

	return deps, nil
}

// 获取版本
// Deployment, StatefulSet, DaemonSet资源从spec.template中获取版本和依赖 getVersionByPodTemplate
func (reg *Registry) getVersion(ctx context.Context, krt K8sResourceType, obj *unstructured.Unstructured) (string, error) {
	version := obj.GetLabels()[K8sLabelVersion]
	if version != "" {
		return version, nil
	}

	switch krt {
	case KRTDeployment, KRTStatefulSet, KRTDaemonSet:
		podSpec, err := getObjPodTemplate(obj)
		if err != nil {
			return "", err
		}
		return reg.getVersionByPodTemplate(podSpec), nil
	default:
		return "", errors.New("不支持的资源类型")
	}
}

// GetVersionV GetVersion 获取版本
// Deployment, StatefulSet, DaemonSet资源从spec.template中获取版本和依赖 getVersionByPodTemplate
func (reg *Registry) GetVersionV(krt K8sResourceType, obj *unstructured.Unstructured) (string, error) {
	version := obj.GetLabels()[K8sLabelVersion]
	if version != "" {
		return version, nil
	}

	switch krt {
	case KRTDeployment, KRTStatefulSet, KRTDaemonSet:
		podSpec, err := getObjPodTemplate(obj)
		if err != nil {
			return "", err
		}
		return reg.getVersionByPodTemplateV(podSpec), nil
	default:
		return "", errors.New("不支持的资源类型")
	}
}

// GetImageV  GetVersion 获取版本
// Deployment, StatefulSet, DaemonSet资源从spec.template中获取版本和依赖 getVersionByPodTemplate
func (reg *Registry) GetImageV(krt K8sResourceType, obj *unstructured.Unstructured) (string, error) {
	version := obj.GetLabels()[K8sLabelVersion]
	if version != "" {
		return version, nil
	}

	switch krt {
	case KRTDeployment, KRTStatefulSet, KRTDaemonSet:
		podSpec, err := getObjPodTemplate(obj)
		if err != nil {
			return "", err
		}
		return reg.getVersionByPodTemplateVV(podSpec), nil
	default:
		return "", errors.New("不支持的资源类型")
	}
}

// GetVersionAndDependence 获取版本和依赖约束
func (reg *Registry) GetVersionAndDependence(krt K8sResourceType, obj *unstructured.Unstructured) (string, map[string]string, error) {
	switch krt {
	case KRTDeployment, KRTStatefulSet, KRTDaemonSet:
		podSpec, err := getObjPodTemplate(obj)
		if err != nil {
			return "", nil, err
		}
		version := reg.getVersionByPodTemplate(podSpec)
		deps, err := reg.getDependenceByPodTemplate(podSpec)
		return version, deps, err
	default:
		return "", nil, errors.New("不支持的资源类型")
	}
}

// 正向依赖检查
// dep: 被检查的服务的版本依赖约束
// 被依赖服务不存在时, 跳过检查, 而不是返回检查不通过
//
// 特殊情况1: 两个服务互相依赖时, 并且双方都不兼容对方老版本时, 会无法更新
func (reg *Registry) checkForwardDependence(ctx context.Context, objs map[string]*unstructured.Unstructured, deps map[string]string) error {
	log.Printf("正向依赖检查: %v", deps)
	for svc, constraint := range deps {
		c, err := semver.NewConstraint(constraint)
		if err != nil {
			return err
		}

		obj := objs[svc]
		if obj == nil {
			log.Printf("被依赖的服务不存在: %s", svc)
			continue
		}

		version, err := reg.getVersion(ctx, ParseResourceTypeFromObject(obj.Object), obj)
		if err != nil {
			return err
		}
		if version == "" {
			log.Printf("被依赖的服务版本为空: %s", svc)
			continue
		}

		v, err := semver.NewVersion(version)
		if err != nil {
			return err
		}
		if !c.Check(v) {
			return err
		}
	}
	return nil
}

func (reg *Registry) CheckForwardDependenceV(objs map[string]*unstructured.Unstructured, deps map[string]string) error {
	log.Printf("正向依赖检查: %v", deps)
	for svc, constraint := range deps {
		c, err := semver.NewConstraint(constraint)
		if err != nil {
			return err
		}

		obj := objs[svc]
		if obj == nil {
			log.Printf("被依赖的服务不存在: %s", svc)
			continue
		}

		version, err := reg.GetVersionV(ParseResourceTypeFromObject(obj.Object), obj)
		if err != nil {
			return err
		}
		if version == "" {
			log.Printf("被依赖的服务版本为空: %s", svc)
			continue
		}

		v, err := semver.NewVersion(version)
		if err != nil {
			return err
		}
		if !c.Check(v) {
			return errors.New("版本不符合约束")
		}
	}
	return nil
}

// 反向依赖检查
// version: 被检查的服务待更新的版本号
// svc: 待更新的服务版本名称
//
// 特殊情况1: 某个服务升级时, 依赖项从~1.1.0变到~1.2.0, 即不向下兼容老版本
//
//	此时正向依赖检查会失败, 需要升级依赖的服务至兼容的版本
//	此时被依赖的服务进行反向依赖检查会失败, 因为新版本(1.2.x)不符合服务当前版本约束(~1.1.0)
func (reg *Registry) checkReverseDependence(ctx context.Context, objs map[string]*unstructured.Unstructured, svc string, version string) error {
	log.Printf("反向依赖检查: %s %s", svc, version)
	if version == "" {
		return nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	key := svc + K8sAnnotationDependence
	for _, obj := range objs {
		depRaw := obj.GetAnnotations()[key]
		if depRaw == "" {
			continue
		}
		deps := strings.Split(depRaw, ",")
		for _, dep := range deps {
			c, err := semver.NewConstraint(dep)
			if err != nil {
				return err
			}
			if !c.Check(v) {
				return errors.New("反向依赖检查失败")
			}
		}
	}
	return nil
}

func (reg *Registry) CheckReverseDependenceV(objs map[string]*unstructured.Unstructured, svc string, version string) error {
	log.Printf("反向依赖检查: %s %s", svc, version)
	if version == "" {
		return nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	key := svc + K8sAnnotationDependence
	for _, obj := range objs {
		depRaw := obj.GetAnnotations()[key]
		if depRaw == "" {
			continue
		}
		deps := strings.Split(depRaw, ",")
		for _, dep := range deps {
			c, err := semver.NewConstraint(dep)
			if err != nil {
				return err
			}
			if !c.Check(v) {
				return errors.New("反向依赖检查失败")
			}
		}
	}
	return nil
}

//func (s *Service) queryPodOwnerObjects(ctx context.Context, namespace string) (map[string]*unstructured.Unstructured, error) {
//	pods, err := s.pod.Pods(namespace).List(labels.Everything())
//	if err != nil {
//		return nil, ErrQueryPods.Set(err)
//	}
//
//	objs := make(map[string]*unstructured.Unstructured)
//	for _, p := range pods {
//		obj, err := s.getPodOwner(ctx, p)
//		if err != nil {
//			s.logger.Error("无法获取PodOwner的meta", zap.String("pod", p.Name), zap.Error(err), log.Context(ctx))
//			continue
//		}
//		objs[obj.GetName()] = obj
//	}
//	return objs, nil
//}

//// 版本依赖检查
//func (s *Service) checkDependence(ctx context.Context, namespace, name, version string, deps map[string]string) error {
//	if s.cfg.DependencePolicy != dependencePolicyComplete {
//		return nil
//	}
//
//	objs, err := s.queryPodOwnerObjects(ctx, namespace)
//	if err != nil {
//		return err
//	}
//	if err := s.checkForwardDependence(ctx, objs, deps); err != nil {
//		return err
//	}
//	if err := s.checkReverseDependence(ctx, objs, name, version); err != nil {
//		return err
//	}
//	return nil
//}

//func setObjVersion(obj *unstructured.Unstructured, version string, deps map[string]string) {
//	Labels := obj.GetLabels()
//	if Labels == nil {
//		Labels = map[string]string{}
//	}
//	Labels[pkg.K8sLabelVersion] = version
//	obj.SetLabels(Labels)
//
//	annotations := obj.GetAnnotations()
//	if annotations == nil {
//		annotations = map[string]string{}
//	}
//	for k, v := range deps {
//		annotations[k+pkg.K8sAnnotationDependence] = v
//	}
//	obj.SetAnnotations(annotations)
//}

func getObjPodTemplate(obj *unstructured.Unstructured) (*corev1.PodTemplateSpec, error) {
	specRaw, ok, err := unstructured.NestedMap(obj.Object, "spec", "template")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("未查询到spec.template")
	}
	var podSpec corev1.PodTemplateSpec
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(specRaw, &podSpec); err != nil {
		return nil, err
	}
	return &podSpec, nil
}

const (
	KRTUnknown             K8sResourceType = iota // unknown
	KRTDeployment                                 // Deployment
	KRTStatefulSet                                // StatefulSet
	KRTDaemonSet                                  // DaemonSet
	KRTMonitorCrdKafka                            // MonitorKafka
	KRTMonitorCrdMysql                            // MonitorMysql
	KRTMonitorCrdRedis                            // MonitorRedis
	KRTMonitorCrdZookeeper                        // MonitorZookeeper
	KRTWellcloudCms                               // WellcloudCms
	KRTClusterRole                                // ClusterRole
	KRTClusterRoleBinding                         // ClusterRoleBinding
	KRTServiceAccount                             // ServiceAccount
	KRTService                                    // Service
	KRTConfigMap                                  // ConfigMap
	KRTPod                                        // Pod
	KrtReplicaSet                                 // ReplicaSet
	KRTMonitorCrdRabbitMQ                         // RabbitMQ
	KRTJob                                        // Job
	KRTCronJob                                    // CronJob
)

//go:generate stringer -type=K8sResourceType -linecomment
type K8sResourceType int

func ParseResourceType(kind string) K8sResourceType {
	switch kind {
	case "Deployment":
		return KRTDeployment
	case "StatefulSet":
		return KRTStatefulSet
	case "DaemonSet":
		return KRTDaemonSet
	case "Kafka":
		return KRTMonitorCrdKafka
	case "Mysql":
		return KRTMonitorCrdMysql
	case "Redis":
		return KRTMonitorCrdRedis
	case "Zookeeper":
		return KRTMonitorCrdZookeeper
	case "Cms":
		return KRTWellcloudCms
	case "ClusterRole":
		return KRTClusterRole
	case "ClusterRoleBinding":
		return KRTClusterRoleBinding
	case "ServiceAccount":
		return KRTServiceAccount
	case "Service":
		return KRTService
	case "ConfigMap":
		return KRTConfigMap
	case "Pod":
		return KRTPod
	case "ReplicaSet":
		return KrtReplicaSet
	case "RabbitMQ":
		return KRTMonitorCrdRabbitMQ
	case "Job":
		return KRTJob
	case "CronJob":
		return KRTCronJob
	default:
		return KRTUnknown
	}
}

func ParseResourceTypeFromObject(obj map[string]interface{}) K8sResourceType {
	gotKind := obj["kind"]
	kind, ok := gotKind.(string)
	if !ok {
		return KRTUnknown
	}
	return ParseResourceType(kind)
}

func (k K8sResourceType) IsCrd() bool {
	return (k >= KRTMonitorCrdKafka && k <= KRTWellcloudCms) || k == KRTMonitorCrdRabbitMQ
}

func (k K8sResourceType) ShouldCheckVersion() bool {
	return k == KRTDeployment || k == KRTDaemonSet || k == KRTStatefulSet
}

var gvrMap = map[K8sResourceType]schema.GroupVersionResource{
	KRTDeployment:          {Group: "apps", Version: "v1", Resource: "deployments"},
	KRTDaemonSet:           {Group: "apps", Version: "v1", Resource: "daemonsets"},
	KRTStatefulSet:         {Group: "apps", Version: "v1", Resource: "statefulsets"},
	KRTMonitorCrdKafka:     {Group: "monitor.welljoint.com", Version: "v1alpha1", Resource: "kafkas"},
	KRTMonitorCrdMysql:     {Group: "monitor.welljoint.com", Version: "v1alpha1", Resource: "mysqls"},
	KRTMonitorCrdRedis:     {Group: "monitor.welljoint.com", Version: "v1alpha1", Resource: "redis"},
	KRTMonitorCrdZookeeper: {Group: "monitor.welljoint.com", Version: "v1alpha1", Resource: "zookeepers"},
	KRTWellcloudCms:        {Group: "wellcloud.welljoint.com", Version: "v1alpha1", Resource: "cms"},
	KRTClusterRole:         {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	KRTClusterRoleBinding:  {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	KRTServiceAccount:      {Group: "", Version: "v1", Resource: "serviceaccounts"},
	KRTService:             {Group: "", Version: "v1", Resource: "services"},
	KRTConfigMap:           {Group: "", Version: "v1", Resource: "configmaps"},
	KRTPod:                 {Group: "", Version: "v1", Resource: "pods"},
	KrtReplicaSet:          {Group: "apps", Version: "v1", Resource: "replicasets"},
	KRTMonitorCrdRabbitMQ:  {Group: "monitor.welljoint.com", Version: "v1alpha1", Resource: "rabbitmqs"},
	KRTJob:                 {Group: "batch", Version: "v1", Resource: "jobs"},
	// TODO: http://172.16.200.215:8080/browse/WEL2X-2558
	KRTCronJob: {Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
}

func (k K8sResourceType) GVR() schema.GroupVersionResource { return gvrMap[k] }

//var schemaMap = map[K8sResourceType]interface{}{
//	KRTDeployment:          appsv1.Deployment{},
//	KRTDaemonSet:           appsv1.DaemonSet{},
//	KRTStatefulSet:         appsv1.StatefulSet{},
//	KRTMonitorCrdKafka:     monitorv1alpha1.Kafka{},
//	KRTMonitorCrdMysql:     monitorv1alpha1.Mysql{},
//	KRTMonitorCrdRedis:     monitorv1alpha1.Redis{},
//	KRTMonitorCrdZookeeper: monitorv1alpha1.Zookeeper{},
//	KRTWellcloudCms:        wellcloudv1alpha1.Cms{},
//	KRTClusterRole:         rbacv1.ClusterRole{},
//	KRTClusterRoleBinding:  rbacv1.ClusterRoleBinding{},
//	KRTServiceAccount:      corev1.ServiceAccount{},
//	KRTService:             corev1.Service{},
//	KRTConfigMap:           corev1.ConfigMap{},
//	KRTPod:                 corev1.Pod{},
//	KrtReplicaSet:          appsv1.ReplicaSet{},
//	KRTMonitorCrdRabbitMQ:  monitorv1alpha1.RabbitMQ{},
//	KRTJob:                 batchv1.Job{},
//	KRTCronJob:             batchv1beta1.CronJob{},
//}

//func (k K8sResourceType) SchemaStruct() interface{} { return schemaMap[k] }

const (
	K8sLabelName            = "wkm.welljoint.com/name"        // 服务名称
	K8sLabelVersion         = "wkm.welljoint.com/version"     // 服务版本
	K8sAnnotationDependence = ".wkm.welljoint.com/dependence" // 依赖约束
)

// SetObjVersion 设置对象的版本号
func SetObjVersion(obj *unstructured.Unstructured, version string, deps map[string]string) {
	Labels := obj.GetLabels()
	if Labels == nil {
		Labels = map[string]string{}
	}
	Labels[K8sLabelVersion] = version
	obj.SetLabels(Labels)

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	for k, v := range deps {
		annotations[k+K8sAnnotationDependence] = v
	}
	obj.SetAnnotations(annotations)
}

var count = 0

// GetResourceOwner 获取资源的owner
func (reg *Registry) GetResourceOwner(obj *unstructured.Unstructured, krt K8sResourceType, ff cmdutil.Factory) (*unstructured.Unstructured, K8sResourceType, error) {
	refs := obj.GetOwnerReferences()
	if len(refs) == 0 {
		return obj, krt, nil
	}

	refKrt := ParseResourceType(refs[0].Kind)
	if refKrt == KRTUnknown {
		return nil, KRTUnknown, errors.New("unknown owner kind: " + refs[0].Kind)
	}
	count++
	println("count:", count)
	s := ff.NewBuilder().Unstructured().
		NamespaceParam(obj.GetNamespace()).
		ContinueOnError().
		Latest().
		Flatten().
		ResourceTypeOrNameArgs(true, refs[0].Kind, refs[0].Name).
		Do()

	sInfos, err := s.Infos()
	if err != nil {
		return nil, KRTUnknown, err
	}

	return reg.GetResourceOwner(sInfos[0].Object.(*unstructured.Unstructured), refKrt, ff)
}
