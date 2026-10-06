package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestOptionalCaptureOriginalProductAvailability(t *testing.T) {
	for _, body := range []struct{ name, text string }{
		{"ordinary", "you may gain 1 life and return target creature card from your graveyard to the battlefield."},
		{"ineffective first", "you may discard a card and return target creature card from your graveyard to the battlefield."},
		{"expanded first", "you may add {R}{G} and return target creature card from your graveyard to the battlefield."},
		{"independent intervener", "you may return target creature card from your graveyard to the battlefield and gain 1 life. You gain 2 life."},
	} {
		t.Run(body.name, func(t *testing.T) {
			sequence := compiledCaptureSequence(t, "Tap target artifact. If you have no cards in hand, "+body.text+
				" It gains haste until end of turn. Exile it at the beginning of the next end step.")
			for _, reason := range []string{"declined", "false", "failed", "null source", "reentered"} {
				t.Run(reason, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := NewEngine(nil)
					source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
					decoy := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}})
					cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
					obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
						Targets:      []game.Target{game.PermanentTarget(decoy.ObjectID), currentCardTarget(t, g, cardID)},
						TargetCounts: []int{1, 1}}
					agent := &scopedMayAgent{accept: []bool{true, false}}
					agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
					engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
					entered, exists := reanimatedPermanent(g, cardID)
					if !exists || len(g.DelayedTriggers) != 1 ||
						g.DelayedTriggers[0].CapturedObjectID != entered.ObjectID || !hasKeyword(g, entered, game.Haste) {
						t.Fatal("accepted group did not freeze its original actual entered permanent")
					}
					first := entered.ObjectID
					var reentered game.ObjectID
					switch reason {
					case "declined":
					case "false":
						addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
					case "failed":
						agent.accept[1] = true
					case "null source":
						obj.SourceID, obj.SourceCardID = 0, 0
					case "reentered":
						departure := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(first)}}
						resolveInstruction(engine, g, departure, game.MovePermanent{
							Object: game.TargetPermanentReference(0), Destination: zone.Graveyard,
						}, nil)
						fresh := &game.StackObject{Controller: game.Player1, Targets: []game.Target{currentCardTarget(t, g, cardID)}}
						resolveInstruction(engine, g, fresh, game.PutOnBattlefield{
							Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
						}, nil)
						current, live := reanimatedPermanent(g, cardID)
						if !live || current.ObjectID == first {
							t.Fatal("subject did not acquire a new incarnation")
						}
						reentered = current.ObjectID
						agent.accept[1] = true
					default:
						t.Fatalf("unknown reason %q", reason)
					}
					engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
					if len(g.DelayedTriggers) != 2 || g.DelayedTriggers[1].CapturedObjectID != 0 ||
						len(g.DelayedTriggers[1].CapturedObjectIDs) != 0 || g.DelayedTriggers[1].CapturedCardID != 0 ||
						g.DelayedTriggers[0].CapturedObjectID != first {
						t.Fatal("unavailable original product or mandatory haste shim recaptured a stale object")
					}
					wantChoices := 2
					if reason == "false" {
						wantChoices = 1
					}
					if agent.next != wantChoices {
						t.Fatalf("asked %d times, want %d single group decisions", agent.next, wantChoices)
					}
					engine.runEndingPhase(g, agents)
					if reentered != 0 {
						if _, live := permanentByObjectID(g, reentered); !live {
							t.Fatal("frozen capture followed the card into its new incarnation")
						}
					} else if _, live := permanentByObjectID(g, first); live {
						t.Fatal("first frozen product escaped its delayed exile")
					}
					if _, live := permanentByObjectID(g, decoy.ObjectID); !live {
						t.Fatal("capture redirected to a qualifying target decoy")
					}
					if _, live := permanentByObjectID(g, source.ObjectID); !live {
						t.Fatal("capture substituted the source for the actual product")
					}
				})
			}
		})
	}
}

func TestOptionalCaptureIndependentProducedGroups(t *testing.T) {
	sequence := compiledCaptureSequence(t,
		"You may discard a card and create a 1/1 green Insect creature token. You gain 1 life. Sacrifice it at the beginning of the next end step. "+
			"You may add {R}{G} and create a 1/1 green Insect creature token. You gain 2 life. Exile it at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	decoy := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	agent := &scopedMayAgent{accept: []bool{true, true, false, true}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	for range 2 {
		engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
	}
	if len(g.DelayedTriggers) != 4 || len(g.DelayedTriggers[2].CapturedObjectIDs) != 0 ||
		g.DelayedTriggers[2].CapturedObjectID != 0 || agent.next != 4 {
		t.Fatal("independent optional decisions or declined capture aliased")
	}
	seen := make(map[game.ObjectID]bool)
	for _, i := range []int{0, 1, 3} {
		if len(g.DelayedTriggers[i].CapturedObjectIDs) != 2 {
			t.Fatal("capture did not preserve the actual replacement-expanded batch")
		}
		for _, id := range g.DelayedTriggers[i].CapturedObjectIDs {
			if seen[id] || id == source.ObjectID || id == decoy.ObjectID {
				t.Fatal("independent products or resolutions captured an alias")
			}
			seen[id] = true
		}
	}
	engine.runEndingPhase(g, agents)
	for id := range seen {
		if _, live := permanentByObjectID(g, id); live {
			t.Fatal("a genuine captured product escaped its delayed action")
		}
	}
	if _, live := permanentByObjectID(g, decoy.ObjectID); !live {
		t.Fatal("delayed action selected an unrelated creature")
	}
}

func TestOptionalCaptureMoveDeclinePreservesFrozenAndPersistentProducts(t *testing.T) {
	sequence := compiledCaptureSequence(t,
		"You may gain 1 life and exile target creature. You gain 2 life. Return that card to its owner's hand at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	subject := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{game.PermanentTarget(subject.ObjectID)}, TargetCounts: []int{1}}
	persistent := linkedObjectSourceKey(g, obj, "persistent")
	rememberLinkedObject(g, persistent, permanentLinkedObjectRef(source))
	agent := &scopedMayAgent{accept: []bool{true, false}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	for range 2 {
		engine.resolveInstructionSequence(g, obj, sequence, agents, &TurnLog{})
	}
	if len(g.DelayedTriggers) != 2 || g.DelayedTriggers[0].CapturedCardID != subject.CardInstanceID ||
		g.DelayedTriggers[1].CapturedCardID != 0 || agent.next != 2 || g.Players[game.Player1].Life != 45 {
		t.Fatal("declined group aliased a frozen card, duplicated choices, or gated the independent rider")
	}
	if refs := linkedObjects(g, persistent); len(refs) != 1 || refs[0].ObjectID != source.ObjectID {
		t.Fatal("transient optional capture invalidation erased a persistent CR607 product")
	}
	engine.runEndingPhase(g, agents)
	if !g.Players[game.Player2].Hand.Contains(subject.CardInstanceID) ||
		g.Players[game.Player1].Hand.Contains(subject.CardInstanceID) {
		t.Fatal("actual original exiled card lost its owner's destination")
	}
}
