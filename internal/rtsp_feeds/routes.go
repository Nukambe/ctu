package rtsp_feeds

import (
	"encoding/json"
	"log"
	"net/http"
)

func HandleGetFeeds(feeds map[string]Feed) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		var feedIds []string
		for id, _ := range feeds {
			feedIds = append(feedIds, id)
		}

		data, err := json.Marshal(feedIds)
		if err != nil {
			log.Printf("Error marshalling feed Ids: %v", err)
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(500)
			_, _ = res.Write([]byte("Server error"))
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(200)
		_, _ = res.Write(data)
	}
}
