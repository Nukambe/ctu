package rtsp_feeds

import (
	"fmt"
	"log"
)

func KillFFmpeg(feeds map[string]Feed) []error {
	var errs []error

	for id, feed := range feeds {
		log.Printf("Stopping FFmpeg for stream: %s", id)
		// Kill the FFmpeg process
		err := feed.cmd.Process.Kill()
		if err != nil {
			errs = append(errs, fmt.Errorf("error stopping [feed %s] with [PID %d]: %v", id, feed.cmd.Process.Pid, err))
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
