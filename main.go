package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// External Secrets data structures
type SecretStoreRef struct {
	Name string
	Kind string
}

type Template struct {
	Data map[string]*string
}

type Target struct {
	Template Template
}

type RemoteRef struct {
	Key string
}

type SingleData struct {
	SecretKey string    `yaml:"secretKey"`
	RemoteRef RemoteRef `yaml:"remoteRef"`
}

type ExternalSecretSpec struct {
	RefreshInterval string         `yaml:"refreshInterval"`
	SecretStoreRef  SecretStoreRef `yaml:"secretStoreRef"`
	Target          Target
	Data            []SingleData
}

// Get environment variable or panic
func getEnvPanic(name string) string {
	value := os.Getenv(name)
	if len(value) == 0 {
		panic("ERROR: Environment variable '$" + name + "'is not set")
	}
	return value
}

// Get environment variable or use default value
func getEnvDefault(name, defaultValue string) string {
	value := os.Getenv(name)
	if len(value) > 0 {
		return value
	}
	return defaultValue
}

func printYaml(items interface{}) {
	yb, _ := yaml.Marshal(items)
	fmt.Println(string(yb))
}

// Create basic external secret spec
func createBasicExternalSecretSpec() ExternalSecretSpec {
	return ExternalSecretSpec{
		RefreshInterval: getEnvDefault("REFRESH_INTERVAL", "1h"),
		SecretStoreRef: SecretStoreRef{
			Name: getEnvPanic("STORE_NAME"),
			Kind: getEnvPanic("STORE_KIND"),
		},
		Target: Target{Template: Template{
			Data: make(map[string]*string),
		}},
		Data: make([]SingleData, 0),
	}
}

// Parse k8s secret string data and data into non base64 data
func parseStringDataAndData(secretManifest map[string]interface{}) map[string]interface{} {
	// If secretData just use it
	data := make(map[string]interface{})
	if secretManifest["stringData"] != nil {
		data = secretManifest["stringData"].(map[string]interface{})
	}
	// If data, convert all from base64 to ordinary strings
	if secretManifest["data"] != nil {
		for k, v := range secretManifest["data"].(map[string]interface{}) {
			decodedBytes, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				errors.Wrap(err, "Could not decode base64 string")
				return data
			}
			data[k] = string(decodedBytes)
		}
	}
	return data
}

// Get keyvault variables from a data map
func getKeyvaultVariables(data map[string]interface{}) []string {
	var keyvaultKeys []string

	for _, v := range data {
		// Search for {{ SOMETHING }}
		r := regexp.MustCompile(`\{{([^}]+)\}}`)
		if v == nil {
			continue
		}
		matches := r.FindAllString(v.(string), -1)
		for _, match := range matches {
			formatted := match[2 : len(match)-2]
			words := strings.Fields(formatted)
			// Only save keys matching {{ .VALUE }}
			for _, w := range words {
				if w[0] == '.' {
					keyvaultKeys = append(keyvaultKeys, w)
					break
				}
			}
		}

	}
	return keyvaultKeys
}

// Create an externalsecrets object from the secret and parsed data
func createExternalSecretObject(data, doc map[string]interface{}, keyvaultKeys []string) map[string]interface{} {
	// Create external secret spec
	spec := createBasicExternalSecretSpec()

	// Create template data section with values
	for k, v := range data {
		if v == nil {
			spec.Target.Template.Data[k] = nil
		} else {
			tmp := v.(string)
			spec.Target.Template.Data[k] = &tmp
		}
	}

	// Create data remote reference with values
	for _, key := range keyvaultKeys {
		if !strings.HasPrefix(key, ".") {
			panic("ERROR: Key is missing '.' prefix '" + key + "'")
		}
		spec.Data = append(spec.Data, SingleData{
			SecretKey: key[1:],
			RemoteRef: RemoteRef{Key: key[1:]},
		})
	}

	// Set k8s external secrets object info
	doc["kind"] = "ExternalSecret"
	doc["apiVersion"] = "external-secrets.io/v1beta1"
	delete(doc, "data")
	delete(doc, "stringData")
	doc["spec"] = spec
	delete(doc, "type")

	sort.Slice(spec.Data, func(i, j int) bool { return spec.Data[i].SecretKey < spec.Data[j].SecretKey })
	return doc
}

func transformManifest(doc map[string]interface{}) map[string]interface{} {
	kind := doc["kind"].(string)
	apiVersion := doc["apiVersion"].(string)
	// If it is a secret, try to convert it
	if kind == "Secret" && apiVersion == "v1" {
		// Convert data from base64 and merge stringData with data
		data := parseStringDataAndData(doc)

		// Find all keyvault variables
		keyvaultKeys := getKeyvaultVariables(data)

		// Only convert the secrets with references to keyvault secrets
		if len(keyvaultKeys) > 0 {
			return createExternalSecretObject(data, doc, keyvaultKeys)
		}
		// If kustomize ResourceList
	} else if apiVersion == "config.kubernetes.io/v1" && kind == "ResourceList" {
		items := doc["items"].([]interface{})
		// Transform all items
		for i, item := range items {
			resource := item.(map[string]interface{})
			items[i] = transformManifest(resource)
		}
		doc["items"] = items
	}
	return doc
}

func TransformFullManifest(dec *yaml.Decoder) string {
	manifest := ""
	// For each yaml document
	for {
		var doc map[string]interface{}
		if dec.Decode(&doc) != nil {
			break
		}
		if doc == nil {
			continue
		}

		doc = transformManifest(doc)
		// Create output document
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		yamlEncoder.Encode(&doc)
		manifest += b.String()
		manifest += "---\n"
	}
	return manifest
}

func main() {
	stdinBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic("ERROR: Could not read ResourceList from STDIN.")
	}
	dec := yaml.NewDecoder(bytes.NewReader(stdinBytes))
	manifest := TransformFullManifest(dec)
	fmt.Println(manifest)
}
