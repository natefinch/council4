package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
)

// cardDefTypeLine returns def's type line as its own characteristic-defining
// abilities define it (CR 604.3, which function in every zone; a Changeling
// card is every creature type). The slices alias def and must not be mutated.
func cardDefTypeLine(def *game.CardDef) spellTypeLine {
	face := characteristicDefinedFace(def)
	if face == nil {
		return spellTypeLine{}
	}
	return spellTypeLine{supertypes: face.Supertypes, types: face.Types, subtypes: face.Subtypes}
}

// ownedCardTypeLine returns the effective type line of card while it is in its
// owner's hand, library, graveyard, exile or command zone. base is the card's
// type line as the reader combines its printed faces, including its own
// characteristic-defining abilities. Battlefield static abilities whose
// type-changing effects declare an OwnedCardSelection then apply in layer order
// (CR 613.1d, 613.7), each to the cards its controller owns that match the
// selection as of the start of the type layer ("The same is true for ...
// creature cards you own that aren't on the battlefield"). A card on the stack
// is a spell (spellTypeLineFor), and a synthetic card without an ID has no
// zone. changed reports whether any effect applied; otherwise base is returned
// unchanged without being copied. Nothing is stored and nothing is revealed:
// the result tracks each source's current presence, controller and condition.
func ownedCardTypeLine(g *game.Game, card *game.CardInstance, base spellTypeLine) (line spellTypeLine, changed bool) {
	if g == nil || card == nil || card.ID == 0 {
		return base, false
	}
	effects := typeLineEffects(g, card.Owner, base, ownedCardTypeSelection)
	if len(effects) == 0 {
		return base, false
	}
	// cardZone finds only the library, hand, graveyard, exile and command zone.
	if _, ok := cardZone(g, card.ID); !ok {
		return base, false
	}
	owned := spellTypeLine{
		supertypes: slices.Clone(base.supertypes),
		types:      slices.Clone(base.types),
		subtypes:   slices.Clone(base.subtypes),
	}
	return applyTypeLineEffects(g, card.Owner, owned, effects), true
}

// offBattlefieldCardTypeLine returns the effective type line of card's default
// face (ownedCardTypeLine over cardDefTypeLine) for card-zone Selection reads.
func offBattlefieldCardTypeLine(g *game.Game, card *game.CardInstance) spellTypeLine {
	line, _ := ownedCardTypeLine(g, card, cardDefTypeLine(card.Def))
	return line
}
