package media_analysis

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VideoProcessor handles the extraction of frames.
type VideoProcessor struct{}

// NewVideoProcessor creates a new VideoProcessor.
func NewVideoProcessor() *VideoProcessor {
	return &VideoProcessor{}
}

// ProcessVideo takes video data, extracts the first frame, and returns it as a JPEG byte slice.
func (vp *VideoProcessor) ProcessVideo(videoData []byte) ([]byte, error) {
	// Create a temporary directory for this operation
	tempDir, err := os.MkdirTemp("", "video-processing-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write video data to a temporary file
	tempVideoPath := filepath.Join(tempDir, "input.mp4")
	if err := os.WriteFile(tempVideoPath, videoData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp video file: %w", err)
	}

	// Extract the first frame from the video
	framePath := filepath.Join(tempDir, "frame.jpg")
	if err := vp.extractFirstFrame(tempVideoPath, framePath); err != nil {
		return nil, fmt.Errorf("failed to extract first frame: %w", err)
	}

	// Read the frame image data
	frameData, err := os.ReadFile(framePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read frame data: %w", err)
	}

	return frameData, nil
}

// extractFirstFrame uses ffmpeg to extract the first frame from a video file.
func (vp *VideoProcessor) extractFirstFrame(videoPath, outputPath string) error {
	// Command to extract the very first frame (-vframes 1)
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vframes", "1", "-q:v", "2", outputPath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg execution failed: %s\n%s", err, stderr.String())
	}

	return nil
}

func init() {
	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		fmt.Println("WARNING: ffmpeg not found in PATH. Video processing will not be available.")
	}
}
