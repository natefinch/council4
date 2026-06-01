package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CardLayout identifies a card layout that affects how printed card faces are
// represented.
type CardLayout string

const (
	LayoutNormal           CardLayout = ""
	LayoutTransform        CardLayout = "transform"
	LayoutModalDFC         CardLayout = "modal_dfc"
	LayoutMeld             CardLayout = "meld"
	LayoutDoubleFacedToken CardLayout = "double_faced_token"
	LayoutReversibleCard   CardLayout = "reversible_card"
)

// FaceIndex identifies one printed face of a card. The zero value is the front
// face so existing single-face cards and actions keep their historical meaning.
type FaceIndex int

const (
	FaceFront FaceIndex = iota
	FaceBack
)

// PT represents a creature's power or toughness. It can be a numeric
// value or a star (*) indicating a characteristic-defining ability (CR 208.2).
type PT struct {
	// Value is the numeric power or toughness. Ignored if IsStar is true.
	Value int

	// IsStar is true if this is a * value determined by a CDA.
	IsStar bool
}

// CardDef is the immutable definition of a Magic card — the "printed" card
// data from the card database. Multiple CardInstances in a game may reference
// the same CardDef.
type CardDef struct {
	// Name is the card's name (CR 201).
	Name string

	// Layout records the card layout when it changes face behavior.
	// Empty means a normal single-faced card.
	Layout CardLayout

	// ManaCost is the mana cost printed in the upper right (CR 202).
	// Absent for lands and some special cards.
	ManaCost opt.V[mana.Cost]

	// ManaValue is the card's mana value / converted mana cost (CR 202.3).
	ManaValue int

	// Colors are the colors of this card, determined by its mana cost
	// and color indicator (CR 105, 202.2).
	Colors []mana.Color

	// ColorIdentity is the card's color identity for Commander deck
	// construction (CR 903.4). Includes colors from mana cost, color
	// indicator, and mana symbols in rules text.
	ColorIdentity mana.ColorIdentity

	// Supertypes are the card's supertypes (types.Legendary, types.Basic, etc.).
	Supertypes []types.Super

	// Types are the card's primary types (Creature, Instant, etc.).
	Types []types.Card

	// Subtypes are the card's subtypes (Goblin, Equipment, Aura, etc.).
	Subtypes []types.Sub

	// Power is the creature's base power. Absent for non-creatures.
	Power opt.V[PT]

	// Toughness is the creature's base toughness. Absent for non-creatures.
	Toughness opt.V[PT]

	// DynamicPower and DynamicToughness describe characteristic-defining
	// abilities for star P/T values.
	DynamicPower     opt.V[DynamicValue]
	DynamicToughness opt.V[DynamicValue]

	// Loyalty is the planeswalker's starting loyalty. Absent for non-planeswalkers.
	Loyalty opt.V[int]

	// Defense is the battle's starting defense. Absent for non-battles.
	Defense opt.V[int]

	// EntersTapped, EntersTappedCondition, EntersTappedUnlessPaid, and
	// EntersWithCounters model common ETB replacement effects.
	// EntersTappedCondition means this permanent enters tapped when the condition
	// is true. EntersTappedUnlessPaid means the controller may pay the cost as it
	// enters; if they do not, it enters tapped.
	EntersTapped           bool
	EntersTappedCondition  opt.V[Condition]
	EntersTappedUnlessPaid opt.V[ResolutionPayment]
	EntersWithCounters     []CounterPlacement

	// Abilities lists all abilities on this card, parsed from the text box.
	Abilities []AbilityDef

	// ImplementationID names an optional rules-side hand-written card
	// implementation for behavior too complex to express declaratively.
	ImplementationID string

	// OracleText is the full oracle (rules) text of the card.
	OracleText string

	// Back holds the printed back-face characteristics for double-faced cards.
	// The CardDef root fields are the printed front-face characteristics.
	Back opt.V[CardFace]
}

// CardFace is one printed face of a card. It mirrors the printed
// characteristics from CardDef that can differ between faces.
type CardFace struct {
	Name                   string
	ManaCost               opt.V[mana.Cost]
	ManaValue              int
	Colors                 []mana.Color
	Supertypes             []types.Super
	Types                  []types.Card
	Subtypes               []types.Sub
	Power                  opt.V[PT]
	Toughness              opt.V[PT]
	DynamicPower           opt.V[DynamicValue]
	DynamicToughness       opt.V[DynamicValue]
	Loyalty                opt.V[int]
	Defense                opt.V[int]
	EntersTapped           bool
	EntersTappedCondition  opt.V[Condition]
	EntersTappedUnlessPaid opt.V[ResolutionPayment]
	EntersWithCounters     []CounterPlacement
	Abilities              []AbilityDef
	ImplementationID       string
	OracleText             string
}

// IsLegendary reports whether this card has the types.Legendary supertype.
func (c *CardDef) IsLegendary() bool {
	return c.HasSupertype(types.Legendary)
}

// HasSupertype reports whether this card has the given supertype.
func (c *CardDef) HasSupertype(supertype types.Super) bool {
	return c.DefaultFace().HasSupertype(supertype)
}

// HasType reports whether this card has the given card type.
func (c *CardDef) HasType(t types.Card) bool {
	return c.DefaultFace().HasType(t)
}

// HasSubtype reports whether this card has the given subtype.
func (c *CardDef) HasSubtype(sub types.Sub) bool {
	return c.DefaultFace().HasSubtype(sub)
}

// HasAnySubtype reports whether this card has any of the given subtypes.
func (c *CardDef) HasAnySubtype(subtypes ...types.Sub) bool {
	return c.DefaultFace().HasAnySubtype(subtypes...)
}

// HasKeyword reports whether any of this card's abilities grants the
// given keyword.
func (c *CardDef) HasKeyword(kw Keyword) bool {
	return c.DefaultFace().HasKeyword(kw)
}

// IsPermanent reports whether this card becomes a permanent when it resolves
// (i.e., it has at least one permanent card type).
func (c *CardDef) IsPermanent() bool {
	return c.DefaultFace().IsPermanent()
}

// DefaultFace returns the card characteristics used outside the stack and
// battlefield. For double-faced cards, that is the front face.
func (c *CardDef) DefaultFace() CardFace {
	return c.rootFace()
}

// Face returns the requested printed face. For single-faced cards, FaceFront
// maps to the root card characteristics.
func (c *CardDef) Face(index FaceIndex) (CardFace, bool) {
	switch index {
	case FaceFront:
		return c.rootFace(), true
	case FaceBack:
		return c.Back.Val, c.Back.Exists
	default:
		return CardFace{}, false
	}
}

// FaceDef returns a CardDef-shaped copy of one face's characteristics. It is a
// bridge for rules helpers that still operate on CardDef values.
func (c *CardDef) FaceDef(index FaceIndex) (*CardDef, bool) {
	face, ok := c.Face(index)
	if !ok {
		return nil, false
	}
	return face.ToCardDef(c), true
}

// FaceIndexes returns the printed faces available on this card.
func (c *CardDef) FaceIndexes() []FaceIndex {
	if c.Back.Exists {
		return []FaceIndex{FaceFront, FaceBack}
	}
	return []FaceIndex{FaceFront}
}

// CanChooseCastFace reports whether this face can be chosen while casting the
// card as a spell. Modal DFCs may choose any non-land face; other layouts cast
// only their front face.
func (c *CardDef) CanChooseCastFace(index FaceIndex) bool {
	face, ok := c.Face(index)
	if !ok || face.HasType(types.Land) {
		return false
	}
	if c.IsModalDoubleFaced() {
		return true
	}
	return index == FaceFront
}

// CanChooseLandFace reports whether this face can be played as a land.
func (c *CardDef) CanChooseLandFace(index FaceIndex) bool {
	face, ok := c.Face(index)
	if !ok || !face.HasType(types.Land) {
		return false
	}
	if !c.Back.Exists {
		return index == FaceFront
	}
	if c.IsModalDoubleFaced() {
		return true
	}
	return index == FaceFront
}

// LegalCastFaces returns all faces that may be chosen while casting this card.
func (c *CardDef) LegalCastFaces() []FaceIndex {
	var faces []FaceIndex
	for _, face := range c.FaceIndexes() {
		if c.CanChooseCastFace(face) {
			faces = append(faces, face)
		}
	}
	return faces
}

// IsModalDoubleFaced reports whether this card is a modal double-faced card.
func (c *CardDef) IsModalDoubleFaced() bool {
	return c.Layout == LayoutModalDFC
}

// IsTransformingDoubleFaced reports whether this card can use transform-style
// face switching. Meld and reversible cards are intentionally excluded.
func (c *CardDef) IsTransformingDoubleFaced() bool {
	return c.Layout == LayoutTransform || c.Layout == LayoutDoubleFacedToken
}

func (c *CardDef) rootFace() CardFace {
	return CardFace{
		Name:                   c.Name,
		ManaCost:               c.ManaCost,
		ManaValue:              c.ManaValue,
		Colors:                 append([]mana.Color(nil), c.Colors...),
		Supertypes:             append([]types.Super(nil), c.Supertypes...),
		Types:                  append([]types.Card(nil), c.Types...),
		Subtypes:               append([]types.Sub(nil), c.Subtypes...),
		Power:                  c.Power,
		Toughness:              c.Toughness,
		DynamicPower:           c.DynamicPower,
		DynamicToughness:       c.DynamicToughness,
		Loyalty:                c.Loyalty,
		Defense:                c.Defense,
		EntersTapped:           c.EntersTapped,
		EntersTappedCondition:  c.EntersTappedCondition,
		EntersTappedUnlessPaid: c.EntersTappedUnlessPaid,
		EntersWithCounters:     append([]CounterPlacement(nil), c.EntersWithCounters...),
		Abilities:              append([]AbilityDef(nil), c.Abilities...),
		ImplementationID:       c.ImplementationID,
		OracleText:             c.OracleText,
	}
}

// HasSupertype reports whether this face has the given supertype.
func (f CardFace) HasSupertype(supertype types.Super) bool {
	for _, st := range f.Supertypes {
		if st == supertype {
			return true
		}
	}
	return false
}

// HasType reports whether this face has the given card type.
func (f CardFace) HasType(t types.Card) bool {
	for _, ct := range f.Types {
		if ct == t {
			return true
		}
	}
	return false
}

// HasSubtype reports whether this face has the given subtype.
func (f CardFace) HasSubtype(sub types.Sub) bool {
	for _, s := range f.Subtypes {
		if s == sub {
			return true
		}
	}
	return false
}

// HasAnySubtype reports whether this face has any of the given subtypes.
func (f CardFace) HasAnySubtype(subtypes ...types.Sub) bool {
	for _, sub := range subtypes {
		if f.HasSubtype(sub) {
			return true
		}
	}
	return false
}

// HasKeyword reports whether any ability on this face grants the given keyword.
func (f CardFace) HasKeyword(kw Keyword) bool {
	for _, a := range f.Abilities {
		for _, k := range a.Keywords {
			if k == kw {
				return true
			}
		}
	}
	return false
}

// IsPermanent reports whether this face becomes a permanent when it resolves.
func (f CardFace) IsPermanent() bool {
	for _, t := range f.Types {
		if t.IsPermanent() {
			return true
		}
	}
	return false
}

// ToCardDef converts a face into a CardDef-shaped value for existing rules
// helpers. ColorIdentity stays on the physical card and is copied from parent.
func (f CardFace) ToCardDef(parent *CardDef) *CardDef {
	return &CardDef{
		Name:                   f.Name,
		ManaCost:               f.ManaCost,
		ManaValue:              f.ManaValue,
		Colors:                 append([]mana.Color(nil), f.Colors...),
		ColorIdentity:          parent.ColorIdentity,
		Supertypes:             append([]types.Super(nil), f.Supertypes...),
		Types:                  append([]types.Card(nil), f.Types...),
		Subtypes:               append([]types.Sub(nil), f.Subtypes...),
		Power:                  f.Power,
		Toughness:              f.Toughness,
		DynamicPower:           f.DynamicPower,
		DynamicToughness:       f.DynamicToughness,
		Loyalty:                f.Loyalty,
		Defense:                f.Defense,
		EntersTapped:           f.EntersTapped,
		EntersTappedCondition:  f.EntersTappedCondition,
		EntersTappedUnlessPaid: f.EntersTappedUnlessPaid,
		EntersWithCounters:     append([]CounterPlacement(nil), f.EntersWithCounters...),
		Abilities:              append([]AbilityDef(nil), f.Abilities...),
		ImplementationID:       f.ImplementationID,
		OracleText:             f.OracleText,
	}
}

// CardInstance represents a specific card in a game — one of the 100 cards
// in a player's deck, or a card created during play. Each CardInstance has a
// unique ID and references an immutable CardDef.
type CardInstance struct {
	// ID is the unique identifier for this card instance in the game.
	ID id.ID

	// Def is the static card definition this instance is based on.
	Def *CardDef

	// Owner is the player who owns this card (the player whose deck it
	// started in). Owner never changes during a game (CR 108.3).
	Owner PlayerID
}
