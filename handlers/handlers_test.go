package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetBlockReward_InvalidSlot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	router := gin.New()
	handler := &Handler{
		// In a real test, you would inject a mock client here
		client: nil,
	}
	router.GET("/blockreward/:slot", handler.GetBlockReward)

	// Test invalid slot format
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/blockreward/invalid", nil)
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.Error != "Invalid slot number" {
		t.Errorf("Expected error message 'Invalid slot number', got '%s'", response.Error)
	}

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected error code %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestErrorResponse_JSONFormat(t *testing.T) {
	response := ErrorResponse{
		Error: "Test error",
		Code:  400,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal ErrorResponse: %v", err)
	}

	var decoded ErrorResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal ErrorResponse: %v", err)
	}

	if decoded.Error != response.Error {
		t.Errorf("Expected error '%s', got '%s'", response.Error, decoded.Error)
	}

	if decoded.Code != response.Code {
		t.Errorf("Expected code %d, got %d", response.Code, decoded.Code)
	}
}

func TestBlockRewardResponse_JSONFormat(t *testing.T) {
	response := BlockRewardResponse{
		Status: "MEV",
		Reward: 123456,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal BlockRewardResponse: %v", err)
	}

	var decoded BlockRewardResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Errorf("Failed to unmarshal BlockRewardResponse: %v", err)
	}

	if decoded.Status != response.Status {
		t.Errorf("Expected status '%s', got '%s'", response.Status, decoded.Status)
	}

	if decoded.Reward != response.Reward {
		t.Errorf("Expected reward %d, got %d", response.Reward, decoded.Reward)
	}
}
