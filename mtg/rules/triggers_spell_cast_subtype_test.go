package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// elfCreature returns a green Elf creature spell used to exercise the
// single-subtype "Whenever you cast an Elf spell" trigger condition.
func elfCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Llanowar Elves",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}}
}

func castSpellCastSubtypeFixture(t *testing.T, spell *game.CardDef) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:         game.EventSpellCast,
		Controller:    game.TriggerControllerYou,
		CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Elf}},
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spellID := addCardToHand(g, game.Player1, spell)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatalf("cast %q failed", spell.Name)
	}
	return g, engine
}

func TestSpellCastSubtypeTriggerFiresOnMatchingSubtype(t *testing.T) {
	g, engine := castSpellCastSubtypeFixture(t, elfCreature())
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Elf spell did not fire the single-subtype cast trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want the subtype cast trigger to draw one card", got)
	}
}

func TestSpellCastSubtypeTriggerFailsClosedOnOtherSubtype(t *testing.T) {
	g, engine := castSpellCastSubtypeFixture(t, greenCreature())
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("non-Elf spell fired the single-subtype cast trigger")
	}
}
