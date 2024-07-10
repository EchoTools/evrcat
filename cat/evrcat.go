package cat

import (
	"regexp"
	"strings"

	"github.com/heroiclabs/nakama/v3/server/evr"
)

var (
	symbolPattern = regexp.MustCompile(`0x[0-9a-fA-F]{16}`)
)

type EVRCat struct {
	cache map[string]string
}

func NewEVRCat() *EVRCat {
	return &EVRCat{
		cache: make(map[string]string),
	}
}

func (c *EVRCat) FromHashStringToToken(s string) string {
	t, ok := c.cache[s]
	if !ok {
		t = evr.ToSymbol(s).Token().String()
		c.cache[s] = t
	}
	return t
}

func (c *EVRCat) BuildHashMap(tokens []string) map[evr.Symbol]evr.SymbolToken {
	hashMap := make(map[evr.Symbol]evr.SymbolToken, len(tokens))
	for _, t := range tokens {
		if len(t) == 0 {
			continue
		}
		sym := evr.ToSymbol(t)
		hashMap[sym] = evr.SymbolToken(t)
	}
	return hashMap
}

// ReplaceTokens replaces all strings with hashes. (If hashmap is not nil, it will also populate the hashmap)
func (c *EVRCat) ReplaceTokens(line string, upper bool, hashmap map[evr.Symbol]evr.SymbolToken) string {
	tokens := strings.Split(line, " ")
	hashes := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if len(t) == 0 {
			continue
		}
		sym := evr.ToSymbol(t)
		t := sym.HexString()
		if upper {
			t = strings.ToUpper(t)
		}
		hashes = append(hashes, t)
		if hashmap != nil {
			hashmap[sym] = evr.SymbolToken(t)
		}
	}
	return strings.Join(hashes, " ")
}

// ReplaceHashes replaces (known) hashes with strings, otherwise leaves them as is
func (c *EVRCat) ReplaceHashes(line string) string {
	matches := symbolPattern.FindAllString(line, -1)
	replacements := make([]string, 0, len(matches)*2)
	for _, m := range matches {
		replacement := c.FromHashStringToToken(m)
		if !strings.HasPrefix(replacement, "0x") {
			replacements = append(replacements, m, replacement)
		}
	}
	replacer := strings.NewReplacer(replacements...)
	line = replacer.Replace(line)
	return line
}
