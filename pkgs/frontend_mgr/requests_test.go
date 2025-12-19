package frontend_mgr

import (
	"testing"
)

func TestFetchUnpkgMeta(t *testing.T) {
	result, err := FetchUnpkgMeta("react", "18.2.0")
	if err != nil {
		t.Fatalf("Failed to fetch UNPKG metadata: %v", err)
	}

	if result.Package != "react" {
		t.Errorf("Expected package 'react', got '%s'", result.Package)
	}
	if result.Version != "18.2.0" {
		t.Errorf("Expected version '18.2.0', got '%s'", result.Version)
	}
	if len(result.Files) == 0 {
		t.Error("Expected files array to have items")
	}
}

func TestFetchCdnjsVersion(t *testing.T) {
	result, err := FetchCdnjsVersion("react", "18.2.0")
	if err != nil {
		t.Fatalf("Failed to fetch CDNJS version data: %v", err)
	}

	if result.Name != "react" {
		t.Errorf("Expected name 'react', got '%s'", result.Name)
	}
	if result.Version != "18.2.0" {
		t.Errorf("Expected version '18.2.0', got '%s'", result.Version)
	}
	if len(result.Files) == 0 {
		t.Error("Expected files array to have items")
	}
}

func TestFetchJsdelivrPackage(t *testing.T) {
	result, err := FetchJsdelivrPackage("react", "18.2.0")
	if err != nil {
		t.Fatalf("Failed to fetch jsDelivr package data: %v", err)
	}

	if result.Name != "react" {
		t.Errorf("Expected name 'react', got '%s'", result.Name)
	}
	if result.Version != "18.2.0" {
		t.Errorf("Expected version '18.2.0', got '%s'", result.Version)
	}
	if result.Type != "npm" {
		t.Errorf("Expected type 'npm', got '%s'", result.Type)
	}
	if len(result.Files) == 0 {
		t.Error("Expected files array to have items")
	}
}
