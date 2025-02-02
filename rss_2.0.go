package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"
)

func parseRSS2(data []byte) (*Feed, error) {
	warnings := false
	feed := rss20Feed{}
	p := xml.NewDecoder(bytes.NewReader(data))
	p.CharsetReader = charsetReader
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}
	if feed.Channel == nil {
		return nil, fmt.Errorf("no channel found in %q", string(data))
	}

	channel := feed.Channel

	out := new(Feed)
	out.Title = channel.Title
	out.Language = channel.Language
	out.Author = channel.Author
	out.Description = channel.Description
	out.Categories = channel.Categories.toArray()
	out.Link = extractLink(channel.Link)
	out.Image = channel.Image.Image()
	if channel.MinsToLive != 0 {
		sort.Ints(channel.SkipHours)
		next := time.Now().Add(time.Duration(channel.MinsToLive) * time.Minute)
		for _, hour := range channel.SkipHours {
			if hour == next.Hour() {
				next = next.Add(time.Duration(60-next.Minute()) * time.Minute)
			}
		}
		trying := true
		for trying {
			trying = false
			for _, day := range channel.SkipDays {
				if strings.Title(day) == next.Weekday().String() {
					next = next.Add(time.Duration(24-next.Hour()) * time.Hour)
					trying = true
					break
				}
			}
		}

		out.Refresh = next
	}

	if out.Refresh.IsZero() {
		out.Refresh = time.Now().Add(DefaultRefreshInterval)
	}

	out.Items = make([]*Item, 0, len(channel.Items))
	out.ItemMap = make(map[string]struct{})

	// Process items.
	for _, item := range channel.Items {

		if item.ID == "" {
			link := extractLink(item.Link)
			if link == "" {
				if debug {
					fmt.Printf("[w] Item %q has no ID or link and will be ignored.\n", item.Title)
					fmt.Printf("[w] %#v\n", item)
				}
				warnings = true
				continue
			}
			item.ID = link
		}

		// Skip items already known.
		if _, ok := out.ItemMap[item.ID]; ok {
			continue
		}

		next := new(Item)
		next.Title = item.Title
		next.Summary = item.Description
		next.Content = item.Content
		next.Categories = item.Categories
		next.Link = extractLink(item.Link)
		next.Image = item.Image.Image()
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err == nil {
				next.DateValid = true
			}
		} else if item.PubDate != "" {
			next.Date, err = parseTime(item.PubDate)
			if err == nil {
				next.DateValid = true
			}
		}
		next.ID = item.ID

		// Also convert `media:thumbnail` entries into enclosures
		hasMedia := item.MediaContent != mrssContent{} || item.MediaThumbnail != mrssThumbnail{}

		if len(item.Enclosures) > 0 || hasMedia {
			encLen := len(item.Enclosures)
			if hasMedia {
				encLen += 1
			}

			next.Enclosures = make([]*Enclosure, encLen)
			for i := range item.Enclosures {
				next.Enclosures[i] = item.Enclosures[i].Enclosure()
			}

			if hasMedia {
				next.Enclosures[len(item.Enclosures)] = item.MediaThumbnail.ToEnclosure()
			}
		}
		next.Read = false

		out.Items = append(out.Items, next)
		out.ItemMap[next.ID] = struct{}{}
		out.Unread++
	}

	if warnings && debug {
		fmt.Printf("[i] Encountered warnings:\n%s\n", data)
	}

	return out, nil
}

func extractLink(links []rss20Link) string {
	for _, link := range links {
		if link.Rel == "" && link.Type == "" && link.Href == "" && link.Chardata != "" {
			return link.Chardata
		}
	}
	return ""
}

type rss20Feed struct {
	XMLName xml.Name      `xml:"rss"`
	Channel *rss20Channel `xml:"channel"`
}

type rss20Category struct {
	XMLName xml.Name `xml:"category"`
	Name    string   `xml:"text,attr"`
}

type rss20CategorySlice []rss20Category

func (r rss20CategorySlice) toArray() (result []string) {
	count := len(r)
	if count == 0 || r == nil {
		return
	}
	result = make([]string, count)
	for i := range r {
		result[i] = r[i].Name
	}
	return
}

type rss20Channel struct {
	XMLName     xml.Name           `xml:"channel"`
	Title       string             `xml:"title"`
	Language    string             `xml:"language"`
	Author      string             `xml:"author"`
	Description string             `xml:"description"`
	Link        []rss20Link        `xml:"link"`
	Image       rss20Image         `xml:"image"`
	Categories  rss20CategorySlice `xml:"category"`
	Items       []rss20item        `xml:"item"`
	MinsToLive  int                `xml:"ttl"`
	SkipHours   []int              `xml:"skipHours>hour"`
	SkipDays    []string           `xml:"skipDays>day"`
}

type rss20Link struct {
	Rel      string `xml:"rel,attr"`
	Href     string `xml:"href,attr"`
	Type     string `xml:"type,attr"`
	Chardata string `xml:",chardata"`
}

type rss20Categories []string

type rss20item struct {
	XMLName     xml.Name        `xml:"item"`
	Title       string          `xml:"title"`
	Description string          `xml:"description"`
	Content     string          `xml:"encoded"`
	Categories  rss20Categories `xml:"category"`
	PubDate     string          `xml:"pubDate"`
	Date        string          `xml:"date"`
	Image       rss20Image      `xml:"image"`
	Link        []rss20Link     `xml:"link"`
	DateValid   bool
	ID          string           `xml:"guid"`
	Enclosures  []rss20Enclosure `xml:"enclosure"`
	// Support for Yahoo Media RSS, see https://www.rssboard.org/media-rss.
	MediaGroup     mrssGroup     `xml:"http://search.yahoo.com/mrss/ group"`
	MediaContent   mrssContent   `xml:"http://search.yahoo.com/mrss/ content"`
	MediaThumbnail mrssThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail"`
}

type rss20Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Type    string   `xml:"type,attr"`
	Length  uint     `xml:"length,attr"`
}

func (r *rss20Enclosure) Enclosure() *Enclosure {
	out := new(Enclosure)
	out.URL = r.URL
	out.Type = r.Type
	out.Length = r.Length
	return out
}

type rss20Image struct {
	XMLName xml.Name `xml:"image"`
	Href    string   `xml:"href,attr"`
	Title   string   `xml:"title"`
	URL     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

func (i *rss20Image) Image() *Image {
	out := new(Image)
	out.Title = i.Title
	out.Href = i.Href
	out.URL = i.URL
	out.Height = uint32(i.Height)
	out.Width = uint32(i.Width)
	return out
}

// See https://www.rssboard.org/media-rss#media-group for details.
type mrssGroup struct {
	Contents []mrssContent `xml:"http://search.yahoo.com/mrss/ content"`
}

// See https://www.rssboard.org/media-rss#media-content for details.
type mrssContent struct {
	URL          string `xml:"url,attr"`
	FileSize     uint   `xml:"fileSize,attr"`
	Type         string `xml:"type,attr"`
	Medium       string `xml:"medium,attr"`
	IsDefault    string `xml:"isDefault,attr"`
	Expression   string `xml:"expression,attr"`
	Bitrate      uint   `xml:"bitrate,attr"`
	Framerate    uint   `xml:"framerate,attr"`
	SamplingRate uint   `xml:"samplingrate,attr"`
	Channels     uint   `xml:"channels,attr"`
	Duration     uint   `xml:"duration,attr"`
	Height       uint   `xml:"height,attr"`
	Width        uint   `xml:"width,attr"`
	Lang         string `xml:"lang,attr"`
}

func (r *mrssContent) ToEnclosure() *Enclosure {
	out := new(Enclosure)
	out.URL = r.URL
	out.Type = r.Type
	out.Length = r.FileSize
	return out
}

// See https://www.rssboard.org/media-rss#media-thumbnails for details.
type mrssThumbnail struct {
	URL    string `xml:"url,attr"`
	Height uint   `xml:"height,attr"`
	Width  uint   `xml:"width,attr"`
	Time   string `xml:"time,attr"`
}

func (r *mrssThumbnail) ToEnclosure() *Enclosure {
	ext := r.URL[strings.LastIndex(r.URL, ".")+1:]

	out := new(Enclosure)
	out.URL = r.URL
	// This is a heuristic since the MIME type isn't actually specified
	out.Type = fmt.Sprintf("image/%s", ext)
	// This is incorrect since the length (in bytes) is not specified
	out.Length = 0
	return out
}
