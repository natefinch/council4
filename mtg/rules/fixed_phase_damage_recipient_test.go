package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestFixedPhaseDamageSourceFilterCapturesDamagedRecipient(t *testing.T) {
	t.Parallel()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Sosuke, Son of Seshiro", Layout: "normal", ManaCost: "{2}{G}{G}",
		TypeLine: "Legendary Creature — Snake Warrior", Power: new("3"), Toughness: new("4"),
		OracleText: "Other Snake creatures you control get +1/+0.\n" +
			"Whenever a Warrior you control deals combat damage to a creature, destroy that creature at end of combat.",
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: %v; diagnostics: %v; definitions: %d", err, diagnostics, len(defs))
	}
	ability := defs[0].TriggeredAbilities[0]
	capture := ability.Content.Modes[0].Sequence[0].Primitive.(game.CreateDelayedTrigger).Trigger
	if ability.Trigger.Pattern.Subject != game.TriggerSubjectDamageSource ||
		!capture.CapturedObject.Exists || capture.CapturedObject.Val != game.EventPermanentReference() ||
		capture.Timing != game.DelayedAtEndOfCombat {
		t.Fatal("damage-source filtering changed the delayed consequence's recipient identity or timing")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, defs[0])
	warriorDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Warrior", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Warrior},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 8}),
	}}
	warrior := addCombatPermanent(g, game.Player1, warriorDef)
	opponentWarrior := addCombatPermanent(g, game.Player2, warriorDef)
	nonWarrior := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	recipient := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Damaged Recipient", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 8}),
	}})
	decoy := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	for _, dealer := range []*game.Permanent{opponentWarrior, nonWarrior} {
		dealPermanentDamage(g, dealer.CardInstanceID, dealer.ObjectID, dealer.Controller, recipient, 1, true)
		if engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("an uncontrolled Warrior or controlled non-Warrior passed the damage-source filter")
		}
	}
	dealPermanentDamage(g, warrior.CardInstanceID, warrior.ObjectID, game.Player1, recipient, 1, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("the controlled Warrior's actual combat damage did not trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedObjectID != recipient.ObjectID {
		t.Fatal("the damage-source trigger captured its Warrior instead of the damaged creature")
	}
	for _, departing := range []*game.Permanent{source, warrior} {
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1, SourceID: departing.ObjectID,
			SourceCardID: departing.CardInstanceID},
			game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile}, nil)
	}
	emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepEndOfCombat})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("the frozen consequence did not trigger after its sources departed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, live := permanentByObjectID(g, recipient.ObjectID); live {
		t.Fatal("the actual damaged recipient survived delayed destruction")
	}
	for _, unrelated := range []*game.Permanent{opponentWarrior, nonWarrior, decoy} {
		if _, live := permanentByObjectID(g, unrelated.ObjectID); !live {
			t.Fatal("delayed destruction removed an unrelated decoy")
		}
	}
}
