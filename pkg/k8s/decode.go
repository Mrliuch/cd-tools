package k8s

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"gthub.com/Mrliuch/cd-tools/pkg/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func parsePostedData(data []byte) [][]byte {
	var finalData [][]byte
	reader := bytes.NewReader(data)
	dec := yaml.NewDecoder(reader)
	var strParts interface{}
	for dec.Decode(&strParts) == nil {
		out, _ := yaml.Marshal(strParts)
		finalData = append(finalData, out)
	}
	return finalData
}

func ReadYamlByPath(p string) (object runtime.Object, meta *v1.ObjectMeta, kind, namespace, name, k8sYaml string) {
	content, err := utils.ReadYaml(p)
	if err != nil {
		logrus.Fatalf("读取yaml文件失败：%s", err.Error())
	}
	list := parsePostedData(content)
	if len(list) == 0 || len(list) > 1 {
		logrus.Fatalf("读取yaml文件失败：未获取到文件内容或文件内含多yaml")
	}
	object, meta, err = DecodeYamlOrJson(string(list[0]))
	if err != nil {
		logrus.Fatalf("解析yaml文件失败：%s", err.Error())
	}
	return object, meta, object.GetObjectKind().GroupVersionKind().Kind, meta.GetNamespace(), meta.GetName(), string(list[0])
}
func objectMetaFor(obj interface{}) (*v1.ObjectMeta, error) {
	if newObj, isOK := obj.(runtime.Object); isOK {
		v, err := conversion.EnforcePtr(newObj)
		if err != nil {
			return nil, err
		}
		var meta *v1.ObjectMeta
		err = runtime.FieldPtr(v, "ObjectMeta", &meta)
		return meta, err
	}
	return nil, fmt.Errorf("Unsupported object: %#v", obj)
}
func DecodeYamlOrJson(content string) (runtime.Object, *v1.ObjectMeta, error) {
	var decoder func(data []byte, defaults *schema.GroupVersionKind, into runtime.Object) (runtime.Object, *schema.GroupVersionKind, error)
	decoder = scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decoder([]byte(content), nil, nil)
	if err != nil {
		return nil, nil, err
	}
	meta, err := objectMetaFor(obj)
	if err != nil {
		return nil, nil, err
	}
	o := meta.GetObjectMeta()
	if o.GetNamespace() == "" {
		o.SetNamespace("default")
	}
	return obj, meta, err
}
