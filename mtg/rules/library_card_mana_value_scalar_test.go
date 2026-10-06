package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestObservedManaValueScalarAvailabilityAndIncarnation(t *testing.T) {
	divination := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Divination", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleText: "Draw two cards.",
	})
	forest := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Forest", Layout: "normal", TypeLine: "Basic Land — Forest", OracleText: "{T}: Add {G}.",
	})
	for _, outcome := range []string{"known zero", "nonzero", "empty", "skipped", "declined", "missing definition", "reincarnated input"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := divination
			if outcome == "known zero" {
				def = forest
			}
			card := addCardToLibrary(g, game.Player3, def)
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				Targets: []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)}}
			instruction := game.Instruction{Primitive: game.Reveal{
				Player: game.TargetPlayerReference(1), Amount: game.Fixed(1), PublishLinked: "observed",
				PublishCharacteristics: game.LibraryCardCharacteristics{ManaValue: "mana"},
			}}
			if outcome == "empty" {
				g.Players[game.Player3].Library.Remove(card)
			}
			if outcome == "missing definition" {
				g.CardInstances[card].Def = nil
			}
			if outcome == "skipped" {
				instruction.ConditionGate = "missing"
			}
			if outcome == "declined" {
				instruction.Optional = true
			}
			if outcome == "reincarnated input" {
				newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).
					resolveInstruction(&game.Instruction{Primitive: game.LookAtLibraryTop{
						Player: game.TargetPlayerReference(1), PublishLinked: "old",
					}})
				if !moveCardBetweenZones(g, game.Player3, card, zone.Library, zone.Hand) ||
					!moveCardBetweenZones(g, game.Player3, card, zone.Hand, zone.Library) {
					t.Fatal("failed to create a different actual card incarnation")
				}
				instruction.Primitive = game.Reveal{
					Card: game.CardReference{Kind: game.CardReferenceLinked, LinkID: "old"}, PublishLinked: "observed",
					PublishCharacteristics: game.LibraryCardCharacteristics{ManaValue: "mana"},
				}
			}
			obj.ResolvedAmounts = map[string]int{"mana": 91, "independent": 92}
			obj.ResolutionResults = map[string]game.InstructionResolutionResult{"mana": {Succeeded: true, Amount: 91}}
			obj.ResolutionResultObjects = map[string][]game.ObjectSnapshot{"mana": {{CardID: card}}}
			obj.ResolvedExcessDamage = map[string]int{"mana": 93}
			resolver := newEffectResolver(NewEngine(nil), g, obj,
				[game.NumPlayers]PlayerAgent{game.Player1: &libraryPaymentAgent{}}, &TurnLog{})
			resolver.resolveInstruction(&instruction)
			wantAvailable := outcome == "known zero" || outcome == "nonzero"
			mana, available := obj.ResolvedAmounts["mana"]
			receipt, hasReceipt := obj.ResolutionResults["mana"]
			if available != wantAvailable || hasReceipt != wantAvailable ||
				available && (mana != def.ManaValue() || !receipt.Succeeded || receipt.Amount != mana) {
				t.Fatalf("scalar=%d available=%v receipt=%#v, want available=%v exact value=%d",
					mana, available, receipt, wantAvailable, def.ManaValue())
			}
			if _, stale := obj.ResolutionResultObjects["mana"]; stale {
				t.Fatal("scalar publication retained old object members")
			}
			if _, stale := obj.ResolvedExcessDamage["mana"]; stale {
				t.Fatal("scalar publication retained old excess damage")
			}
			if obj.ResolvedAmounts["independent"] != 92 {
				t.Fatal("scalar cleanup invalidated an unrelated amount")
			}
			resolver.resolveInstruction(&game.Instruction{
				Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				ResultGate: opt.Val(game.InstructionResultGate{Key: "mana", AmountAvailable: true}),
			})
			wantLife := 40
			if wantAvailable {
				wantLife++
				refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "observed"))
				if len(refs) != 1 || refs[0].CardID != card || refs[0].CardZoneVersion != 0 || !refs[0].CardZoneVersionSet {
					t.Fatal("scalar capture weakened actual identity or known-zero incarnation")
				}
				if !moveCardBetweenZones(g, game.Player3, card, zone.Library, zone.Hand) {
					t.Fatal("failed to move actual observed card")
				}
				g.CardInstances[card].Def = forest
				if _, live := resolveObjectReference(g, obj, game.LinkedObjectReference("observed")); live {
					t.Fatal("observation link ignored its invalidated incarnation")
				}
				if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
					Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "mana", Multiplier: 1,
				}); got != mana {
					t.Fatalf("frozen mana=%d, want %d despite later definition/move", got, mana)
				}
			}
			if g.Players[game.Player1].Life != wantLife || g.Players[game.Player3].Life != 40 {
				t.Fatal("known zero was unavailable or an unavailable scalar invented a controller/owner effect")
			}
		})
	}
}
