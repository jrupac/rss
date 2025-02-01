package rss

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

func TestParseAtomTitle(t *testing.T) {
	tests := map[string]string{
		"atom_1.0":           "Titel des Weblogs",
		"atom_1.0_enclosure": "Titel des Weblogs",
		"atom_1.0-1":         "Golem.de",
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

		if feed.Title != want {
			t.Errorf("%s: got %q, want %q", name, feed.Title, want)
		}
	}
}

func TestParseAtomContent(t *testing.T) {
	tests := map[string]string{
		"atom_1.0":           "Volltext des Weblog-Eintrags",
		"atom_1.0_enclosure": "Volltext des Weblog-Eintrags",
		"atom_1.0-1":         "",
		"atom_1.0_html":      "<body>html</body>",
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
			t.Errorf("%s: got %q, want %q", name, feed.Items[0].Content, want)
		}

		if !feed.Items[0].DateValid {
			t.Errorf("%s: Invalid date: %q", name, feed.Items[0].Date)
		}
	}
}

func TestParseAtomDate(t *testing.T) {
	tests := map[string]string{
		"atom_1.0_html":           "2003-12-13T18:30:02Z",
		"atom_1.0_only_published": "2003-12-13T18:30:02Z",
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

		date := feed.Items[0].Date
		wantDate, err := time.Parse(time.RFC3339Nano, want)

		if date != wantDate {
			t.Errorf("%s: got %q, want %q", name, date, want)
		}

		if !feed.Items[0].DateValid {
			t.Errorf("%s: Invalid date: %q", name, feed.Items[0].Date)
		}
	}
}
