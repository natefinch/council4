package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCompiledNumericSourceDoesNotReadEventOrTarget(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Source Numeric", Layout: "normal", TypeLine: "Creature",
		Power: new("1"), Toughness: new("2"), OracleText: "Whenever another creature enters, tap target creature. This creature gets +1/+1 until end of turn. If this creature's toughness is 4 or greater, you gain 2 life."})
	for _, extra := range []int{0, 1, 2} {
		t.Run(strconv.Itoa(extra), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			source.Counters.Add(counter.PlusOnePlusOne, extra)
			event := addCreatureWithPowerToughness(g, game.Player2, 9, 9)
			target := addCreatureWithPowerToughness(g, game.Player3, 9, 9)
			obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, HasTriggerEvent: true, TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: event.ObjectID},
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 40
			if extra >= 1 {
				want += 2
			}
			if g.Players[game.Player1].Life != want {
				t.Fatalf("life=%d, want %d", g.Players[game.Player1].Life, want)
			}
		})
	}
}

func TestDepartedNumericSnapshotPreservesActualCharacteristics(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"copied mana value", "alternate face", "face down zero"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			permanent := addCreatureWithPowerToughness(g, game.Player2, 2, 3)
			def := g.CardInstances[permanent.CardInstanceID].Def
			def.ManaCost = opt.Val(cost.Mana{cost.O(8)})
			want := 3
			if name == "copied mana value" {
				copied := addArtifactWithManaValue(g, game.Player2, 3)
				resolveInstruction(NewEngine(nil), g, &game.StackObject{Controller: game.Player2,
					SourceID: permanent.ObjectID, SourceCardID: permanent.CardInstanceID,
					Targets: []game.Target{game.PermanentTarget(copied.ObjectID)}},
					game.BecomeCopy{Object: game.TargetPermanentReference(0)}, nil)
			}
			if name == "alternate face" {
				def.Alternate = opt.Val(def.CardFace)
				def.Alternate.Val.ManaCost = opt.Val(cost.Mana{cost.O(3)})
				permanent.Face = game.FaceAlternate
			}
			if name == "face down zero" {
				permanent.FaceDown = true
				want = 0
			}
			objectID := permanent.ObjectID
			if !movePermanentToZone(g, permanent, zone.Graveyard) {
				t.Fatal("departure failed")
			}
			for _, value := range []int{want, want + 1} {
				condition := opt.Val(game.Condition{Object: opt.Val(game.EventPermanentReference()), ObjectMatches: opt.Val(game.Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: value})})})
				obj := &game.StackObject{Controller: game.Player1, HasTriggerEvent: true, TriggerEvent: game.Event{Kind: game.EventPermanentDied, PermanentID: objectID}}
				if got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition); got != (value == want) {
					t.Fatalf("value=%d got=%t, actual departed value=%d", value, got, want)
				}
			}
		})
	}
}

func TestNumericConditionConjunctionAndPolarity(t *testing.T) {
	t.Parallel()
	for _, power := range []int{2, 3, 4} {
		for _, life := range []int{20, 40} {
			for _, negate := range []bool{false, true} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				permanent := addCreatureWithPowerToughness(g, game.Player2, power, 5)
				obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}
				g.Players[game.Player1].Life = life
				condition := opt.Val(game.Condition{Negate: negate, Object: opt.Val(game.TargetPermanentReference(0)), Types: []types.Card{types.Creature},
					Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 30}},
					ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3}),
						Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 5})})})
				want := (power >= 3 && life >= 30) != negate
				if got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition); got != want {
					t.Fatalf("power=%d life=%d negate=%t got=%t, want %t", power, life, negate, got, want)
				}
			}
		}
	}
}
