package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestFixedPhaseMovePublicationClearsOnlyTransientLinks(t *testing.T) {
	t.Parallel()
	for _, reason := range []string{"declined", "condition", "failed"} {
		t.Run(reason, func(t *testing.T) {
			t.Parallel()
			sequence := compiledCaptureSequence(t, "Exile target creature. You gain 1 life. Return that card to its owner's hand at the beginning of the next end step.")
			if !sequence[0].ClearLinkedBeforeGate {
				t.Fatal("capture-owned move did not opt into transient clearing")
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			subject := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			obj := &game.StackObject{
				Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}, TargetCounts: []int{1},
			}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedCardID != subject.CardInstanceID {
				t.Fatal("successful producer did not freeze the actual exiled card")
			}
			persistent := linkedObjectSourceKey(g, obj, "persistent")
			rememberLinkedObject(g, persistent, permanentLinkedObjectRef(source))
			switch reason {
			case "declined":
				sequence[0].Optional = true
			case "condition":
				sequence[0].Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true, Negate: true})})
			}
			engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{}}, &TurnLog{})
			if len(g.DelayedTriggers) != 2 || g.DelayedTriggers[1].CapturedCardID != 0 {
				t.Fatal("skipped or failed move captured a stale card")
			}
			if refs := linkedObjects(g, persistent); len(refs) != 1 || refs[0].ObjectID != source.ObjectID {
				t.Fatal("transient clearing erased an unrelated persistent CR 607 link")
			}
			legacy := game.Instruction{
				Primitive: game.MovePermanent{
					Object: game.SourcePermanentReference(), Destination: zone.Exile, PublishLinked: "persistent",
				},
				Optional: true,
			}
			engine.resolveInstructionWithChoices(g, obj, &legacy, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{}}, &TurnLog{})
			if refs := linkedObjects(g, persistent); len(refs) != 1 || refs[0].ObjectID != source.ObjectID {
				t.Fatal("declining a persistent move publisher erased its existing link")
			}
			engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
			if !g.Players[game.Player2].Hand.Contains(subject.CardInstanceID) || g.Players[game.Player1].Hand.Contains(subject.CardInstanceID) {
				t.Fatal("delayed card return did not preserve the subject's owner")
			}
		})
	}
}

func TestFixedPhaseDivertedMoveDoesNotPublishCard(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "Exile target creature. You gain 1 life. Return that card to its owner's hand at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	subject := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{
		Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}, TargetCounts: []int{1},
	}
	resolveInstruction(engine, g, obj, game.CreateReplacement{
		Object: game.TargetPermanentReference(0),
		Replacement: &game.ReplacementEffect{
			MatchEvent: game.EventZoneChanged, MatchFromZone: true,
			FromZone: zone.Battlefield, ReplaceToZone: zone.Graveyard,
		},
	}, nil)
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(subject.CardInstanceID) {
		t.Fatal("replacement did not divert the actual move")
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].CapturedCardID != 0 {
		t.Fatal("a diverted exile published a card as if it had reached exile")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player2].Graveyard.Contains(subject.CardInstanceID) {
		t.Fatal("a missing capture redirected the delayed return")
	}
}

func TestFixedPhaseCapturesDoubledProductsAndCopiedController(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "Create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	copied := *obj
	copied.Controller = game.Player2
	engine.resolveInstructionSequence(g, &copied, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 2 || len(g.DelayedTriggers[0].CapturedObjectIDs) != 2 ||
		len(g.DelayedTriggers[1].CapturedObjectIDs) != 1 || g.DelayedTriggers[1].Controller != game.Player2 {
		t.Fatal("capture substituted printed quantity or original controller for actual products")
	}
	for _, captured := range g.DelayedTriggers[0].CapturedObjectIDs {
		if captured == g.DelayedTriggers[1].CapturedObjectIDs[0] {
			t.Fatal("copied resolution redirected an earlier frozen product")
		}
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			t.Fatal("an actual produced member escaped delayed sacrifice")
		}
	}
}
