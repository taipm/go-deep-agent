package agent

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
)

// ImageDetail specifies the detail level of image understanding
type ImageDetail string

const (
	ImageDetailAuto ImageDetail = "auto" // Default: let model decide
	ImageDetailLow  ImageDetail = "low"  // Faster, less tokens
	ImageDetailHigh ImageDetail = "high" // More detailed analysis
)

// ImageContent represents an image to be sent to the model
type ImageContent struct {
	URL    string      // URL or base64 data URI
	Detail ImageDetail // Level of detail for image analysis
}

// WithImage adds an image URL to the current message (simple version with auto detail).
// For more control, use WithImageURL to specify detail level.
//
// Example:
//
//	builder.WithImage("https://example.com/image.jpg")
func (b *Builder) WithImage(url string) *Builder {
	return b.WithImageURL(url, ImageDetailAuto)
}

// WithImageURL adds an image URL with specified detail level.
// Detail can be "auto" (default), "low" (faster), or "high" (more detailed).
//
// Note: Images are added to the next Ask/Stream call.
// For persistent images across conversation, add them to system message or use WithMessages.
//
// Example:
//
//	builder.WithImageURL("https://example.com/chart.jpg", agent.ImageDetailHigh)
func (b *Builder) WithImageURL(url string, detail ImageDetail) *Builder {
	if b.pendingImages == nil {
		b.pendingImages = []ImageContent{}
	}
	b.pendingImages = append(b.pendingImages, ImageContent{
		URL:    url,
		Detail: detail,
	})
	return b
}

// WithImageFile reads an image file and adds it as base64 encoded data.
// Supports common image formats: jpeg, jpg, png, gif, webp.
//
// Example:
//
//	builder.WithImageFile("./photo.jpg", agent.ImageDetailHigh)
func (b *Builder) WithImageFile(filePath string, detail ImageDetail) *Builder {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Store error to return later in Ask/Stream
		b.lastError = fmt.Errorf("failed to read image file: %w", err)
		return b
	}

	// Detect mime type from extension
	mimeType := detectImageMimeType(filePath)
	if mimeType == "" {
		b.lastError = fmt.Errorf("unsupported image file type: %s", filePath)
		return b
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	return b.WithImageURL(dataURI, detail)
}

// WithImageBase64 adds a base64-encoded image with specified MIME type.
// MIME type should be one of: image/jpeg, image/png, image/gif, image/webp.
//
// Example:
//
//	base64Data := "iVBORw0KGgoAAAANSUhEUg..." // your base64 string
//	builder.WithImageBase64(base64Data, "image/png", agent.ImageDetailAuto)
func (b *Builder) WithImageBase64(base64Data, mimeType string, detail ImageDetail) *Builder {
	// Validate mime type
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validTypes[mimeType] {
		b.lastError = fmt.Errorf("unsupported image MIME type: %s", mimeType)
		return b
	}

	// Create data URI
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	return b.WithImageURL(dataURI, detail)
}

// ClearImages removes all pending images.
// This is useful if you want to send multiple messages with different images.
//
// Example:
//
//	builder.WithImage("image1.jpg")
//	response1, _ := builder.Ask(ctx, "What's in this image?")
//	builder.ClearImages() // Remove image1.jpg
//	builder.WithImage("image2.jpg")
//	response2, _ := builder.Ask(ctx, "What's in this image?")
func (b *Builder) ClearImages() *Builder {
	b.pendingImages = nil
	return b
}

// detectImageMimeType detects MIME type from file extension
func detectImageMimeType(filePath string) string {
	ext := strings.ToLower(filePath)
	if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(ext, ".png") {
		return "image/png"
	} else if strings.HasSuffix(ext, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(ext, ".webp") {
		return "image/webp"
	}
	// Default to jpeg for unknown formats
	return "image/jpeg"
}

// buildContentParts creates content parts for multimodal messages.
// Returns either a simple string or an array of content parts (text + images).
func (b *Builder) buildContentParts(text string) interface{} {
	// If no images, return simple text
	if len(b.pendingImages) == 0 {
		return text
	}

	// Build array of content parts: [text, image1, image2, ...]
	parts := []openai.ChatCompletionContentPartUnionParam{}

	// Add text part first
	if text != "" {
		parts = append(parts, openai.TextContentPart(text))
	}

	// Add image parts
	for _, img := range b.pendingImages {
		imageURL := openai.ChatCompletionContentPartImageImageURLParam{
			URL:    img.URL,
			Detail: string(img.Detail),
		}
		parts = append(parts, openai.ImageContentPart(imageURL))
	}

	return parts
}
