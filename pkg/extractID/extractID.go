package extractid

import	(
	"regexp"
)

func fullLinkParser(fullLink string) string {
	re := regexp.MustCompile(`(.*?)//(.*?)@(.*)`)
	match := re.FindStringSubmatch(fullLink)
	if len(match) >=3 {
		id := match[2]
		return id
	} else {
		return ""
	}
}