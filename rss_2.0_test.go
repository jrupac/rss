package rss

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestParseItemLen(t *testing.T) {
	tests := map[string]int{
		"rss_2.0":                 2,
		"rss_2.0_content_encoded": 1,
		"rss_2.0_enclosure":       1,
		"rss_2.0-1":               4,
		"rss_2.0-1_enclosure":     1,
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if len(feed.Items) != want {
			t.Errorf("%s: got %d, want %d", name, len(feed.Items), want)
		}
	}
}

func TestParseContent(t *testing.T) {
	tests := map[string]string{
		"rss_2.0_content_encoded": "<p><a href=\"https://example.com/\">Example.com</a> is an example site.</p>",
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if feed.Items[0].Content != want {
			t.Errorf("%s: got %s, want %s", name, feed.Items[0].Content, want)
		}
	}
}

func TestParseItemDateOK(t *testing.T) {
	tests := map[string]string{
		"rss_2.0":                 "2009-09-06 16:45:00 +0000 UTC",
		"rss_2.0_content_encoded": "2009-09-06 16:45:00 +0000 UTC",
		"rss_2.0_enclosure":       "2009-09-06 16:45:00 +0000 UTC",
		"rss_2.0-1":               "2003-06-03 09:39:21 +0000 UTC",
		"rss_2.0-1_enclosure":     "2016-05-14 15:39:34 +0000 UTC",
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if !feed.Items[0].DateValid {
			t.Errorf("%s: date %q invalid!", name, feed.Items[0].Date)
		} else if got := feed.Items[0].Date.UTC().String(); got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestParseItemDateFailure(t *testing.T) {
	tests := map[string]string{
		"rss_2.0": "0001-01-01 00:00:00 +0000 UTC",
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if fmt.Sprintf("%s", feed.Items[1].Date) != want {
			t.Errorf("%s: got %q, want %q", name, feed.Items[1].Date, want)
		}

		if feed.Items[1].DateValid {
			t.Errorf("%s: got unexpected valid date", name)
		}
	}
}

func TestParseCategories(t *testing.T) {
	tests := map[string]int{
		"rss_2.0-1_enclosure": 2,
		"rss_2.0_enclosure":   0,
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if len(feed.Items[0].Categories) != want {
			t.Errorf("%s: got %q, want %q", name, feed.Items[0].Categories, want)
		}
	}
}

func TestParseChannelCategories(t *testing.T) {
	tests := map[string]int{
		"rss_2.0-1_enclosure": 2,
		"rss_2.0_enclosure":   1,
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if len(feed.Categories) != want {
			t.Errorf("%s: got %q, want %q", name, feed.Items[0].Categories, want)
		}
	}
}

func TestChannelProperties(t *testing.T) {
	tests := []struct {
		name     string
		testdata string
		verify   func(t *testing.T, feed *Feed)
	}{{
		name:     "normal case",
		testdata: "rss_2.0_content_encoded",
		verify: func(t *testing.T, feed *Feed) {
			assertEqual("en", feed.Language, t)
			assertEqual("someone", feed.Author, t)
		},
	}}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			name := filepath.Join("testdata", tt.testdata)
			data, err := ioutil.ReadFile(name)
			if err != nil {
				t.Fatalf("Reading %s: %v", name, err)
			}

			feed, err := Parse(data)
			if err != nil {
				t.Fatalf("Parsing %s: %v", name, err)
			}
			tt.verify(t, feed)
		})
	}
}

func assertEqual(expected, got string, t *testing.T) {
	if expected != got {
		t.Errorf("expect '%s', got '%s'", expected, got)
	}
}

func TestParseMultipleLinks(t *testing.T) {
	tests := map[string]string{
		"rss_2.0_links_single":   "link_a",
		"rss_2.0_links_multiple": "link_b",
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if feed.Items[0].Link != want {
			t.Errorf("%s: got %q, want %q", name, feed.Items[0].Link, want)
		}
	}
}

func TestRss2_0ParseMediaThumbnail(t *testing.T) {
	tests := map[string][]string{
		"rss_2.0_media_thumbnail": {"http://example.com/image.jpg", "image/jpg"},
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		enc := feed.Items[0].Enclosures[0]

		if enc.URL != want[0] {
			t.Errorf("%s: got %q, want %q", name, enc.URL, want[0])
		}

		if enc.Type != want[1] {
			t.Errorf("%s: got %q, want %q", name, enc.Type, want[1])
		}

	}
}