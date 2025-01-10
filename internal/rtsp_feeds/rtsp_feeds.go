package rtsp_feeds

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Feed struct {
	url string
	cmd *exec.Cmd
}

func GetFeeds(hlsDir string) map[string]Feed {
	main := os.Getenv("FEED_MAIN")
	//main_sub := os.Getenv("FEED_MAIN_SUB")

	return map[string]Feed{
		"main": {
			url: main,
			cmd: captureFeed(main, fmt.Sprintf("%s/%s", hlsDir, "main")),
		},
		//"main_sub": {
		//	url: os.Getenv("FEED_MAIN_SUB"),
		//},
	}
}

func captureFeed(url, id string) *exec.Cmd {
	// Create directory for HLS if it doesn't exist
	err := os.MkdirAll(id, 0755)
	if err != nil {
		log.Fatalf("Error creating HLS directory: %v", err)
	}

	// Run FFmpeg to capture RTSP and transcode to HLS
	cmd := exec.Command("ffmpeg",
		"-hide_banner", "-y",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-i", url,
		"-c:v", "libx264", // Use H.264 encoding for better browser compatibility
		"-preset", "veryfast",
		"-c:a", "aac", // Audio codec
		"-f", "hls", // Output format HLS
		"-hls_time", "4", // Segment duration in seconds
		"-hls_list_size", "10", // Keep only the latest 10 segments in the playlist
		"-hls_flags", "delete_segments", // delete old segments
		"-hls_segment_filename", id+"/%03d.ts", // Segment filename pattern
		id+"/index.m3u8", // Playlist filename
	)

	// Start FFmpeg process
	err = cmd.Start()
	if err != nil {
		log.Fatalf("Error starting FFmpeg process: %v", err)
	}

	log.Printf("FFmpeg is transcoding the RTSP stream [%s] to HLS.", id)

	return cmd
}
