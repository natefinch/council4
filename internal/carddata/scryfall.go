// Package carddata provides the Oracle corpus metadata used by the card queue.
package carddata

// ScryfallCard holds the subset of Scryfall card JSON used by the repository.
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
