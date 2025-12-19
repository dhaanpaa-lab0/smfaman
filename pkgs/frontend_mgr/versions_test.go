package frontend_mgr

import (
	"testing"
)

func TestFetchCdnjsVersions(t *testing.T) {
	// Test with a well-known package
	result, err := FetchCdnjsVersions("jquery")
	if err != nil {
		t.Fatalf("failed to fetch versions from CDNJS: %v", err)
	}

	if result.Name != "jquery" {
		t.Errorf("expected package name 'jquery', got %q", result.Name)
	}

	if len(result.Versions) == 0 {
		t.Error("expected at least one version, got none")
	}

	if result.Version == "" {
		t.Error("expected latest version to be set")
	}

	t.Logf("CDNJS - Package: %s, Latest: %s, Total versions: %d", result.Name, result.Version, len(result.Versions))
}

func TestFetchJsdelivrVersions(t *testing.T) {
	// Test with a well-known package
	result, err := FetchJsdelivrVersions("react")
	if err != nil {
		t.Fatalf("failed to fetch versions from jsDelivr: %v", err)
	}

	if result.Name != "react" {
		t.Errorf("expected package name 'react', got %q", result.Name)
	}

	if len(result.Versions) == 0 {
		t.Error("expected at least one version, got none")
	}

	if result.Tags["latest"] == "" {
		t.Error("expected latest tag to be set")
	}

	t.Logf("jsDelivr - Package: %s, Latest: %s, Total versions: %d", result.Name, result.Tags["latest"], len(result.Versions))
}

func TestFetchUnpkgVersions(t *testing.T) {
	// Test with a well-known package
	result, err := FetchUnpkgVersions("lodash")
	if err != nil {
		t.Fatalf("failed to fetch versions from npm registry: %v", err)
	}

	if result.Name != "lodash" {
		t.Errorf("expected package name 'lodash', got %q", result.Name)
	}

	if len(result.Versions) == 0 {
		t.Error("expected at least one version, got none")
	}

	if result.DistTags["latest"] == "" {
		t.Error("expected latest dist-tag to be set")
	}

	t.Logf("UNPKG/npm - Package: %s, Latest: %s, Total versions: %d", result.Name, result.DistTags["latest"], len(result.Versions))
}

func TestSortVersions(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "basic semantic versions",
			input:    []string{"1.0.0", "2.0.0", "1.5.0", "1.0.1"},
			expected: []string{"2.0.0", "1.5.0", "1.0.1", "1.0.0"},
		},
		{
			name:     "versions with prerelease",
			input:    []string{"1.0.0", "1.0.0-alpha", "1.0.0-beta", "2.0.0"},
			expected: []string{"2.0.0", "1.0.0", "1.0.0-beta", "1.0.0-alpha"},
		},
		{
			name:     "mixed version formats",
			input:    []string{"3.0.0", "2.5.1", "3.0.1", "2.10.0"},
			expected: []string{"3.0.1", "3.0.0", "2.10.0", "2.5.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortVersions(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d versions, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("at index %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestSortVersionsWithInvalidVersions(t *testing.T) {
	input := []string{"1.0.0", "invalid", "2.0.0", "not-a-version", "1.5.0"}
	result := SortVersions(input)

	// Should only include valid versions
	expected := []string{"2.0.0", "1.5.0", "1.0.0"}

	if len(result) != len(expected) {
		t.Errorf("expected %d valid versions, got %d", len(expected), len(result))
		return
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("at index %d: expected %q, got %q", i, exp, result[i])
		}
	}
}
