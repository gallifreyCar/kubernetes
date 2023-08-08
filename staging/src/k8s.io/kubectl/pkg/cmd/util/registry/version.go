package registry

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	_ "net/http"
	"strings"
)

// 获取版本
// 从init容器和普通容器中依次遍历, 找到第一个符合语义化版本的镜像tag
func getVersionByPodTemplate(podSpec *corev1.PodTemplateSpec) string {
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

// GetImageByPodTemplate 获取版本
// 从init容器和普通容器中依次遍历, 找到第一个符合语义化版本的镜像
func GetImageByPodTemplate(podSpec *corev1.PodTemplateSpec) string {
	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		if c.Image == "" {
			continue
		}
		return c.Image
	}

	return ""
}

// 获取依赖约束
// 从init容器和普通容器中依次遍历, 获取每个镜像的依赖约束
func getDependenceByPodTemplate(podSpec *corev1.PodTemplateSpec) (map[string]string, error) {
	deps := make(map[string]string)

	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		i := strings.LastIndexByte(c.Image, ':')
		if i == -1 {
			continue
		}

		dependence, err := GetImageDependenceRaw(c.Image)
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

// GetVersion
// Deployment, StatefulSet, DaemonSet资源从spec.template中获取版本和依赖 getVersionByPodTemplate
func GetVersion(krt K8sResourceType, obj *unstructured.Unstructured) (string, error) {
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
		return getVersionByPodTemplate(podSpec), nil
	default:
		return "", errors.New("不支持的资源类型")
	}
}

// GetVersionAndDependence 获取版本和依赖约束
func GetVersionAndDependence(krt K8sResourceType, obj *unstructured.Unstructured) (string, map[string]string, error) {
	switch krt {
	case KRTDeployment, KRTStatefulSet, KRTDaemonSet:
		podSpec, err := getObjPodTemplate(obj)
		if err != nil {
			return "", nil, err
		}
		version := getVersionByPodTemplate(podSpec)
		deps, err := getDependenceByPodTemplate(podSpec)
		return version, deps, err
	default:
		return "", nil, errors.New("不支持的资源类型")
	}
}

func CheckForwardDependence(objs map[string]*unstructured.Unstructured, deps map[string]string) error {
	klog.V(4).Infof("正向依赖检查: %v\n", deps)
	for svc, constraint := range deps {
		c, err := semver.NewConstraint(constraint)
		if err != nil {
			return err
		}

		obj := objs[svc]
		if obj == nil {
			klog.V(4).Info("被依赖的服务不存在: %s\n", svc)
			continue
		}

		version, err := GetVersion(ParseResourceTypeFromObject(obj.Object), obj)
		if err != nil {
			return err
		}
		if version == "" {
			klog.V(4).Infof("被依赖的服务版本为空: %s\n", svc)
			continue
		}

		v, err := semver.NewVersion(version)
		if err != nil {
			return err
		}
		if !c.Check(v) {
			return errors.New(fmt.Sprintf("正向依赖检查失败，%s版本(%s)不符合依赖约束(%s)", svc, version, constraint))
		}
	}
	return nil
}

func CheckReverseDependence(objs map[string]*unstructured.Unstructured, svc string, version string) error {
	klog.V(4).Infof("反向依赖检查: %s %s\n", svc, version)
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
				return errors.New(fmt.Sprintf("反向依赖检查失败，%s版本(%s)不符合%s的依赖约束(%s)", svc, version, obj.GetName(), dep))
			}
		}
	}
	return nil
}

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

func CheckDep(info *resource.Info, ff cmdutil.Factory) error {
	krt := ParseResourceType(info.Object.GetObjectKind().GroupVersionKind().Kind)
	if krt < KRTDeployment || krt > KRTDaemonSet {
		return nil
	}

	//通过镜像获取版本和反向依赖
	gVersion, deps, err := GetVersionAndDependence(krt, info.Object.(*unstructured.Unstructured))
	if err != nil {
		return err
	}
	//设置反向依赖的annotation
	SetObjVersion(info.Object.(*unstructured.Unstructured), gVersion, deps)

	g := ff.NewBuilder().Unstructured().
		NamespaceParam(info.Namespace).
		ContinueOnError().
		Latest().
		Flatten().
		ResourceTypeOrNameArgs(true, "Pod").
		Do()
	gInfos, err := g.Infos()
	if err != nil {
		return err
	}
	var vInfos []*resource.Info
	//查4次，分别查deployment、statefulset、daemonset、ReplicaSet
	keys := []string{"Deployment", "StatefulSet", "ReplicaSet", "DaemonSet"}
	for _, key := range keys {
		g = ff.NewBuilder().Unstructured().
			NamespaceParam(info.Namespace).
			ContinueOnError().
			Latest().
			Flatten().
			ResourceTypeOrNameArgs(true, key).
			Do()
		tInfos, err := g.Infos()
		if err != nil {
			return err
		}
		vInfos = append(vInfos, tInfos...)
	}

	g = ff.NewBuilder().Unstructured().
		NamespaceParam(info.Namespace).
		ContinueOnError().
		Latest().
		Flatten().
		ResourceTypeOrNameArgs(true, "ReplicaSet").
		Do()
	tInfos, err := g.Infos()
	if err != nil {
		return err
	}
	vInfos = append(vInfos, tInfos...)

	g = ff.NewBuilder().Unstructured().
		NamespaceParam(info.Namespace).
		ContinueOnError().
		Latest().
		Flatten().
		ResourceTypeOrNameArgs(true, "DaemonSet").
		Do()

	tInfos, err = g.Infos()
	if err != nil {
		return err
	}
	vInfos = append(vInfos, tInfos...)

	var vMap = make(map[string]*unstructured.Unstructured)
	for _, v := range vInfos {
		vMap[v.Name+v.Object.GetObjectKind().GroupVersionKind().Kind] = v.Object.(*unstructured.Unstructured)
	}

	objs := map[string]*unstructured.Unstructured{}
	for _, i := range gInfos {
		got := i.Object.(*unstructured.Unstructured)
		owner, _, err := GetResourceOwner(got, krt, vMap)
		if err != nil {
			continue
		}

		ownerObj := owner
		objs[owner.GetName()] = ownerObj

	}

	//检测依赖
	if err = CheckForwardDependence(objs, deps); err != nil {
		return err
	}
	if err = CheckReverseDependence(objs, info.Name, gVersion); err != nil {
		return err
	}

	return nil
}

// GetResourceOwner 获取资源的owner
func GetResourceOwner(obj *unstructured.Unstructured, krt K8sResourceType, vMap map[string]*unstructured.Unstructured) (*unstructured.Unstructured, K8sResourceType, error) {
	refs := obj.GetOwnerReferences()
	if len(refs) == 0 {
		return obj, krt, nil
	}

	refKrt := ParseResourceType(refs[0].Kind)
	if refKrt == KRTUnknown {
		return nil, KRTUnknown, errors.New("unknown owner kind: " + refs[0].Kind)
	}

	next, ok := vMap[refs[0].Name+refs[0].Kind]
	if !ok {
		return nil, KRTUnknown, errors.New("unknown owner kind: " + refs[0].Kind)
	}

	return GetResourceOwner(next, refKrt, vMap)
}
