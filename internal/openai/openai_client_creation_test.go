// Package openai provides a client for interacting with the OpenAI API
package openai

import (
	"testing"
)

// TestClientCreation tests basic client creation with API key
func TestClientCreation(t *testing.T) {
	// Create a client with a valid API key and model name
	client, err := NewClient("test-api-key", "gpt-4", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Check the model name is set correctly
	if client.GetModelName() != "gpt-4" {
		t.Errorf("Expected model name to be gpt-4, got %s", client.GetModelName())
	}
}

// TestClientCreationWithCustomBaseURL tests client creation with a custom base URL
func TestClientCreationWithCustomBaseURL(t *testing.T) {
	// Create a client with a valid API key, model name, and custom base URL
	client, err := NewClient("test-api-key", "gpt-4", "https://custom-openai-endpoint.example.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
}

// TestEmptyAPIKeyHandling tests how the client handles empty API keys
func TestEmptyAPIKeyHandling(t *testing.T) {
	// Create a client with an empty API key
	_, err := NewClient("", "gpt-4", "")
	if err == nil {
		t.Fatal("Expected error for empty API key, got nil")
	}
	if err.Error() != "API key is required" {
		t.Errorf("Expected error message 'API key is required', got '%s'", err.Error())
	}
}

// TestEmptyModelNameHandling tests how the client handles empty model names
func TestEmptyModelNameHandling(t *testing.T) {
	// Create a client with an empty model name
	_, err := NewClient("test-api-key", "", "")
	if err == nil {
		t.Fatal("Expected error for empty model name, got nil")
	}
	if err.Error() != "model name is required" {
		t.Errorf("Expected error message 'model name is required', got '%s'", err.Error())
	}
}
