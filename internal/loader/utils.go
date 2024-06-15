package loader

import (
	"path/filepath"
	"strings"
)

// DirBase behavior:
//
//		             path  |       dir  | base
//		-------------------+------------+-----------
//		   {empty string}  |         .  |  .
//		                .  |         .  |  .
//		               ./  |         .  |  .
//	                 /  |         /  |  /
//		            ./foo  |         .  | foo
//		           ../foo  |        ..  | foo
//		             /foo  |         /  | foo
//		   /usr/local/foo  | /usr/local | foo
func DirBase(path string) (dir, base string) {
	return filepath.Dir(path), filepath.Base(path)
}

func CommentBody(s string) string {
	const (
		begin = "<!--"
		end   = "-->"
	)
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, begin) {
		return ""
	}
	if !strings.HasSuffix(s, end) {
		return ""
	}
	return s[len(begin) : len(s)-len(end)]
}

func ParseLabels(s string) (result []Label) {
	const labelPrefixChar = uint8('@')
	items := strings.Split(s, " ")
	for _, word := range items {
		i := 0
		for i < len(word) && word[i] == labelPrefixChar {
			i++
		}
		if i > 0 && i < len(word) && word[i-1] == labelPrefixChar {
			result = append(result, Label(word[i:]))
		}
	}
	return
}
