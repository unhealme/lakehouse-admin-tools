package obs

import (
	"strings"
)

const GlobToken = "*?\\![]{}"

type globSegment struct {
	segment string
	isGlob  bool
}

func splitGlobSegments(path string) (splitKeys []globSegment) {
	var keySegment []string
	for segment := range strings.SplitSeq(path, "/") {
		if !strings.ContainsAny(segment, GlobToken) {
			keySegment = append(keySegment, segment)
		} else {
			if len(keySegment) > 0 {
				gs := globSegment{isGlob: false}
				if len(splitKeys) < 1 && strings.HasPrefix(path, "/") {
					gs.segment = "/" + strings.Join(keySegment, "/")
				} else {
					gs.segment = strings.Join(keySegment, "/")
				}
				splitKeys = appendGlobSegment(splitKeys, gs)
				keySegment = nil
			}
			splitKeys = appendGlobSegment(splitKeys, globSegment{segment, true})
		}
	}
	if len(keySegment) > 0 {
		splitKeys = appendGlobSegment(splitKeys, globSegment{strings.Join(keySegment, "/"), false})
	}
	return
}

func appendGlobSegment(keys []globSegment, segment globSegment) []globSegment {
	if len(keys) > 0 {
		keys[len(keys)-1].segment += "/"
	}
	keys = append(keys, segment)
	return keys
}
