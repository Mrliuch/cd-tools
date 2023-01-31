package apply

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gthub.com/Mrliuch/cd-tools/pkg/k8s"
)

func Apply(cmd *cobra.Command, workdir string) error {
	file, err := cmd.Flags().GetString("file")
	if err != nil || len(file) == 0 {
		logrus.Fatal("获取参数\"file\"失败,请提供yaml文件路径")
	}
	config, err := cmd.Flags().GetString("kube-config")
	if err != nil {
		logrus.Fatal("获取参数\"kube-config\"失败")
	}
	timeout, _ := cmd.Flags().GetInt("timeout")
	clusterName, _ := cmd.Flags().GetString("cluster-name")
	k8sController := k8s.KubernetesControllerImpl{
		Workdir:     workdir,
		ClusterName: clusterName,
		TimeOut:     timeout,
	}
	k8s.InitK8sClient(config, k8sController).CheckDeployByYaml(file).ApplyByYaml().Check().Rollback()

	return nil
}

// 检查kubeConfig文件以及yaml文件
func check() error {
	return nil
}
