package cardgen

import (
	"regexp"
)

// manaSymbolRe matches a single mana symbol like {W}, {2}, {W/U}, {2/W}, {W/P}, {X}, {C}, {S}.
var manaSymbolRe = regexp.MustCompile(`\{([^}]+)\}`)
