package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestTargetCardConditionRequiresActualReachedIncarnation(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"success", "unpublished", "original target left", "redirected", "later leave reentry",
		"reassigned target", "reassigned same card incarnation", "other resolving object",
		"skipped condition", "missing definition", "unknown target version",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Selected", Types: []types.Card{types.Artifact}}})
			decoy := addCardToGraveyard(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Decoy", Types: []types.Card{types.Artifact}}})
			target := currentCardTarget(t, g, cardID)
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{currentCardTarget(t, g, decoy), target}}
			instruction := game.Instruction{PublishResult: "move", Primitive: game.MoveCard{
				Card: game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}, FromZone: zone.Graveyard, Destination: zone.Exile}}
			condition := opt.Val(game.Condition{TargetCardResultKey: "move", Object: opt.Val(game.TargetCardReference(1)),
				ObjectMatches: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}})})
			if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition) {
				t.Fatal("forward publication matched")
			}
			engine := NewEngine(nil)
			if name == "original target left" {
				moveCardBetweenZones(g, game.Player2, cardID, zone.Graveyard, zone.Hand)
			}
			if name == "redirected" {
				resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{Replacement: &game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard, MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Hand,
				}}, nil)
			}
			if name == "unpublished" {
				instruction.PublishResult = ""
			}
			if name == "unknown target version" {
				obj.Targets[1].CardZoneVersionSet = false
			}
			engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if name == "later leave reentry" {
				moveCardBetweenZones(g, game.Player2, cardID, zone.Exile, zone.Hand)
				moveCardBetweenZones(g, game.Player2, cardID, zone.Hand, zone.Exile)
			}
			if name == "reassigned target" {
				obj.Targets[1] = currentCardTarget(t, g, decoy)
			}
			if name == "reassigned same card incarnation" {
				obj.Targets[1] = currentCardTarget(t, g, cardID)
			}
			if name == "other resolving object" {
				obj = &game.StackObject{Controller: game.Player3, Targets: obj.Targets}
			}
			if name == "missing definition" {
				g.CardInstances[cardID].Def = nil
			}
			if name == "skipped condition" {
				instruction.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{ControllerHandEmpty: true})})
				addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Hand"}})
				engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if _, available := obj.ResolutionResults["move"]; available {
					t.Fatal("skipped producer published an available failure")
				}
				if instructionResultGateSatisfied(g, obj, game.InstructionResultGate{Key: "move", Succeeded: game.TriTrue, Negate: true}) {
					t.Fatal("skipped producer satisfied Otherwise complement")
				}
				if instructionResultGateSatisfied(g, obj, game.InstructionResultGate{Key: "move", Succeeded: game.TriFalse}) {
					t.Fatal("skipped producer satisfied literal failure gate")
				}
			}
			want := name == "success"
			for _, negate := range []bool{false, true} {
				condition.Val.Negate = negate
				got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition)
				if got != (want && !negate) {
					t.Fatalf("negate=%t matched=%t, want %t", negate, got, want && !negate)
				}
			}
			if !g.Players[game.Player3].Graveyard.Contains(decoy) {
				t.Fatal("decoy slot moved")
			}
		})
	}
}

func TestCompiledTargetCardTypeConditionUsesPublishedMove(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Card Type Context", Layout: "normal", TypeLine: "Creature",
		Power: new("1"), Toughness: new("1"), OracleText: "{1}: Exile target card from a graveyard. You gain 1 life. If it was a permanent card, put a +1/+1 counter on this creature."})
	for _, cardType := range []types.Card{types.Artifact, types.Creature, types.Land, types.Instant} {
		t.Run(string(cardType), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			cardID := addCardToGraveyard(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Card", Types: []types.Card{cardType}}})
			obj := &game.StackObject{Kind: game.StackActivatedAbility, Controller: game.Player1, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, Targets: []game.Target{currentCardTarget(t, g, cardID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.ActivatedAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 1
			if cardType == types.Instant {
				want = 0
			}
			if got := source.Counters.Get(counter.PlusOnePlusOne); got != want {
				t.Fatalf("counters=%d, want %d", got, want)
			}
			if !g.Players[game.Player3].Exile.Contains(cardID) || g.Players[game.Player1].Life != 41 {
				t.Fatal("move/payoff controller isolation failed")
			}
		})
	}
}
