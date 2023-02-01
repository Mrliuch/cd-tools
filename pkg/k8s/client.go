package k8s

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gthub.com/Mrliuch/cd-tools/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var schemeNew = runtime.NewScheme()

type KubernetesControllerImpl struct {
	Client        client.Client
	ClientSet     *k8s.Clientset
	DynamicClient *dynamic.DynamicClient
	Object        runtime.Object
	Meta          *v1.ObjectMeta
	Workdir       string
	Name          string
	Namespace     string
	K8SYaml       string
	TimeOut       time.Duration
	ClusterName   string
	CheckPass     bool
	BackupPath    string
}

func InitK8sClient(configPath string, k KubernetesControllerImpl) *KubernetesControllerImpl {
	cli, err := NewK8sClient(configPath, &k)
	if err != nil {
		logrus.Fatalf("初始化K8S集群失败: %s", err.Error())
	}
	if len(k.ClusterName) == 0 {
		r, _ := regexp.Compile("((25[0-5]|2[0-4]\\d|[01]?\\d\\d?)\\.){3}(25[0-5]|2[0-4]\\d|[01]?\\d\\d?)")
		config, err := utils.ReadYaml(configPath)
		if err != nil {
			logrus.Fatalf("读取kube-config文件失败：%s", err.Error())
		}
		ipByte := r.Find([]byte(config))
		if len(ipByte) == 0 {
			logrus.Fatalf("请设置集群名称")
		}
		k.ClusterName = string(ipByte)
	}
	logrus.Infof("初始化集群名称：%s", k.ClusterName)
	k.Client = cli
	k.Workdir = path.Join(k.Workdir, k.ClusterName)
	err = utils.PathExistsOrCreate(k.Workdir)
	if err != nil {
		logrus.Fatalf("设置工作目录错误：%s", err.Error())
	}

	return &k
}
func NewK8sClient(configPath string, k *KubernetesControllerImpl) (client.Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}
	clientSet, err := k8s.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("初始化K8S集群失败: %s", err.Error())
	}
	k.ClientSet = clientSet
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("初始化K8S集群失败: %s", err.Error())
	}
	k.DynamicClient = dynamicClient
	mgr, err := ctrl.NewManager(config, ctrl.Options{Scheme: schemeNew})
	if err != nil {
		logrus.Fatalf("初始化K8S集群失败: %s", err.Error())
		return nil, err
	}
	logrus.Info("初始化K8S集群中...")
	c := mgr.GetClient()
	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			logrus.Fatalf("初始化K8S集群失败:: %s", err.Error())
		}
	}()
	if !mgr.GetCache().WaitForCacheSync(context.TODO()) {
		return nil, fmt.Errorf("加载K8S集群失败")
	}
	return c, err
}
