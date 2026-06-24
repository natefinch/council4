package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// matchSelectionFromSource matches sel against permanent while resolving the
// predicate's source object to sourceObjectID, mirroring the card-condition
// path that backs the Kinship "shares a creature type with this creature" gate.
func matchSelectionFromSource(
	g *game.Game,
	controller game.PlayerID,
	sourceObjectID id.ID,
	sel game.Selection,
	permanent *game.Permanent,
) bool {
	values := effectivePermanentValues(g, permanent)
	subject := selectionSubject{
		kind:           subjectPermanent,
		g:              g,
		permanent:      permanent,
		values:         &values,
		viewer:         controller,
		sourceObjectID: sourceObjectID,
		clampPower:     true,
	}
	return matchSelection(&subject, &sel)
}

// TestMatchSelectionSharesCreatureTypeWithSource exercises the Kinship
// shares-a-creature-type-with-this-creature gate (Wolf-Skull Shaman). A subject
// that shares any of the source's creature subtypes matches; one that shares
// none does not, and a source without creature subtypes matches nothing.
func TestMatchSelectionSharesCreatureTypeWithSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Wolf-Skull Shaman",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Shaman},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	sharer := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Elf Warrior",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Warrior},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	stranger := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Goblin Raider",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	typeless := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Stone Golem",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})

	sel := game.Selection{SharesCreatureTypeWithSource: true}

	if !matchSelectionFromSource(g, game.Player1, source.ObjectID, sel, sharer) {
		t.Error("a creature sharing the source's Elf type should match")
	}
	if matchSelectionFromSource(g, game.Player1, source.ObjectID, sel, stranger) {
		t.Error("a creature sharing none of the source's creature types must not match")
	}
	if matchSelectionFromSource(g, game.Player1, typeless.ObjectID, sel, sharer) {
		t.Error("a source without creature subtypes must match nothing")
	}
	if matchSelectionFromSource(g, game.Player1, 0, sel, sharer) {
		t.Error("an absent source object must not match")
	}
}
