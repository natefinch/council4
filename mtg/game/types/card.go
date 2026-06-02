package types

// Card represents a card's primary type (CR 300.1).
type Card string

// Card values identify supported primary card types.
const (
	Land         Card = "Land"
	Creature     Card = "Creature"
	Artifact     Card = "Artifact"
	Enchantment  Card = "Enchantment"
	Instant      Card = "Instant"
	Sorcery      Card = "Sorcery"
	Planeswalker Card = "Planeswalker"
	Battle       Card = "Battle"
	Kindred      Card = "Kindred"
	Plane        Card = "Plane"
	Dungeon      Card = "Dungeon"
	Phenomenon   Card = "Phenomenon"
	Scheme       Card = "Scheme"
	Vanguard     Card = "Vanguard"
	Conspiracy   Card = "Conspiracy"
)

// IsPermanent reports whether this card type represents a permanent.
func (c Card) IsPermanent() bool {
	switch c {
	case Land, Creature, Artifact, Enchantment, Planeswalker, Battle:
		return true
	default:
		return false
	}
}
