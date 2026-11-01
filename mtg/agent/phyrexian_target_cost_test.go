package agent

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules"
)

// Source review: Comprehensive Rules effective 2026-09-25. CR 601.2b-c: the
// payment of each Phyrexian symbol is announced before the spell's targets are
// chosen; CR 601.2f: the total cost includes cost increases; CR 733.1: a cast
// that can't be legally completed is reversed. Oracle (pinned corpus):
// Dismember fd74f8eb-0253-42dd-8277-186d4934da38 {1}{B/P}{B/P} "Target
// creature gets -5/-5 until end of turn."; Elderwood Scion
// 99c8647f-e08d-4217-b637-3f8f2bffdaa3 "Spells your opponents cast that target
// this creature cost {2} more to cast."; Mutagenic Growth
// 2e0861a2-1858-47af-8154-20a977c2b298 {G/P} "Target creature gets +2/+2 until
// end of turn."; Grizzly Bears 2/2 vanilla. Engine gap #3573.

// TestGenericStrategyCastsPhyrexianSpellAtPayableTarget runs Player1's
// priority with GenericStrategy's choices, the registered Dismember in hand,
// one colorless mana, and Player2's Grizzly Bears and Elderwood Scion on the
// battlefield. Only the Bears can be targeted for a cost Player1 can pay, so
// the target prompt after the Phyrexian announcements offers only the Bears
// and Dismember is cast and resolves, rather than the strategy preferring the
// Scion and the cast being reversed.
func TestGenericStrategyCastsPhyrexianSpellAtPayableTarget(t *testing.T) {
	e := searchTestEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setAgentMainPhasePriority(g, game.Player1)
	reg := cards.NewDefaultRegistry()
	card := addObservedHandCard(g, game.Player1, reg.Lookup("Dismember"))
	bears := addObservedPermanent(g, game.Player2, reg.Lookup("Grizzly Bears"))
	addObservedPermanent(g, game.Player2, reg.Lookup("Elderwood Scion"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	life := g.Players[game.Player1].Life
	var agents [game.NumPlayers]rules.PlayerAgent
	agents[game.Player1] = castFirstSpellAgent{}

	result := e.Simulator().ResolvePriority(g, agents)

	if !result.Players[game.Player1].Graveyard.Contains(card) || result.Players[game.Player1].Life != life-4 {
		t.Fatalf("Dismember in graveyard=%t life=%d, want it cast for 4 life and resolved",
			result.Players[game.Player1].Graveyard.Contains(card), result.Players[game.Player1].Life)
	}
	if slices.ContainsFunc(result.Battlefield, func(permanent *game.Permanent) bool { return permanent.ObjectID == bears.ObjectID }) {
		t.Fatal("Grizzly Bears survived, want Dismember cast at it")
	}
}

// castFirstSpellAgent casts the first spell it is offered and makes every
// choice as GenericStrategy does.
type castFirstSpellAgent struct{}

func (castFirstSpellAgent) ChooseAction(_ rules.PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if act.Kind == action.ActionCastSpell {
			return act
		}
	}
	return action.Pass()
}

func (castFirstSpellAgent) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	return GenericStrategy{}.ChooseChoice(obs, request)
}

// TestStrategiesKeepProvisionalPhyrexianSpellTargets applies a cast action
// for a registered Phyrexian spell naming a target, with every choice made by
// GenericStrategy, BaselineStrategy or a RandomAgent. The target prompt after
// the Phyrexian announcements (CR 601.2b-c) offers the action's targets as the
// provisional default and other targets with which the cost can still be
// paid, so each strategy keeps the target it chose with the action rather
// than re-picking one by its own target policy: Mutagenic Growth pumps
// Player1's own Grizzly Bears rather than Player2's, and Dismember with three
// colorless mana stays aimed at Player2's Bears rather than Player2's
// Elderwood Scion.
func TestStrategiesKeepProvisionalPhyrexianSpellTargets(t *testing.T) {
	strategies := map[string]func() rules.PlayerAgent{
		"GenericStrategy":  func() rules.PlayerAgent { return Agent{Strategy: GenericStrategy{}} },
		"BaselineStrategy": func() rules.PlayerAgent { return Agent{Strategy: BaselineStrategy{}} },
		"RandomAgent":      func() rules.PlayerAgent { return NewRandomAgent(rand.New(rand.NewPCG(1, 2))) },
	}
	for _, tc := range []struct {
		name      string
		spell     string
		colorless int
		other     string
		ownTarget bool
	}{
		{name: "Mutagenic Growth at own Bears", spell: "Mutagenic Growth", other: "Grizzly Bears", ownTarget: true},
		{name: "Dismember at Bears with three colorless", spell: "Dismember", colorless: 3, other: "Elderwood Scion"},
	} {
		for strategyName, strategy := range strategies {
			t.Run(tc.name+" with "+strategyName, func(t *testing.T) {
				e := searchTestEngine()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				setAgentMainPhasePriority(g, game.Player1)
				reg := cards.NewDefaultRegistry()
				card := addObservedHandCard(g, game.Player1, reg.Lookup(tc.spell))
				bearsController := game.Player2
				if tc.ownTarget {
					bearsController = game.Player1
				}
				bears := addObservedPermanent(g, bearsController, reg.Lookup("Grizzly Bears"))
				other := addObservedPermanent(g, game.Player2, reg.Lookup(tc.other))
				g.Players[game.Player1].ManaPool.Add(mana.C, tc.colorless)
				want := []game.Target{game.PermanentTarget(bears.ObjectID)}
				var chosen action.Action
				var otherOffered bool
				for _, act := range e.Simulator().LegalActions(g, game.Player1) {
					cast, ok := act.CastSpellPayload()
					if !ok || cast.CardID != card {
						continue
					}
					if slices.Equal(cast.Targets, want) {
						chosen = act
					}
					otherOffered = otherOffered || slices.Equal(cast.Targets, []game.Target{game.PermanentTarget(other.ObjectID)})
				}
				if chosen.Kind != action.ActionCastSpell || !otherOffered {
					t.Fatalf("%s casts at the Bears offered=%t and at the %s offered=%t, want both", tc.spell,
						chosen.Kind == action.ActionCastSpell, tc.other, otherOffered)
				}
				var agents [game.NumPlayers]rules.PlayerAgent
				agents[game.Player1] = strategy()

				result, ok := e.Simulator().Apply(g, game.Player1, chosen, agents)

				obj, found := result.Stack.Peek()
				if !ok || !found || obj.SourceID != card || !slices.Equal(obj.Targets, want) {
					t.Fatalf("cast=%t stack top=%+v, want %s cast at the Bears %v as its action proposed, not the %s %d",
						ok, obj, tc.spell, want, tc.other, other.ObjectID)
				}
			})
		}
	}
}
