package k8s

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KubernetesControllerImpl) GetNode() {
	nodeList, err := k.ClientSet.CoreV1().Nodes().List(context.Background(), meta_v1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logrus.Fatalf("获取node节点失败：%s", err.Error())
	}
	for _, node := range nodeList.Items {
		fmt.Println(node.Name)
	}
}

func (k *KubernetesControllerImpl) GetPodWithLabel(labels map[string]string, namespace string) ([]corev1.Pod, error) {
	podList := []corev1.Pod{}
	pods, err := k.ClientSet.CoreV1().Pods(namespace).List(context.Background(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		flag := make(map[string]bool)
		for kk, v := range labels {
			if _, ok := pod.Labels[kk]; ok {
				if ok {
					if pod.Labels[kk] == v {
						flag[kk] = true
					} else {
						flag[kk] = false
					}
				}
			} else {
				flag[kk] = false
			}
		}
		f := true
		for _, st := range flag {
			if !st {
				f = false
			}
		}
		if f {
			podList = append(podList, pod)
		}
	}
	return podList, nil
}
