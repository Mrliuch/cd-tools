package k8s

import (
	"github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	"time"
)

func (k *KubernetesControllerImpl) Check() *KubernetesControllerImpl {
	logrus.Infof("正在进行Pod检测，时间：%ds，请等待", int(k.TimeOut.Seconds()))
	now := time.Now()
	time.Sleep(time.Second * 2) // 延迟2秒，以等待pod资源被k8s分发创建
	for {
		podStatus, err := k.check()
		if err != nil {
			logrus.Warning("Pod检测失败：%s", err.Error())
		}
		flagPhase := true
		for _, status := range podStatus {
			if status == 1 {
				flagPhase = false
				break
			}
		}
		if flagPhase {
			k.CheckPass = true
			break
		}
		time.Sleep(time.Second * 2)
		if time.Now().Sub(now) > k.TimeOut {
			logrus.Errorf("规定时间内Pod未成功启动")
			return k
		}
	}
	return k
}

func (k *KubernetesControllerImpl) check() (map[string]int, error) {
	label := k.Meta.Labels
	pods, err := k.GetPodWithLabel(label, k.Namespace)
	if err != nil {
		logrus.Errorf("未获取到Pod，等待下次获取")
	}
	podStatus := make(map[string]int)
	for _, pod := range pods {
		switch pod.Status.Phase {
		case coreV1.PodUnknown:
			podStatus[pod.Name] = 1
			logrus.Warnf("Pod名称：%s, 当前状态：%s", pod.Name, pod.Status.Phase)

		case coreV1.PodFailed:
			podStatus[pod.Name] = 1
			//logrus.Infof("Pod名称：%s, 当前状态：%s", pod.Name, pod.Status.Phase)
			for _, containerStatus := range pod.Status.ContainerStatuses { // 删除镜像拉取失败POD以重新拉取
				//i.AnsibleController.WriteMessage(fmt.Sprintf("Pod: %s, 当前Reason: %s", pod.Name, containerStatus.State.Waiting.Reason))
				if containerStatus.State.Waiting != nil {
					logrus.Warnf("Pod名称：%s, 当前状态：%s，原因:%s", pod.Name, pod.Status.Phase, containerStatus.State.Waiting.Reason)
				}
			}
		case coreV1.PodRunning:
			logrus.Infof("Pod：%s, 状态：%s", pod.Name, pod.Status.Phase)
			podStatus[pod.Name] = 0
		case coreV1.PodPending:
			for _, containerStatus := range pod.Status.ContainerStatuses { // 删除镜像拉取失败POD以重新拉取
				//i.AnsibleController.WriteMessage(fmt.Sprintf("Pod: %s, 当前Reason: %s", pod.Name, containerStatus.State.Waiting.Reason))
				if containerStatus.State.Waiting != nil {
					logrus.Warnf("Pod名称：%s, 当前状态：%s，原因:%s", pod.Name, pod.Status.Phase, containerStatus.State.Waiting.Reason)
				}
			}
			podStatus[pod.Name] = 1
		case coreV1.PodSucceeded:
			logrus.Infof("Pod：%s, 状态：%s", pod.Name, pod.Status.Phase)
			podStatus[pod.Name] = 0
		default:
			logrus.Warnf("Pod：%s, 状态：%s", pod.Name, pod.Status.Phase)
			podStatus[pod.Name] = 1
		}
	}
	return podStatus, nil
}
