package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostNumericTokenReplacementReadsCapturedToughness(t *testing.T) {
	for _, test := range []struct {
		name      string
		toughness opt.V[game.PT]
		delta     int
		want      int
	}{
		{"below threshold", opt.Val(game.PT{Value: 3}), 0, 1},
		{"at threshold", opt.Val(game.PT{Value: 4}), 0, 2},
		{"effective not printed", opt.Val(game.PT{Value: 2}), 2, 2},
		{"known zero", opt.Val(game.PT{Value: 0}), 0, 1},
		{"unavailable denies both complements", opt.V[game.PT]{}, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
				"{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature had toughness 4 or greater, create two Food tokens instead.", "Artifact"))
			paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Paid", Types: []types.Card{types.Creature}, Toughness: test.toughness,
			}})
			if test.delta != 0 {
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					Layer: game.LayerPowerToughnessModify, AffectedObjectID: paid.ObjectID, ToughnessDelta: test.delta,
				})
			}
			setSorcerySpeedTurn(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
				t.Fatal("replacement condition incorrectly affected payment")
			}
			_, stillPresent := permanentByObjectID(g, paid.ObjectID)
			if !source.Tapped || stillPresent || !g.Players[game.Player1].Graveyard.Contains(paid.CardInstanceID) {
				t.Fatal("tap/sacrifice costs were not actually paid")
			}
			g.ContinuousEffects = nil
			delete(g.LastKnownInformation, paid.ObjectID)
			engine.resolveTopOfStack(g, &TurnLog{})
			foods := 0
			for _, permanent := range g.Battlefield {
				if permanent.Token && slices.Contains(effectivePermanentValues(g, permanent).subtypes, types.Food) {
					foods++
				}
			}
			if foods != test.want {
				t.Fatalf("Food count=%d, want %d; conditional replacement must not add both branches", foods, test.want)
			}
		})
	}
}

func TestPaidCostDiscardNumericUsesActualDefaultFace(t *testing.T) {
	for _, manaValue := range []int{0, 3} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
			"Discard a card: You gain 1 life. If the discarded card had mana value 0 or less, draw a card.", "Artifact"))
		paidID := addCardToHand(g, game.Player1, &game.CardDef{
			CardFace: game.CardFace{Name: "Front", Types: []types.Card{types.Creature},
				ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}), Colors: []color.Color{color.Blue}},
			Back: opt.Val(game.CardFace{Name: "Back", Types: []types.Card{types.Instant},
				ManaCost: opt.Val(cost.Mana{cost.O(9)}), Colors: []color.Color{color.Red}}),
		})
		reward := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
		setSorcerySpeedTurn(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("discard failed")
		}
		obj, _ := g.Stack.Peek()
		fact := obj.PaidCostSubjects[0]
		if !fact.Snapshot.ManaValue.Exists || fact.Snapshot.ManaValue.Val != manaValue ||
			fact.Snapshot.Face != game.FaceFront || !slices.Equal(fact.Snapshot.Colors, []color.Color{color.Blue}) {
			t.Fatal("discarded numeric subject used another face")
		}
		card, _ := g.GetCardInstance(paidID)
		card.Def = &game.CardDef{CardFace: game.CardFace{Name: "Later", ManaCost: opt.Val(cost.Mana{cost.O(8)})}}
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player1].Hand.Contains(reward) != (manaValue == 0) {
			t.Fatal("discarded numeric condition substituted a later definition or unavailable zero")
		}
	}
}
