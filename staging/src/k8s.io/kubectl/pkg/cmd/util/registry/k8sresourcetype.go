package registry

import "k8s.io/apimachinery/pkg/runtime/schema"

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

const (
	K8sLabelName            = "wkm.welljoint.com/name"        // 服务名称
	K8sLabelVersion         = "wkm.welljoint.com/version"     // 服务版本
	K8sAnnotationDependence = ".wkm.welljoint.com/dependence" // 依赖约束
)
