package youtube

import (
	"errors"
	"regexp"
)

var ErrInvalidURL = errors.New("invalid or unrecognized YouTube URL")

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`youtu\.be/([a-zA-Z0-9_-]{11})`),           // youtu.be/xxxxxxxxxxx
	regexp.MustCompile(`youtube\.com/watch\?v=([a-zA-Z0-9_-]{11})`), // youtube.com/watch?v=xxxxxxxxxxx
	regexp.MustCompile(`youtube\.com/embed/([a-zA-Z0-9_-]{11})`),    // youtube.com/embed/xxxxxxxxxxx
	regexp.MustCompile(`youtube\.com/shorts/([a-zA-Z0-9_-]{11})`),   // youtube.com/shorts/xxxxxxxxxxx
	regexp.MustCompile(`music\.youtube\.com/watch\?v=([a-zA-Z0-9_-]{11})`), // music.youtube.com
}

func ExtractID(rawURL string) (string, error) {
	for _, p := range patterns {
		if match := p.FindStringSubmatch(rawURL); len(match) > 1 {
			return match[1], nil
		}
	}
	return "", ErrInvalidURL
}