// Package cardgen compiles Scryfall card data and Oracle text into executable
// CardDef Go source files for the council4 card registry.
package cardgen

// ScryfallCard holds the subset of Scryfall card JSON we care about.
type ScryfallCard struct {
	ID            string             `json:"id"`
	OracleID      string             `json:"oracle_id"`
	Name          string             `json:"name"`
	Layout        string             `json:"layout"`
	SetType       string             `json:"set_type"`
	Games         []string           `json:"games"`
	Digital       bool               `json:"digital"`
	Legalities    map[string]string  `json:"legalities"`
	ManaCost      string             `json:"mana_cost"`
	TypeLine      string             `json:"type_line"`
	OracleText    string             `json:"oracle_text"`
	Colors        []string           `json:"colors"`
	ColorIdentity []string           `json:"color_identity"`
	Power         *string            `json:"power"`
	Toughness     *string            `json:"toughness"`
	Loyalty       *string            `json:"loyalty"`
	Defense       *string            `json:"defense"`
	CardFaces     []ScryfallCardFace `json:"card_faces"`
}

// ScryfallCardFace holds per-face Scryfall fields for multi-face cards.
type ScryfallCardFace struct {
	Name       string   `json:"name"`
	ManaCost   string   `json:"mana_cost"`
	TypeLine   string   `json:"type_line"`
	OracleText string   `json:"oracle_text"`
	Colors     []string `json:"colors"`
	Power      *string  `json:"power"`
	Toughness  *string  `json:"toughness"`
	Loyalty    *string  `json:"loyalty"`
	Defense    *string  `json:"defense"`
}

// supportedLayouts are Scryfall card layouts that the generator can handle.
var supportedLayouts = map[string]bool{
	"normal":             true,
	"token":              true,
	"leveler":            true,
	"saga":               true,
	"class":              true,
	"case":               true,
	"prototype":          true,
	"host":               true,
	"augment":            true,
	"emblem":             true,
	"mutate":             true,
	"planar":             true,
	"scheme":             true,
	"vanguard":           true,
	"transform":          true,
	"modal_dfc":          true,
	"meld":               true,
	"double_faced_token": true,
	"reversible_card":    true,
	"adventure":          true,
	"split":              true,
	"prepare":            true,
}
