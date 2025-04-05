package utils

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"os"
	"sigs.k8s.io/yaml/goyaml.v3"
)

func LoadServiceEntry(filename string) (*v1alpha3.ServiceEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return nil, err
	}
	var jsonData interface{}
	if err := yaml.Unmarshal(data, &jsonData); err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
		return nil, err
	}
	result, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, err
	}

	seJson := v1alpha3.ServiceEntry{}
	err = json.Unmarshal(result, &seJson)

	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}

	return &seJson, err
}

func LoadServiceMonitor(filename string) (*v1.ServiceMonitor, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return nil, err
	}
	var jsonData interface{}
	if err := yaml.Unmarshal(data, &jsonData); err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
		return nil, err
	}
	result, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return nil, err
	}

	smJson := v1.ServiceMonitor{}
	err = json.Unmarshal(result, &smJson)

	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	return &smJson, err

}
