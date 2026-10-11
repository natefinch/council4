package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestInterveningCounterConditionPreservesQueuedSourceOccurrence(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Intervening Source Context", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "At the beginning of each end step, if there are three or more quest counters on this enchantment, you may put a Kraken, Leviathan, Octopus, or Serpent creature card from your hand onto the battlefield.",
	})
	for _, subtype := range []types.Sub{types.Kraken, types.Leviathan, types.Octopus, types.Serpent} {
		for _, scenario := range []string{"original LKI qualifies", "original LKI refuses", "live loses threshold", "capture refuses"} {
			t.Run(string(subtype)+"/"+scenario, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				source := addCombatPermanent(g, game.Player1, def)
				source.Controller = game.Player2
				source.Counters.Add(counter.Quest, 3)
				decoy := addCombatPermanent(g, game.Player3, def)
				decoy.Counters.Add(counter.Quest, 9)
				cardDef := vanillaCreature("Original Controller Hand Card", 3, 4)
				cardDef.Subtypes = []types.Sub{subtype}
				cardID := addCardToHand(g, game.Player2, cardDef)
				foreignCardID := addCardToHand(g, game.Player1, cardDef)
				if scenario == "capture refuses" {
					source.Counters.Remove(counter.Quest, 1)
				}
				g.Turn.ActivePlayer = game.Player4
				g.Turn.Step = game.StepEnd
				emitBeginningOfStepEvent(g, game.StepEnd)
				if !engine.putTriggeredAbilitiesOnStack(g) {
					t.Fatal("decoy's valid end-step trigger missing")
				}
				var original *game.StackObject
				for _, obj := range g.Stack.Objects() {
					if obj.SourceID == source.ObjectID {
						original = obj
					}
				}
				if scenario == "capture refuses" {
					if original != nil {
						t.Fatal("capture predicate borrowed decoy counters")
					}
					return
				}
				if original == nil || original.Controller != game.Player2 {
					t.Fatal("queued original occurrence/controller missing")
				}
				g.Stack.RemoveByID(original.ID)
				g.Stack.Push(original)
				if scenario == "live loses threshold" || scenario == "original LKI refuses" {
					source.Counters.Remove(counter.Quest, 1)
				}
				if scenario != "live loses threshold" {
					if !movePermanentToZone(g, source, zone.Graveyard) {
						t.Fatal("original source departure failed")
					}
					resolveInstruction(engine, g, &game.StackObject{
						Controller: game.Player1, Targets: []game.Target{currentCardTarget(t, g, source.CardInstanceID)},
					}, game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})}, nil)
					reentered := permanentForCard(g, source.CardInstanceID)
					if reentered == nil || reentered.ObjectID == source.ObjectID {
						t.Fatal("source reentry fixture failed")
					}
					reentered.Counters.Add(counter.Quest, 9)
				}
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
					game.Player2: &repeatLandChoiceAgent{
						mayAnswers: []bool{true}, landNames: []string{"Original Controller Hand Card"},
					},
				}, &TurnLog{})
				entered := permanentForCard(g, cardID)
				wantEntry := scenario == "original LKI qualifies"
				if wantEntry && (entered == nil || entered.Owner != game.Player2 || entered.Controller != game.Player2) ||
					!wantEntry && (!g.Players[game.Player2].Hand.Contains(cardID) || entered != nil) ||
					!g.Players[game.Player1].Hand.Contains(foreignCardID) {
					t.Fatal("resolution lost exact original-source/controller/hand context or accepted a decoy")
				}
			})
		}
	}
}
