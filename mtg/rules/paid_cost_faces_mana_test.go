package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostTransformedManaValueDistinguishesCopies(t *testing.T) {
	for _, test := range []struct {
		name     string
		layout   game.CardLayout
		copied   bool
		token    bool
		copyCost int
		backCost int
		want     int
	}{
		{"transformed original", game.LayoutTransform, false, false, 0, 0, 2},
		{"copy of back face", game.LayoutTransform, true, false, 0, 0, 0},
		{"copy with actual mana cost", game.LayoutTransform, true, false, 4, 0, 4},
		{"modal back face", game.LayoutModalDFC, false, false, 0, 5, 5},
		{"double-faced token", game.LayoutDoubleFacedToken, false, true, 0, 0, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
				"Sacrifice a creature: If the sacrificed creature had mana value 1 or greater, draw a card.", "Artifact"))
			def := &game.CardDef{Layout: test.layout,
				CardFace: game.CardFace{Name: "Front", Types: []types.Card{types.Battle}, ManaCost: opt.Val(cost.Mana{cost.O(2)})},
				Back: opt.Val(game.CardFace{Name: "Back", Types: []types.Card{types.Creature},
					ManaCost: opt.Val(cost.Mana{cost.O(test.backCost)}), Toughness: opt.Val(game.PT{Value: 4})}),
			}
			paid := addCombatPermanent(g, game.Player1, def)
			paid.Face = game.FaceBack
			if test.token {
				paid.Token = true
				paid.TokenDef = def
				paid.CardInstanceID = 0
			}
			if test.copied {
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					Layer: game.LayerCopy, AffectedObjectID: paid.ObjectID,
					CopyValues: opt.Val(game.CopyableValues{Name: "Copied Back", Types: []types.Card{types.Creature},
						ManaCost: opt.Val(cost.Mana{cost.O(test.copyCost)})}),
				})
			}
			reward := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
			setSorcerySpeedTurn(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
				t.Fatal("sacrifice failed")
			}
			obj, _ := g.Stack.Peek()
			value := obj.PaidCostSubjects[0].Snapshot.ManaValue
			if !value.Exists || value.Val != test.want {
				t.Fatalf("captured MV=%v, want known %d", value, test.want)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(reward) != (test.want >= 1) {
				t.Fatal("past-cost mana value used the wrong physical/copy face")
			}
		})
	}
}

func TestPaidCostDirectManaActivationRetainsActualFacts(t *testing.T) {
	for _, paidColor := range []color.Color{color.Red, color.Blue} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		def := &game.CardDef{CardFace: game.CardFace{Name: "Mana Source", Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{{
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, SubjectKey: "cost"}},
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Player: opt.Val(game.ControllerReference()), ManaColor: mana.R, Amount: game.Fixed(1)}},
					{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
						Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
							Object:        opt.Val(game.PaidCostReference("cost", game.PaidCostDiscard, "")),
							ObjectMatches: opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
						})})},
				}}.Ability(),
			}},
		}}
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("invalid mana fixture: %v", issues)
		}
		source := addCombatPermanent(g, game.Player1, def)
		paidID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Paid", Types: []types.Card{types.Creature}, Colors: []color.Color{paidColor},
		}})
		setSorcerySpeedTurn(g, game.Player1)
		life := g.Players[game.Player1].Life
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("mana activation failed")
		}
		want := life
		if paidColor == color.Red {
			want++
		}
		if _, stacked := g.Stack.Peek(); stacked || !g.Players[game.Player1].Graveyard.Contains(paidID) ||
			g.Players[game.Player1].ManaPool.Amount(mana.R) != 1 || g.Players[game.Player1].Life != want {
			t.Fatal("ephemeral mana resolution dropped the actual paid subject or used the stack")
		}
	}
}
