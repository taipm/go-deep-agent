package agent

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test ImageDetail constants
func TestImageDetailConstants(t *testing.T) {
	tests := []struct {
		name     string
		detail   ImageDetail
		expected string
	}{
		{"Auto detail", ImageDetailAuto, "auto"},
		{"Low detail", ImageDetailLow, "low"},
		{"High detail", ImageDetailHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.detail) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.detail))
			}
		})
	}
}

// Test WithImage adds image with auto detail
func TestWithImage(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").WithImage("https://example.com/image.jpg")

	if len(builder.pendingImages) != 1 {
		t.Fatalf("Expected 1 pending image, got %d", len(builder.pendingImages))
	}

	img := builder.pendingImages[0]
	if img.URL != "https://example.com/image.jpg" {
		t.Errorf("Expected URL 'https://example.com/image.jpg', got '%s'", img.URL)
	}
	if img.Detail != ImageDetailAuto {
		t.Errorf("Expected detail 'auto', got '%s'", img.Detail)
	}
}

// Test WithImageURL adds image with specific detail
func TestWithImageURL(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		detail ImageDetail
	}{
		{"Low detail", "https://example.com/low.jpg", ImageDetailLow},
		{"High detail", "https://example.com/high.jpg", ImageDetailHigh},
		{"Auto detail", "https://example.com/auto.jpg", ImageDetailAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewOpenAI("gpt-4o", "test-key").WithImageURL(tt.url, tt.detail)

			if len(builder.pendingImages) != 1 {
				t.Fatalf("Expected 1 pending image, got %d", len(builder.pendingImages))
			}

			img := builder.pendingImages[0]
			if img.URL != tt.url {
				t.Errorf("Expected URL '%s', got '%s'", tt.url, img.URL)
			}
			if img.Detail != tt.detail {
				t.Errorf("Expected detail '%s', got '%s'", tt.detail, img.Detail)
			}
		})
	}
}

// Test WithImageBase64 adds base64 image
func TestWithImageBase64(t *testing.T) {
	base64Data := base64.StdEncoding.EncodeToString([]byte("fake image data"))
	mimeType := "image/jpeg"

	builder := NewOpenAI("gpt-4o", "test-key").WithImageBase64(base64Data, mimeType, ImageDetailHigh)

	if len(builder.pendingImages) != 1 {
		t.Fatalf("Expected 1 pending image, got %d", len(builder.pendingImages))
	}

	img := builder.pendingImages[0]
	expectedURL := "data:image/jpeg;base64," + base64Data
	if img.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, img.URL)
	}
	if img.Detail != ImageDetailHigh {
		t.Errorf("Expected detail 'high', got '%s'", img.Detail)
	}
}

// Test multiple images
func TestMultipleImages(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").
		WithImage("https://example.com/image1.jpg").
		WithImageURL("https://example.com/image2.jpg", ImageDetailLow).
		WithImage("https://example.com/image3.jpg")

	if len(builder.pendingImages) != 3 {
		t.Fatalf("Expected 3 pending images, got %d", len(builder.pendingImages))
	}

	// Check first image
	if builder.pendingImages[0].URL != "https://example.com/image1.jpg" {
		t.Errorf("First image URL mismatch")
	}
	if builder.pendingImages[0].Detail != ImageDetailAuto {
		t.Errorf("First image should have auto detail")
	}

	// Check second image
	if builder.pendingImages[1].URL != "https://example.com/image2.jpg" {
		t.Errorf("Second image URL mismatch")
	}
	if builder.pendingImages[1].Detail != ImageDetailLow {
		t.Errorf("Second image should have low detail")
	}

	// Check third image
	if builder.pendingImages[2].URL != "https://example.com/image3.jpg" {
		t.Errorf("Third image URL mismatch")
	}
}

// Test ClearImages
func TestClearImages(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").
		WithImage("https://example.com/image1.jpg").
		WithImage("https://example.com/image2.jpg").
		ClearImages()

	if len(builder.pendingImages) != 0 {
		t.Errorf("Expected 0 pending images after clear, got %d", len(builder.pendingImages))
	}
}

// Test detectImageMimeType
func TestDetectImageMimeType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"JPEG", "photo.jpg", "image/jpeg"},
		{"JPEG uppercase", "photo.JPEG", "image/jpeg"},
		{"PNG", "image.png", "image/png"},
		{"GIF", "animation.gif", "image/gif"},
		{"WebP", "modern.webp", "image/webp"},
		{"Unknown", "file.bmp", "image/jpeg"}, // defaults to jpeg
		{"No extension", "noext", "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectImageMimeType(tt.filePath)
			if result != tt.expected {
				t.Errorf("Expected MIME type '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Test buildContentParts with text only (no images)
func TestBuildContentPartsTextOnly(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key")
	result := builder.buildContentParts("Hello, world!")

	// Should return string for text-only
	str, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result for text-only, got %T", result)
	}

	if str != "Hello, world!" {
		t.Errorf("Expected 'Hello, world!', got '%s'", str)
	}
}

// Test buildContentParts with text and images
func TestBuildContentPartsWithImages(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").
		WithImage("https://example.com/image1.jpg").
		WithImageURL("https://example.com/image2.jpg", ImageDetailHigh)

	result := builder.buildContentParts("Analyze these images")

	// Verify it's not a string
	if _, ok := result.(string); ok {
		t.Fatal("Expected slice type for multimodal content, got string")
	}

	// Verify it's a slice by checking interface conversion
	// (Detailed type checking requires reflection - just verify slice behavior)
	if result == nil {
		t.Fatal("buildContentParts returned nil")
	}
}

// Test WithImageFile with invalid file
func TestWithImageFileInvalidFile(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").WithImageFile("/nonexistent/file.jpg", ImageDetailAuto)

	// Should have stored error in lastError
	if builder.lastError == nil {
		t.Error("Expected error for nonexistent file")
	}

	// Should not have added image
	if len(builder.pendingImages) != 0 {
		t.Errorf("Expected 0 pending images for invalid file, got %d", len(builder.pendingImages))
	}
}

// Test WithImageFile with valid file
func TestWithImageFileValidFile(t *testing.T) {
	// Create temporary image file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.jpg")
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG header
	if err := os.WriteFile(tmpFile, imageData, 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	builder := NewOpenAI("gpt-4o", "test-key").WithImageFile(tmpFile, ImageDetailHigh)

	// Should not have error
	if builder.lastError != nil {
		t.Errorf("Unexpected error: %v", builder.lastError)
	}

	// Should have added image
	if len(builder.pendingImages) != 1 {
		t.Fatalf("Expected 1 pending image, got %d", len(builder.pendingImages))
	}

	img := builder.pendingImages[0]

	// URL should be data URI with base64
	if !strings.HasPrefix(img.URL, "data:image/jpeg;base64,") {
		t.Errorf("Expected data URI, got: %s", img.URL)
	}

	// Detail should be high
	if img.Detail != ImageDetailHigh {
		t.Errorf("Expected high detail, got: %s", img.Detail)
	}

	// Verify base64 content
	base64Part := strings.TrimPrefix(img.URL, "data:image/jpeg;base64,")
	decoded, err := base64.StdEncoding.DecodeString(base64Part)
	if err != nil {
		t.Errorf("Failed to decode base64: %v", err)
	}
	if len(decoded) != len(imageData) {
		t.Errorf("Decoded data length mismatch: expected %d, got %d", len(imageData), len(decoded))
	}
}

// Test error handling in Ask with lastError
func TestAskWithLastError(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").WithImageFile("/nonexistent/file.jpg", ImageDetailAuto)

	// Attempt Ask with error present
	resp, err := builder.Ask(context.Background(), "Test")

	// Should return the lastError
	if err == nil {
		t.Error("Expected error from Ask when lastError is set")
	}

	if resp != "" {
		t.Errorf("Expected empty response, got: %s", resp)
	}
}

// Test image clearing logic in buildMessages
func TestBuildMessagesClearsImagesLogic(t *testing.T) {
	builder := NewOpenAI("gpt-4o", "test-key").WithImage("https://example.com/image.jpg")

	// Verify image is pending
	if len(builder.pendingImages) != 1 {
		t.Fatal("Expected 1 pending image before buildMessages")
	}

	// Call buildMessages to get multimodal content
	result := builder.buildMessages("Test message")

	// Result should be messages array
	if result == nil {
		t.Fatal("buildMessages should return messages")
	}

	// Images still pending - actual clearing happens in Ask/Stream
	if len(builder.pendingImages) != 1 {
		t.Errorf("Images should still be pending after buildMessages (cleared by Ask/Stream)")
	}
}

// Test chaining multiple operations
func TestChainedMultimodalOperations(t *testing.T) {
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithImage("https://example.com/1.jpg").
		WithTemperature(0.7).
		WithImageURL("https://example.com/2.jpg", ImageDetailLow).
		WithMaxTokens(500)

	// Check images
	if len(builder.pendingImages) != 2 {
		t.Fatalf("Expected 2 pending images, got %d", len(builder.pendingImages))
	}

	// Clear and verify
	builder.ClearImages()
	if len(builder.pendingImages) != 0 {
		t.Error("Images not cleared")
	}
}
