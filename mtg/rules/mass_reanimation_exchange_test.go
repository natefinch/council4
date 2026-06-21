package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func battlefieldContainsCard(g *game.Game, cardID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return true
		}
	}
	return false
}

// TestMassReanimationExchangeSwapsGraveyardAndBattlefield verifies the symmetric
// exile-sacrifice-return exchange: every player's matching graveyard creatures
// enter the battlefield while their matching battlefield creatures are
// sacrificed, and the freshly sacrificed creatures are not caught by the return
// (the "cards they exiled this way" back-reference). Non-creature graveyard
// cards are left untouched.
func TestMassReanimationExchangeSwapsGraveyardAndBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	graveCreature1 := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grave Bear 1",
		Types: []types.Card{types.Creature},
	}})
	graveCreature2 := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grave Bear 2",
		Types: []types.Card{types.Creature},
	}})
	graveInstant := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bolt",
		Types: []types.Card{types.Instant},
	}})

	fieldCreature1 := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Field Bear 1",
		Types: []types.Card{types.Creature},
	}})
	fieldCreature2 := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Field Bear 2",
		Types: []types.Card{types.Creature},
	}})

	instr := &game.Instruction{Primitive: game.MassReanimationExchange{
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}}
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveInstructionWithChoices(g, obj, instr, agents, &TurnLog{})

	if !battlefieldContainsCard(g, graveCreature1) {
		t.Error("player 1 graveyard creature did not enter the battlefield")
	}
	if !battlefieldContainsCard(g, graveCreature2) {
		t.Error("player 2 graveyard creature did not enter the battlefield")
	}
	if battlefieldContainsCard(g, fieldCreature1.CardInstanceID) {
		t.Error("player 1 battlefield creature was not sacrificed")
	}
	if battlefieldContainsCard(g, fieldCreature2.CardInstanceID) {
		t.Error("player 2 battlefield creature was not sacrificed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(fieldCreature1.CardInstanceID) {
		t.Error("sacrificed player 1 creature is not in its graveyard")
	}
	if g.Players[game.Player1].Graveyard.Contains(graveCreature1) {
		t.Error("reanimated player 1 creature still in graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveInstant) {
		t.Error("non-creature graveyard card was disturbed")
	}
	if g.Players[game.Player1].Exile.Contains(graveCreature1) {
		t.Error("reanimated creature stranded in exile")
	}
}
