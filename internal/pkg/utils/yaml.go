package utils

import "sigs.k8s.io/yaml"

func MustYamlString(value interface{}) string {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
