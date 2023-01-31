package utils

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const FILEMODE os.FileMode = 0666

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func PathExistsOrCreate(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, FILEMODE)
		return err
	}
	return err
}
func PathExistsOrCreateFile(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		_, err = os.Create(path)
		return err
	}
	return err
}

func ReadYaml(path string) ([]byte, error) {
	var err error
	content := []byte{}
	content, err = ioutil.ReadFile(path)
	if err != nil {
		return content, err
	}
	return content, nil
}

func WriteYaml(path string, c interface{}) error {
	yamlContent, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, yamlContent, FILEMODE)
	return err
}

func WriteFile(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), FILEMODE)
}
