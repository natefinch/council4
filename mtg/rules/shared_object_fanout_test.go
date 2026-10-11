package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestSharedFanoutPayloadsAffectOnlyChosenObjects(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"power", "Up to two target creatures each get +2/+1 until end of turn."},
		{"keyword", "Up to two target creatures each gain flying until end of turn."},
		{"counter", "Put two +1/+1 counters on each of up to two target creatures."},
	} {
		for _, count := range []int{0, 1, 2} {
			t.Run(tc.name+"/"+strconv.Itoa(count), func(t *testing.T) {
				defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
					Name: "Fanout Payload", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.body,
				})
				if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
					t.Fatalf("compile: defs=%d diagnostics=%+v err=%v", len(defs), diagnostics, err)
				}
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				creature := &game.CardDef{CardFace: game.CardFace{
					Name: "Creature", Types: []types.Card{types.Creature},
					Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
				}}
				var targets []game.Target
				var objects []*game.Permanent
				for i := range 3 {
					object := addCombatPermanent(g, game.Player2, creature)
					objects = append(objects, object)
					if i < count {
						targets = append(targets, game.PermanentTarget(object.ObjectID))
					}
				}
				spell := addCardToHand(g, game.Player1, defs[0])
				g.Players[game.Player1].Hand.Remove(spell)
				g.Stack.Push(&game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1,
					Targets: targets, TargetCounts: []int{count},
				})
				NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
				for i, object := range objects {
					chosen := i < count
					wantPower, wantToughness, wantCounters := 3, 3, 0
					if chosen && tc.name == "power" {
						wantPower, wantToughness = 5, 4
					}
					if chosen && tc.name == "counter" {
						wantCounters, wantPower, wantToughness = 2, 5, 5
					}
					toughness, known := effectiveToughness(g, object)
					if effectivePower(g, object) != wantPower || !known || toughness != wantToughness ||
						object.Counters.Get(counter.PlusOnePlusOne) != wantCounters ||
						hasKeyword(g, object, game.Flying) != (chosen && tc.name == "keyword") {
						t.Fatalf("object %d did not preserve its selected payload (%s)", i, tc.name)
					}
				}
			})
		}
	}
}
