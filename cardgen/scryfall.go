// Package cardgen fetches card data from Scryfall and generates partial
// CardDef Go source files for the council4 card registry.
package cardgen

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ScryfallCard holds the subset of Scryfall card JSON we care about.
type ScryfallCard struct {
	Name          string             `json:"name"`
	Layout        string             `json:"layout"`
	ManaCost      string             `json:"mana_cost"`
	CMC           float64            `json:"cmc"`
	TypeLine      string             `json:"type_line"`
	OracleText    string             `json:"oracle_text"`
	Colors        []string           `json:"colors"`
	ColorIdentity []string           `json:"color_identity"`
	Keywords      []string           `json:"keywords"`
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
}

var scryfallClient = &http.Client{Timeout: 15 * time.Second}

// FetchCard fetches a card by exact name from the Scryfall API.
func FetchCard(name string) (*ScryfallCard, error) {
	u := "https://api.scryfall.com/cards/named?exact=" + url.QueryEscape(name)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating scryfall request: %w", err)
	}
	req.Header.Set("User-Agent", "council4/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := scryfallClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scryfall request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scryfall returned %d: %s", resp.StatusCode, string(body))
	}

	var card ScryfallCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("decoding scryfall response: %w", err)
	}

	if !supportedLayouts[card.Layout] {
		return nil, fmt.Errorf("unsupported card layout %q for %q", card.Layout, card.Name)
	}

	return &card, nil
}
