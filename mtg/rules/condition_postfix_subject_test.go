package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledPostfixTargetTypeConditions(t *testing.T) {
	t.Parallel()
	for _, keyword := range []string{"deathtouch", "hexproof"} {
		for _, test := range []struct {
			name          string
			cardTypes     []types.Card
			add, remove   []types.Card
			plus, loyalty int
		}{
			{name: "creature", cardTypes: []types.Card{types.Creature}, plus: 1},
			{name: "planeswalker", cardTypes: []types.Card{types.Planeswalker}, loyalty: 1},
			{name: "both", cardTypes: []types.Card{types.Creature, types.Planeswalker}, plus: 1, loyalty: 1},
			{name: "neither", cardTypes: []types.Card{types.Artifact}},
			{name: "became creature", cardTypes: []types.Card{types.Artifact}, add: []types.Card{types.Creature}, plus: 1},
			{name: "ceased creature", cardTypes: []types.Card{types.Creature}, remove: []types.Card{types.Creature}},
		} {
			t.Run(keyword+"/"+test.name, func(t *testing.T) {
				t.Parallel()
				def := compileUnlessCard(t, cardgen.ScryfallCard{
					Name: "Postfix Subject Probe", Layout: "normal", TypeLine: "Creature — Human",
					ManaCost: "{1}{W}", Power: new("1"), Toughness: new("1"),
					OracleText: "When Postfix Subject Probe enters, target permanent you control gains " + keyword +
						" until end of turn. Put a +1/+1 counter on it if it's a creature. Put a loyalty counter on it if it's a planeswalker.",
				})
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player1, def)
				targetDef := &game.CardDef{CardFace: game.CardFace{Name: "Subject", Types: test.cardTypes}}
				target := addCombatPermanent(g, game.Player1, targetDef)
				decoy := addCombatPermanent(g, game.Player1, targetDef)
				if len(test.add)+len(test.remove) > 0 {
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerType,
						AddTypes: test.add, RemoveTypes: test.remove,
					})
				}
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
					SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
					Targets: []game.Target{game.PermanentTarget(target.ObjectID)},
				}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if got := target.Counters.Get(counter.PlusOnePlusOne); got != test.plus {
					t.Fatalf("target +1/+1 counters=%d, want %d", got, test.plus)
				}
				if got := target.Counters.Get(counter.Loyalty); got != test.loyalty {
					t.Fatalf("target loyalty counters=%d, want %d", got, test.loyalty)
				}
				for _, other := range []*game.Permanent{source, decoy} {
					if other.Counters.Get(counter.PlusOnePlusOne) != 0 || other.Counters.Get(counter.Loyalty) != 0 {
						t.Fatal("postfix condition modified source or same-named untargeted object")
					}
				}
			})
		}
	}
}
