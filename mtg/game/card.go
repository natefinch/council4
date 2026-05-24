package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Supertype represents a card's supertype (CR 205.4).
type Supertype int

const (
	SupertypeNone Supertype = iota
	Legendary
	Basic
	Snow
	World
	Ongoing
)

// String returns the supertype name.
func (s Supertype) String() string {
	switch s {
	case Legendary:
		return "Legendary"
	case Basic:
		return "Basic"
	case Snow:
		return "Snow"
	case World:
		return "World"
	case Ongoing:
		return "Ongoing"
	default:
		return ""
	}
}

// CardType represents a card's primary type (CR 300.1).
type CardType int

const (
	TypeLand CardType = iota
	TypeCreature
	TypeArtifact
	TypeEnchantment
	TypeInstant
	TypeSorcery
	TypePlaneswalker
	TypeBattle
	TypeKindred
)

// String returns the card type name.
func (t CardType) String() string {
	switch t {
	case TypeLand:
		return "Land"
	case TypeCreature:
		return "Creature"
	case TypeArtifact:
		return "Artifact"
	case TypeEnchantment:
		return "Enchantment"
	case TypeInstant:
		return "Instant"
	case TypeSorcery:
		return "Sorcery"
	case TypePlaneswalker:
		return "Planeswalker"
	case TypeBattle:
		return "Battle"
	case TypeKindred:
		return "Kindred"
	default:
		return "Unknown"
	}
}

// IsPermanentType reports whether this card type represents a permanent
// (stays on the battlefield after resolving).
func (t CardType) IsPermanentType() bool {
	switch t {
	case TypeLand, TypeCreature, TypeArtifact, TypeEnchantment,
		TypePlaneswalker, TypeBattle:
		return true
	default:
		return false
	}
}

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

	// ManaCost is the mana cost printed in the upper right (CR 202).
	// Nil for lands and some special cards.
	ManaCost *mana.Cost

	// ManaValue is the card's mana value / converted mana cost (CR 202.3).
	ManaValue int

	// Colors are the colors of this card, determined by its mana cost
	// and color indicator (CR 105, 202.2).
	Colors []mana.Color

	// ColorIdentity is the card's color identity for Commander deck
	// construction (CR 903.4). Includes colors from mana cost, color
	// indicator, and mana symbols in rules text.
	ColorIdentity mana.ColorIdentity

	// Supertypes are the card's supertypes (Legendary, Basic, etc.).
	Supertypes []Supertype

	// Types are the card's primary types (Creature, Instant, etc.).
	Types []CardType

	// Subtypes are the card's subtypes (Goblin, Equipment, Aura, etc.).
	Subtypes []string

	// Power is the creature's base power. Nil for non-creatures.
	Power *PT

	// Toughness is the creature's base toughness. Nil for non-creatures.
	Toughness *PT

	// Loyalty is the planeswalker's starting loyalty. Nil for non-planeswalkers.
	Loyalty *int

	// Defense is the battle's starting defense. Nil for non-battles.
	Defense *int

	// Abilities lists all abilities on this card, parsed from the text box.
	Abilities []AbilityDef

	// ImplementationID names an optional rules-side hand-written card
	// implementation for behavior too complex to express declaratively.
	ImplementationID string

	// OracleText is the full oracle (rules) text of the card.
	OracleText string
}

// IsLegendary reports whether this card has the Legendary supertype.
func (c *CardDef) IsLegendary() bool {
	for _, st := range c.Supertypes {
		if st == Legendary {
			return true
		}
	}
	return false
}

// HasType reports whether this card has the given card type.
func (c *CardDef) HasType(t CardType) bool {
	for _, ct := range c.Types {
		if ct == t {
			return true
		}
	}
	return false
}

// HasSubtype reports whether this card has the given subtype.
func (c *CardDef) HasSubtype(sub string) bool {
	for _, s := range c.Subtypes {
		if s == sub {
			return true
		}
	}
	return false
}

// HasKeyword reports whether any of this card's abilities grants the
// given keyword.
func (c *CardDef) HasKeyword(kw Keyword) bool {
	for _, a := range c.Abilities {
		for _, k := range a.Keywords {
			if k == kw {
				return true
			}
		}
	}
	return false
}

// IsPermanent reports whether this card becomes a permanent when it resolves
// (i.e., it has at least one permanent card type).
func (c *CardDef) IsPermanent() bool {
	for _, t := range c.Types {
		if t.IsPermanentType() {
			return true
		}
	}
	return false
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
