package obs

import (
	"strings"
)

const GlobToken = "*?\\![]{}"

type globSegment struct {
	segment string
	isGlob  bool
}

func splitGlobSegments(path string) (segments []globSegment) {
	var key []string
	for s := range strings.SplitSeq(path, "/") {
		if strings.ContainsAny(s, GlobToken) {
			if len(key) > 0 {
				gs := globSegment{isGlob: false}
				if len(segments) < 1 && strings.HasPrefix(path, "/") {
					gs.segment = "/" + strings.Join(key, "/")
				} else {
					gs.segment = strings.Join(key, "/")
				}
				segments = appendGlobSegment(segments, gs)
				key = nil
			}
			segments = appendGlobSegment(segments, globSegment{s, true})
		} else if s != "" {
			key = append(key, s)
		}
	}
	if len(key) > 0 {
		s := strings.Join(key, "/")
		segments = appendGlobSegment(segments, globSegment{s, false})
	}
	if strings.HasSuffix(path, "/") {
		segments[len(segments)-1].segment += "/"
	}
	return
}

func appendGlobSegment(keys []globSegment, segment globSegment) []globSegment {
	if len(keys) > 0 {
		k := &keys[len(keys)-1]
		if !strings.HasSuffix(k.segment, "/") {
			k.segment += "/"
		}
	}
	return append(keys, segment)
}
