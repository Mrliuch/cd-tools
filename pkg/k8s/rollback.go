package k8s

import (
	"github.com/sirupsen/logrus"
	"gthub.com/Mrliuch/cd-tools/pkg/utils"
)

func (k *KubernetesControllerImpl) Rollback() *KubernetesControllerImpl {
	if k.CheckPass {
		logrus.Infof("恭喜完成了：%s的升级部署操作，Pod已全部Running，本次操作已结束", k.Name)
		return k
	}
	logrus.Infof("正在进行回滚操作")
	oldYaml, err := utils.ReadYaml(k.BackupPath)
	if err != nil {
		logrus.Fatalf("读取备份文件失败：%s", err.Error())
	}
	ey := NewYaml(k, string(oldYaml))
	err = ey.UpdateFromYaml()
	if err != nil {
		logrus.Fatalf("回滚Deployment失败：%s", err.Error())
	}
	logrus.Infof("回滚完成：%s，请检查是否回滚成功，本次操作已结束", k.Name)
	return k
}
