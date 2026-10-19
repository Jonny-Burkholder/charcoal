package convert

import (
	"strings"
)

const (
	capitalA   = 65
	capitalZ   = 90
	underscore = 95
)

// ToSnakeCase takes a string and returns a string.
// Its purpose is to take a string of any casing, and
// convert it to snake_case
func ToSnakeCase(s string) string {
	// trim space
	s = strings.TrimSpace(s)

	// insert _ before capitals and
	// convert capitals to lowercase
	s = lowercaseAndUnderscore(s)

	// convert spaces to underscore
	return strings.ReplaceAll(s, " ", "_")
}

func lowercaseAndUnderscore(s string) string {
	res := strings.Builder{}
	for _, l := range s {
		if l < 91 && l > 64 {
			l += 32
			res.WriteRune(underscore)
		}
		res.WriteRune(l)
	}
	return strings.Trim(res.String(), "_ ")
}
