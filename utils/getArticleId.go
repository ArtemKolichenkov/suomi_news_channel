package utils

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

// Extract 74-20059886 from https://yle.fi/a/74-20059886
func GetYleArticleId(feedItem *gofeed.Item) string {
	stringParts := strings.Split(feedItem.GUID, "/")
	return strings.Split(feedItem.GUID, "/")[len(stringParts)-1]
}
