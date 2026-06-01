// Package types defines Magic card supertypes, card types, and subtypes.
package types

// Super represents a card's supertype (CR 205.4).
type Super string

const (
	Legendary Super = "Legendary"
	Basic     Super = "Basic"
	Snow      Super = "Snow"
	World     Super = "World"
	Ongoing   Super = "Ongoing"
)

// Card represents a card's primary type (CR 300.1).
type Card string

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

// Sub represents a card's subtype.
type Sub string

const (
	Clue      Sub = "Clue"
	Equipment Sub = "Equipment"
)

const (
	Angel       Sub = "Angel"
	Bear        Sub = "Bear"
	Beast       Sub = "Beast"
	Bird        Sub = "Bird"
	Cleric      Sub = "Cleric"
	Construct   Sub = "Construct"
	Druid       Sub = "Druid"
	Golem       Sub = "Golem"
	Human       Sub = "Human"
	Incarnation Sub = "Incarnation"
	Mutant      Sub = "Mutant"
	Ninja       Sub = "Ninja"
	Robot       Sub = "Robot"
	Shaman      Sub = "Shaman"
	Snake       Sub = "Snake"
	Turtle      Sub = "Turtle"
	Zombie      Sub = "Zombie"
)

const (
	Aura  Sub = "Aura"
	Class Sub = "Class"
)

const (
	Forest   Sub = "Forest"
	Island   Sub = "Island"
	Mountain Sub = "Mountain"
	Plains   Sub = "Plains"
	Swamp    Sub = "Swamp"
)

var subtypesByType = map[Card]map[Sub]struct{}{
	Artifact: subtypeSet(
		Clue,
		Equipment,
	),
	Creature: subtypeSet(
		Angel,
		Bear,
		Beast,
		Bird,
		Cleric,
		Construct,
		Druid,
		Golem,
		Human,
		Incarnation,
		Mutant,
		Ninja,
		Robot,
		Shaman,
		Snake,
		Turtle,
		Zombie,
	),
	Enchantment: subtypeSet(
		Aura,
		Class,
	),
	Land: subtypeSet(
		Forest,
		Island,
		Mountain,
		Plains,
		Swamp,
	),
}

func subtypeSet(subtypes ...Sub) map[Sub]struct{} {
	set := make(map[Sub]struct{}, len(subtypes))
	for _, subtype := range subtypes {
		set[subtype] = struct{}{}
	}
	return set
}

// KnownSubtypeForType reports whether subtype is defined for cardType.
func KnownSubtypeForType(cardType Card, subtype Sub) bool {
	if cardType == Kindred {
		cardType = Creature
	}
	subtypes := subtypesByType[cardType]
	_, ok := subtypes[subtype]
	return ok
}
