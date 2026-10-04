package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestResolvingSourceExclusionCompiledSequences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		card    cardgen.ScryfallCard
		subtype types.Sub
		check   func(*testing.T, *game.Game, *game.Permanent, int, bool)
	}{
		{
			card: cardgen.ScryfallCard{
				Name: "Hero in Training", Layout: "normal", ManaCost: "{2}{W}",
				TypeLine:   "Creature \u2014 Human Hero",
				OracleText: "When this creature enters, draw a card. If you control another Hero, you gain 2 life.",
				Power:      new("2"), Toughness: new("2"),
			},
			subtype: types.Hero,
			check: func(t *testing.T, g *game.Game, _ *game.Permanent, before int, other bool) {
				t.Helper()
				if cardInstanceCount(g, g.Players[game.Player1].Hand.All()) != 1 {
					t.Error("unconditional draw did not happen")
				}
				want := before
				if other {
					want += 2
				}
				if got := g.Players[game.Player1].Life; got != want {
					t.Errorf("gated life = %d, want %d", got, want)
				}
			},
		},
		{
			card: cardgen.ScryfallCard{
				Name: "Protocol Knight", Layout: "normal", ManaCost: "{3}{U}",
				TypeLine:   "Creature \u2014 Human Knight",
				OracleText: "When this creature enters, tap target creature an opponent controls. Put a stun counter on that creature if you control another Knight. (If a permanent with a stun counter would become untapped, remove one from it instead.)",
				Power:      new("3"), Toughness: new("4"),
			},
			subtype: types.Knight,
			check: func(t *testing.T, _ *game.Game, target *game.Permanent, _ int, other bool) {
				t.Helper()
				if !target.Tapped {
					t.Error("unconditional tap did not happen")
				}
				want := 0
				if other {
					want = 1
				}
				if got := target.Counters.Get(counter.Stun); got != want {
					t.Errorf("gated stun counters = %d, want %d", got, want)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.card.Name, func(t *testing.T) {
			t.Parallel()
			defs, diagnostics, err := cardgen.CompileCardDefs(&tt.card)
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile = %v, diagnostics = %v, defs = %d", err, diagnostics, len(defs))
			}
			def := defs[0]
			if len(def.TriggeredAbilities) != 1 || len(def.TriggeredAbilities[0].Content.Modes) != 1 {
				t.Fatal("compiled card did not retain its single triggered mode")
			}
			sequence := def.TriggeredAbilities[0].Content.Modes[0].Sequence
			if len(sequence) != 2 || sequence[0].Condition.Exists ||
				!sequence[1].Condition.Exists || !sequence[1].Condition.Val.Condition.Exists ||
				!sequence[1].Condition.Val.Condition.Val.ControlsMatching.Exists ||
				!sequence[1].Condition.Val.Condition.Val.ControlsMatching.Val.Selection.ExcludeSource {
				t.Fatal("compiled sequence did not retain the unconditional instruction and source-excluding payoff")
			}
			for _, controller := range []opt.V[game.PlayerID]{{}, opt.Val(game.Player1), opt.Val(game.Player2)} {
				name := "source only"
				if controller.Exists {
					name = "another controlled creature"
					if controller.Val == game.Player2 {
						name = "opponent creature"
					}
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					source := addCombatPermanent(g, game.Player1, def)
					target := addCreaturePermanent(g, game.Player2)
					if controller.Exists {
						other := resolvingSourceCreature("Other", 4)
						other.Subtypes = []types.Sub{tt.subtype}
						addCombatPermanent(g, controller.Val, other)
					}
					addCardToLibrary(g, game.Player1, resolvingSourceCreature("Draw", 1))
					obj := &game.StackObject{
						ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
						SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
						Controller: game.Player1,
					}
					if len(def.TriggeredAbilities[0].Content.Modes[0].Targets) != 0 {
						obj.Targets = []game.Target{game.PermanentTarget(target.ObjectID)}
					}
					g.Stack.Push(obj)
					before := g.Players[game.Player1].Life
					NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					tt.check(t, g, target, before, controller.Exists && controller.Val == game.Player1)
				})
			}
		})
	}
}
