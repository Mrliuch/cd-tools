package k8s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gthub.com/Mrliuch/cd-tools/pkg/utils"
	"io"
	ko_v1 "k8s.io/api/apps/v1"
	ko_v1beta "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/restmapper"
	"path"
	k8s_yaml "sigs.k8s.io/yaml"
	"time"
)

func (k *KubernetesControllerImpl) CheckDeployByYaml(p string) *KubernetesControllerImpl {
	object, meta, kind, namespace, name, k8sYaml := ReadYamlByPath(p)
	k.Name = name
	k.Namespace = namespace
	k.K8SYaml = k8sYaml
	k.Object = object
	k.Meta = meta
	switch kind {
	case "Deployment":
		k.BackupYamlWithDeployment(object)
	default:
		logrus.Fatalf("暂时进支持Deployment类型，不支持：%s", kind)
	}
	return k
}

func (k *KubernetesControllerImpl) BackupYamlWithDeployment(object runtime.Object) *KubernetesControllerImpl {
	name := k.Name
	namespace := k.Namespace
	var deployment_yaml []byte
	switch object.(type) {
	case *ko_v1beta.Deployment:
		deployment, err := k.ClientSet.AppsV1beta1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			deployment.Kind = object.GetObjectKind().GroupVersionKind().Kind
			deployment.APIVersion = "apps/" + object.GetObjectKind().GroupVersionKind().Version
			deployment.ResourceVersion = ""
			deployment.UID = ""
			deployment.SelfLink = ""
			deployment.ManagedFields = nil
			deployment_yaml, err = k8s_yaml.Marshal(deployment)
			if err != nil {
				logrus.Fatalf("转换备份文件失败：%s", err.Error())
			}
		} else {
			logrus.Warningf("在集群里为找到%s的Deployment定义", err.Error())
			return k
		}

	case *ko_v1.Deployment:
		deployment, err := k.ClientSet.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err == nil {
			deployment.Kind = object.GetObjectKind().GroupVersionKind().Kind
			deployment.APIVersion = "apps/" + object.GetObjectKind().GroupVersionKind().Version
			deployment.ManagedFields = nil
			deployment.ResourceVersion = ""
			deployment.UID = ""
			deployment.SelfLink = ""
			deployment_yaml, err = k8s_yaml.Marshal(deployment)
			if err != nil {
				logrus.Fatalf("转换备份文件失败：%s", err.Error())
			}
		} else {
			logrus.Warningf("在集群里未找到：%s的Deployment定义,准备创建", name)
			return k
		}
	}
	p := path.Join(k.Workdir, name+"_"+namespace)
	_ = utils.PathExistsOrCreate(p)
	filename := path.Join(p, time.Now().Format("20060102_15_04_05")+".yml")
	logrus.Infof("正在备份集群内已部署Deployment文件，备份文件路径为:%s", filename)
	k.BackupPath = filename
	err := utils.PathExistsOrCreateFile(p)
	if err != nil {
		logrus.Fatalf("创建备份文件失败：%s", err.Error())
	}
	err = utils.WriteFile(filename, string(deployment_yaml))
	if err != nil {
		logrus.Fatalf("写入备份文件失败：%s", err.Error())
	}
	return k
}

func (k *KubernetesControllerImpl) ApplyByYaml() *KubernetesControllerImpl {
	logrus.Infof("正在执行更新创建Deployment")
	ey := NewYaml(k, k.K8SYaml)
	err := ey.UpdateFromYaml()
	if err != nil {
		logrus.Fatalf("更新创建Deployment失败：%s", err.Error())
	}
	logrus.Infof("更新集群内名称为：%s，命名空间：%s的Deployment成功", k.Name, k.Namespace)
	return k
}

type ExecuteYaml struct {
	applyYaml string
	namespace string
	k         *KubernetesControllerImpl
}

func NewYaml(k *KubernetesControllerImpl, y string) *ExecuteYaml {
	return &ExecuteYaml{
		applyYaml: y,
		namespace: k.Namespace,
		k:         k,
	}
}

func (y *ExecuteYaml) GtGVR(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {

	clientset := y.k.ClientSet
	gr, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return mapping.Resource, nil
}

func (y *ExecuteYaml) UpdateFromYaml() error {
	dynameicclient := y.k.DynamicClient
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewBufferString(y.applyYaml), 4096)

	for {
		var rawObj runtime.RawExtension
		err := d.Decode(&rawObj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decode is err %v", err)
		}

		obj, _, err := syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return fmt.Errorf("rawobj is err%v", err)
		}

		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return fmt.Errorf("tounstructured is err %v", err)
		}

		unstructureObj := &unstructured.Unstructured{Object: unstructuredMap}
		gvr, err := y.GtGVR(unstructureObj.GroupVersionKind())
		if err != nil {
			return err
		}
		unstructuredYaml, err := k8s_yaml.Marshal(unstructureObj)
		if err != nil {
			return fmt.Errorf("unable to marshal resource as yaml: %w", err)
		}
		_, getErr := dynameicclient.Resource(gvr).Namespace(y.namespace).Get(context.Background(), unstructureObj.GetName(), metav1.GetOptions{})
		if getErr != nil {
			_, createErr := dynameicclient.Resource(gvr).Namespace(y.namespace).Create(context.Background(), unstructureObj, metav1.CreateOptions{})
			if createErr != nil {
				return createErr
			}
		}

		force := true
		if y.namespace == unstructureObj.GetNamespace() {

			_, err = dynameicclient.Resource(gvr).
				Namespace(y.namespace).
				Patch(context.Background(),
					unstructureObj.GetName(),
					types.ApplyPatchType,
					unstructuredYaml, metav1.PatchOptions{
						FieldManager: unstructureObj.GetName(),
						Force:        &force,
					})

			if err != nil {
				return fmt.Errorf("unable to patch resource: %w", err)
			}

		} else {

			_, err = dynameicclient.Resource(gvr).
				Patch(context.Background(),
					unstructureObj.GetName(),
					types.ApplyPatchType,
					unstructuredYaml, metav1.PatchOptions{
						Force:        &force,
						FieldManager: unstructureObj.GetName(),
					})
			if err != nil {
				return fmt.Errorf("ns is nil unable to patch resource: %w", err)
			}
		}
	}
	return nil
}
