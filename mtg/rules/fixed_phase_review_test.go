package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestFixedPhaseDirectTargetRetainsOwnSlot(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "Tap target creature. Destroy target creature at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	decoy := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	subject := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{
		{Kind: game.TargetPermanent, PermanentID: decoy.ObjectID},
		{Kind: game.TargetPermanent, PermanentID: subject.ObjectID},
	}, TargetCounts: []int{1, 1}}
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedObjectID != subject.ObjectID {
		t.Fatal("direct delayed target redirected to the qualifying earlier target")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if _, live := permanentByObjectID(g, subject.ObjectID); live {
		t.Fatal("the actual second target survived its delayed destruction")
	}
	if _, live := permanentByObjectID(g, decoy.ObjectID); !live {
		t.Fatal("delayed destruction removed the first target decoy")
	}
}

func TestFixedPhaseDepartureCardUsesEventIncarnation(t *testing.T) {
	t.Parallel()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Departure Capture", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "Whenever another creature you control leaves the battlefield, return that card to the battlefield at the beginning of the next end step.",
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: %v; diagnostics: %v", err, diagnostics)
	}
	sequence := defs[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	for _, reentered := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		subject := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		obj := &game.StackObject{Controller: game.Player1, SourceID: subject.ObjectID, SourceCardID: subject.CardInstanceID}
		r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
		handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile})
		card, _ := g.GetCardInstance(subject.CardInstanceID)
		obj.HasTriggerEvent = true
		obj.TriggerEvent = game.Event{Kind: game.EventZoneChanged, PermanentID: subject.ObjectID,
			CardID: card.ID, CardZoneVersion: card.ZoneVersion, FromZone: zone.Battlefield, ToZone: zone.Exile}
		engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedCardID != card.ID {
			t.Fatal("generic departure did not capture the actual reached card")
		}
		if reentered {
			handlePutOnBattlefield(r, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource})})
			current, _ := findPermanentByCardID(g, card.ID)
			obj.SourceID = current.ObjectID
			handleMovePermanent(r, game.MovePermanent{Object: game.SourcePermanentReference(), Destination: zone.Exile})
		}
		engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
		_, returned := findPermanentByCardID(g, card.ID)
		if returned == reentered {
			t.Fatal("generic departure return did not preserve the captured card incarnation")
		}
	}
}

func TestFixedPhaseOriginalReturnDoesNotCaptureHasteShim(t *testing.T) {
	t.Parallel()
	for _, reason := range []string{"declined", "condition", "failed"} {
		t.Run(reason, func(t *testing.T) {
			t.Parallel()
			sequence := compiledCaptureSequence(t, "Return target creature card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step.")
			put := sequence[0].Primitive.(game.PutOnBattlefield)
			key := put.PublishLinked
			if key == "" {
				t.Fatal("the exact original returned product was not published")
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			subject := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
			resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1,
				Targets: []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}},
				game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Graveyard}, nil)
			card, _ := g.GetCardInstance(subject.CardInstanceID)
			obj.Targets = []game.Target{{Kind: game.TargetCard, CardID: card.ID, CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true}}
			obj.TargetCounts = []int{1}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			first := g.DelayedTriggers[0].CapturedObjectID
			if first == 0 {
				t.Fatal("accepted return did not freeze its actual permanent")
			}
			switch reason {
			case "declined":
				sequence[0].Optional = true
			case "condition":
				sequence[0].Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true, Negate: true})})
			}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{}}, &TurnLog{})
			if len(g.DelayedTriggers) != 2 || g.DelayedTriggers[1].CapturedObjectID != 0 ||
				len(g.DelayedTriggers[1].CapturedObjectIDs) != 0 {
				t.Fatal("mandatory haste shim republished a previous accepted return after the original producer became unavailable")
			}
			if g.DelayedTriggers[0].CapturedObjectID != first {
				t.Fatal("a skipped later producer redirected the earlier frozen subject")
			}
		})
	}
}

func TestFixedPhaseReturnedCardRetainsEntryTypeEffect(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "Exile target creature. Return that card to the battlefield at the beginning of the next end step. It's an enchantment. (It's not a creature.)")
	put := sequence[1].Primitive.(game.CreateDelayedTrigger).Trigger.Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if len(put.ContinuousEffects) != 1 || put.ContinuousEffects[0].SetTypes[0] != types.Enchantment {
		t.Fatal("captured return discarded its modeled entry type effect")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	subject := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}, TargetCounts: []int{1}}
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	returned, ok := findPermanentByCardID(g, subject.CardInstanceID)
	if !ok || !permanentHasType(g, returned, types.Enchantment) || permanentHasType(g, returned, types.Creature) {
		t.Fatalf("captured return did not execute its existing entry type-effect primitive: returned=%#v; effects=%#v", returned, g.ContinuousEffects)
	}
}
