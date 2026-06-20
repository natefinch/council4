package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// seedbornStaticBody is the Seedborn Muse static ability: untap every permanent
// the controller controls during each other player's untap step.
func seedbornStaticBody(permanentTypes ...types.Card) game.StaticAbility {
	return game.StaticAbility{
		RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,
			AffectedController: game.ControllerYou,
			PermanentTypes:     permanentTypes,
		}},
	}
}

// TestUntapDuringOtherPlayersUntapStepUntapsControllerPermanents verifies that a
// Seedborn Muse-style static ability untaps its controller's permanents (and
// itself) during another player's untap step without clearing summoning sickness.
func TestUntapDuringOtherPlayersUntapStepUntapsControllerPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	muse := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Seedborn Muse",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{seedbornStaticBody()},
	}})
	mine := makeCreaturePermanent(g, game.Player1, "My Creature")
	theirs := makeCreaturePermanent(g, game.Player2, "Their Creature")
	muse.Tapped = true
	mine.Tapped = true
	mine.SummoningSick = true
	theirs.Tapped = true
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if theirs.Tapped {
		t.Fatal("active player's permanent did not untap during its own untap step")
	}
	if mine.Tapped {
		t.Fatal("Seedborn Muse did not untap the controller's permanent during the other player's untap step")
	}
	if muse.Tapped {
		t.Fatal("Seedborn Muse did not untap itself during the other player's untap step")
	}
	if !mine.SummoningSick {
		t.Fatal("untapping during another player's untap step cleared summoning sickness")
	}
}

// TestUntapDuringOtherPlayersUntapStepRespectsTypeFilter verifies that a
// creature-filtered untap static (Drumbellower) untaps only the controller's
// creatures, leaving other permanent types tapped.
func TestUntapDuringOtherPlayersUntapStepRespectsTypeFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	drum := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Drumbellower",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 1}),
		Toughness:       opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{seedbornStaticBody(types.Creature)},
	}})
	creature := makeCreaturePermanent(g, game.Player1, "Other Creature")
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Tapped Land",
		Types: []types.Card{types.Land},
	}})
	drum.Tapped = true
	creature.Tapped = true
	land.Tapped = true
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if creature.Tapped {
		t.Fatal("creature-filtered untap left the controller's creature tapped")
	}
	if !land.Tapped {
		t.Fatal("creature-filtered untap incorrectly untapped a land")
	}
}

// TestUntapDuringOtherPlayersUntapStepSelfForm verifies the self form ("Untap
// this artifact during each other player's untap step.") untaps only the source
// permanent and not its controller's other permanents.
func TestUntapDuringOtherPlayersUntapStepSelfForm(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	clock := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Self Untapper",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectUntapDuringOtherPlayersUntapStep,
				AffectedSource: true,
			}},
		}},
	}})
	other := makeCreaturePermanent(g, game.Player1, "Untouched Creature")
	clock.Tapped = true
	other.Tapped = true
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if clock.Tapped {
		t.Fatal("self-form untap left the source permanent tapped")
	}
	if !other.Tapped {
		t.Fatal("self-form untap incorrectly untapped another permanent")
	}
}
