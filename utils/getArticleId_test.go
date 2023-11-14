package utils

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestGetYleArticleId(t *testing.T) {
	t.Run("returns ID from url", func(t *testing.T) {
		feedItem := &gofeed.Item{
			GUID: "https://yle.fi/a/74-20059886",
		}
		if GetYleArticleId(feedItem) != "74-20059886" {
			t.Error("GetYleArticleId did not return 74-20059886")
		}
	})

	t.Run("returns ID if GUID is already a plain ID", func(t *testing.T) {
		feedItem := &gofeed.Item{
			GUID: "74-20059886",
		}
		if GetYleArticleId(feedItem) != "74-20059886" {
			t.Error("GetYleArticleId did not return 74-20059886")
		}
	})
}
