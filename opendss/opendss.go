package opendss

import "strings"

// Utilities for dealing with OpenDSS files

// Specs is just an arbitrary jsonserializable JSON object.
// We need it to project arbitrary OpenDSS properties into GeoJSON properties.
type Specs map[string]interface{}

func (s Specs) String(key string) string {
	v, _ := s[key].(string)
	return v
}

// Parses a raw OpenDSS line of a given type (e.g. "Line", "Vsource"), returning
// the element's ID and the remaining spec tokens. Returns ok=false if the line
// is not a New/Edit statement for that type.
func parseOpenDSSElement(rawLine, typeName string) (id string, specTokens []string, ok bool) {
	tokens := strings.Fields(rawLine)
	if len(tokens) < 2 {
		return "", nil, false
	}
	if tokens[0] != "New" && tokens[0] != "Edit" {
		return "", nil, false
	}
	if !strings.HasPrefix(tokens[1], "\""+typeName+".") {
		return "", nil, false
	}
	parts := strings.SplitN(strings.Trim(tokens[1], "\""), ".", 2)
	if len(parts) != 2 {
		return "", nil, false
	}
	return parts[1], tokens[2:], true
}

// Pulls out the raw key/value pairs from OpenDSS that may be in various
// formats delimited with double-quotes and square-brackets. This isn't the
// entirity of the OpenDSS spec, but it'll do for now.
func parseSpecs(tokens []string) Specs {
	specs := make(Specs)
	var accumKey string
	var accumParts []string
	var accumEnd byte

	for _, tok := range tokens {
		if accumKey != "" {
			accumParts = append(accumParts, tok)
			if len(tok) > 0 && tok[len(tok)-1] == accumEnd {
				specs[accumKey] = strings.Join(accumParts, " ")
				accumKey = ""
				accumParts = nil
				accumEnd = 0
			}
			continue
		}
		eqIdx := strings.Index(tok, "=")
		if eqIdx >= 0 {
			k := tok[:eqIdx]
			v := tok[eqIdx+1:]
			if strings.HasPrefix(v, "\"") && !strings.HasSuffix(v, "\"") {
				accumKey = k
				accumParts = []string{v}
				accumEnd = '"'
			} else if strings.HasPrefix(v, "[") && !strings.HasSuffix(v, "]") {
				accumKey = k
				accumParts = []string{v}
				accumEnd = ']'
			} else {
				specs[k] = v
			}
		} else {
			specs[tok] = "true"
		}
	}
	return specs
}
