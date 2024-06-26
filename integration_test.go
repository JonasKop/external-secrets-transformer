package main

import (
	"bytes"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEmptyInputProducesEmptyResult(t *testing.T) {
	dec := yaml.NewDecoder(bytes.NewReader([]byte("")))
	manifest := TransformFullManifest(dec)
	if len(manifest) != 0 {
		t.Fatalf("Expected manifest to be empty")
	}
}

func TestFullManifest(t *testing.T) {
	testFile := "testresources/full_manifest.yaml"
	testManifestBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, testFile, err)
	}

	resultFile := "testresources/full_manifest_transformed.yaml"
	resultManifestBytes, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, resultFile, err)
	}

	os.Setenv("STORE_NAME", "my-test-store")
	os.Setenv("STORE_KIND", "ClusterSecretStore")

	dec := yaml.NewDecoder(bytes.NewReader(testManifestBytes))
	manifest := TransformFullManifest(dec)

	if manifest != string(resultManifestBytes) {
		t.Fatalf("Manifest mismatch")
	}
}

func TestResourceList(t *testing.T) {
	testFile := "testresources/resource_list.yaml"
	testManifestBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, testFile, err)
	}

	resultFile := "testresources/resource_list_transformed.yaml"
	resultManifestBytes, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, resultFile, err)
	}

	os.Setenv("STORE_NAME", "my-test-store")
	os.Setenv("STORE_KIND", "ClusterSecretStore")

	dec := yaml.NewDecoder(bytes.NewReader(testManifestBytes))
	manifest := TransformFullManifest(dec)

	if manifest != string(resultManifestBytes) {
		t.Fatalf("Manifest mismatch")
	}
}

func TestDuplicate(t *testing.T) {
	testFile := "testresources/duplicate_secret.yaml"
	testManifestBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, testFile, err)
	}

	resultFile := "testresources/duplicate_secret_transformed.yaml"
	resultManifestBytes, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, resultFile, err)
	}

	os.Setenv("STORE_NAME", "my-test-store")
	os.Setenv("STORE_KIND", "ClusterSecretStore")

	dec := yaml.NewDecoder(bytes.NewReader(testManifestBytes))
	manifest := TransformFullManifest(dec)

	if manifest != string(resultManifestBytes) {
		t.Fatalf("Manifest mismatch")
	}
}

func TestInvalidKeyvaultSecretRefShouldNotTransform(t *testing.T) {
	testFile := "testresources/invalid_keyvault_secret.yaml"
	testManifestBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf(`Could not read file %s: %v`, testFile, err)
	}

	os.Setenv("STORE_NAME", "my-test-store")
	os.Setenv("STORE_KIND", "ClusterSecretStore")

	dec := yaml.NewDecoder(bytes.NewReader(testManifestBytes))
	manifest := TransformFullManifest(dec)
	if manifest != string(testManifestBytes) {
		t.Fatalf("Manifest mismatch")
	}
}
