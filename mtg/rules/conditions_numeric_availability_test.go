package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCapturedNumericConditionUnavailableIsNotFalse(t *testing.T) {
	t.Parallel()
	for _, value := range []int{-1, 2, 3} {
		t.Run(strconv.Itoa(value), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.StackObjectTarget(999)}}
			if value >= 0 {
				spell := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(value)})
				obj.Targets[0] = game.StackObjectTarget(spell.ID)
			}
			sequence := []game.Instruction{
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)}, PublishCondition: "numeric",
					Condition: opt.Val(stackSelectionCondition(0, game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}))},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, ConditionGate: "numeric", ConditionGateNegate: true},
			}
			NewEngine(nil).resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 40
			if value == 2 {
				want++
			}
			if value == 3 {
				want += 2
			}
			if g.Players[game.Player1].Life != want {
				t.Fatalf("life=%d, want %d; unavailable must not become a false complement", g.Players[game.Player1].Life, want)
			}
		})
	}
}

func TestCompiledTargetCardOtherwiseRequiresAvailableActualMove(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Card Otherwise", Layout: "normal", TypeLine: "Instant",
		OracleText: "Exile target card from a graveyard. If it was a creature card, you gain 2 life. Otherwise, you gain 1 life."})
	for _, name := range []string{"creature", "noncreature", "redirected"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardType := types.Creature
			if name == "noncreature" {
				cardType = types.Artifact
			}
			cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Target", Types: []types.Card{cardType}}})
			engine := NewEngine(nil)
			if name == "redirected" {
				resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.CreateReplacement{Replacement: &game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard, MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Hand,
				}}, nil)
			}
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{currentCardTarget(t, g, cardID)}}
			engine.resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			want := 42
			if name == "noncreature" {
				want = 41
			}
			if name == "redirected" {
				want = 40
			}
			if g.Players[game.Player1].Life != want {
				t.Fatalf("life=%d, want %d", g.Players[game.Player1].Life, want)
			}
		})
	}
}
