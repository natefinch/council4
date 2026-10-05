package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostNumericActivationReadsEffectiveSnapshot(t *testing.T) {
	for _, test := range []struct {
		name, predicate string
		power, delta    int
		known, matches  bool
		faceDown        bool
	}{
		{"modified toughness true", "had toughness 4 or greater", 2, 2, true, true, false},
		{"modified toughness near miss", "had toughness 4 or greater", 2, 1, true, false, false},
		{"power known zero", "'s power was 0 or less", 0, 0, true, true, false},
		{"negative power not clamped", "'s power was 0 or less", -1, 0, true, true, false},
		{"power positive", "'s power was 0 or less", 1, 0, true, false, false},
		{"unavailable power", "'s power was 0 or less", 0, 0, false, false, false},
		{"face-down mana value known zero", "had mana value 0 or less", 2, 0, true, true, true},
		{"face-up mana value", "had mana value 0 or less", 2, 0, true, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			subject := "the sacrificed creature"
			if test.predicate[0] != '\'' {
				subject += " "
			}
			source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
				"Sacrifice a creature: You gain 1 life. If "+subject+test.predicate+", draw a card.", "Artifact"))
			paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Paid", Types: []types.Card{types.Creature}, ManaCost: opt.Val(cost.Mana{cost.O(9)}),
				Power: opt.V[game.PT]{Val: game.PT{Value: test.power}, Exists: test.known}, Toughness: opt.Val(game.PT{Value: test.power}),
			}})
			paid.FaceDown = test.faceDown
			if test.delta != 0 {
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					Layer: game.LayerPowerToughnessModify, AffectedObjectID: paid.ObjectID,
					PowerDelta: test.delta, ToughnessDelta: test.delta,
				})
			}
			reward := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
			setSorcerySpeedTurn(g, game.Player1)
			life := g.Players[game.Player1].Life
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
				t.Fatal("resolving comparison affected payment legality")
			}
			delete(g.LastKnownInformation, paid.ObjectID)
			g.ContinuousEffects = nil
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Life != life+1 || g.Players[game.Player1].Hand.Contains(reward) != test.matches {
				t.Fatal("numeric predicate did not use available effective pre-move characteristics")
			}
		})
	}
}

func TestPaidCostNumericInsteadSelectsExactlyOneBranch(t *testing.T) {
	for _, toughness := range []int{3, 4} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Paid", Types: []types.Card{types.Creature}, Toughness: opt.Val(game.PT{Value: toughness}),
		}})
		for range 5 {
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
		}
		spellID := addCardToHand(g, game.Player1, compiledPaidSubjectCard(t,
			"As an additional cost to cast this spell, sacrifice a creature.\nDraw two cards. If the sacrificed creature had toughness 4 or greater, draw three cards instead.", "Sorcery"))
		setSorcerySpeedTurn(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
			t.Fatal("numeric replacement cast failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		want := 2
		if toughness == 4 {
			want = 3
		}
		if g.Players[game.Player1].Hand.Size() != want {
			t.Fatalf("draw count=%d, want %d, not the sum of both branches", g.Players[game.Player1].Hand.Size(), want)
		}
	}
}
