package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Comprehensive Rules effective 2026-11-06. CR 601.2b: "If a
// cost that will be paid as the spell is being cast includes Phyrexian mana
// symbols, the player announces whether they intend to pay 2 life or a
// corresponding colored mana cost for each of those symbols"; CR 601.2c then
// has the player announce "their choice of an appropriate object or player for
// each target the spell requires"; CR 601.2f totals the cost "minus all cost
// reductions". CR 734.1: "If a player takes an illegal action or starts to
// take an action but can't legally complete it, the entire action is reversed
// and any payments already made are canceled ... If the action was casting a
// spell, the spell returns to the zone it came from." CR 734.2: "the player
// who had priority retains it and may take another action or pass. The player
// may redo the reversed action in a legal way or take any other action allowed
// by the rules." Oracle (pinned corpus): Dismember
// fd74f8eb-0253-42dd-8277-186d4934da38 {1}{B/P}{B/P} "Target creature gets
// -5/-5 until end of turn."; Elderwood Scion
// 99c8647f-e08d-4217-b637-3f8f2bffdaa3 {3}{G}{W} 4/4 "Spells you cast that
// target this creature cost {2} less to cast. Spells your opponents cast that
// target this creature cost {2} more to cast."; Grizzly Bears 2/2 vanilla.
// Retained #3576 gameplay evidence: legal payment/target choices remain
// available beyond policy samples, now completing without a payment retry.

// largeTargetDismember returns the spell cast in a large-target retry
// scenario: the registered Dismember, whose single target is enumerated
// exhaustively, or, when sampled, Dismember with {X} added to its mana cost,
// whose choices of targets are enumerated only up to a policy sample.
func largeTargetDismember(reg *cards.Registry, sampled bool) *game.CardDef {
	dismember := reg.Lookup("Dismember")
	if !sampled {
		return dismember
	}
	def := *dismember
	def.Name = "Sampled Dismember"
	def.ManaCost = opt.Val(append(cost.Mana{cost.X}, dismember.ManaCost.Val...))
	return &def
}

// preparedLargeTargetToken returns a prepared token of Player1's whose spell
// face is spell.
func preparedLargeTargetToken(t *testing.T, g *game.Game, spell *game.CardDef) *game.Permanent {
	t.Helper()
	def := prepareCard()
	alternate := def.Alternate.Val
	alternate.ManaCost = spell.ManaCost
	alternate.SpellAbility = spell.SpellAbility
	def.Alternate = opt.Val(alternate)
	token, ok := createTokenPermanent(g, game.Player1, copyCardDef(def))
	if !ok || !token.Prepared {
		t.Fatalf("token=%+v, want a prepared token", token)
	}
	return token
}

// castRecordingAgent is a priorityProbeAgent that records each cast it
// chooses with priority.
type castRecordingAgent struct {
	*priorityProbeAgent
	chosen []action.CastSpellAction
}

func (a *castRecordingAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	act := a.priorityProbeAgent.ChooseAction(obs, legal)
	if cast, ok := act.CastSpellPayload(); ok {
		a.chosen = append(a.chosen, cast)
	}
	return act
}

// Only targeting Player1's Scion and announcing both symbols as life can pay
// this Dismember. Both routes must find that joint completion on the first
// attempt, even if the policy insists on mana and targets exceed its sample.
func TestPhyrexianRedoKeptWithManyLegalTargets(t *testing.T) {
	for _, route := range []string{"priority", "prepared copy"} {
		for _, tc := range []struct {
			sampled      bool
			lifeFrom     int
			wantAttempts int
			wantInvalid  int
		}{
			// CR 601.2b/c/f/h: every announcement has a payment witness.
			{lifeFrom: 5, wantAttempts: 1},
			{lifeFrom: 99, wantAttempts: 1},
			{sampled: true, lifeFrom: 5, wantAttempts: 1},
			{sampled: true, lifeFrom: 99, wantAttempts: 1},
		} {
			name := fmt.Sprintf("%s/%s/life from announcement %d", route,
				map[bool]string{false: "exhaustive", true: "sampled"}[tc.sampled], tc.lifeFrom)
			t.Run(name, func(t *testing.T) {
				e, g := newEffectiveCascadeGame()
				reg := cards.NewDefaultRegistry()
				spell := largeTargetDismember(reg, tc.sampled)
				scion := addCombatPermanent(g, game.Player1, reg.Lookup("Elderwood Scion"))
				bears := maxSampledNumericTargetChoices
				if tc.sampled {
					bears += 6
				}
				for range bears {
					addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears"))
				}
				result := targetChoicesForSpell(g, game.Player1, spell, nil, game.CastBranch{XValue: opt.Val(0)})
				if result.sampled != tc.sampled || len(result.choices) < maxSampledNumericTargetChoices ||
					!slices.ContainsFunc(result.choices, func(targets []game.Target) bool {
						return slices.Equal(targets, []game.Target{game.PermanentTarget(scion.ObjectID)})
					}) {
					t.Fatalf("choices=%d sampled=%t, want at least %d choices including the Scion, sampled=%t",
						len(result.choices), result.sampled, maxSampledNumericTargetChoices, tc.sampled)
				}
				var source id.ID
				if route == "priority" {
					source = addCardToHand(g, game.Player1, spell)
				} else {
					source = preparedLargeTargetToken(t, g, spell).ObjectID
				}
				target := game.PermanentTarget(scion.ObjectID)
				agent := &castRecordingAgent{priorityProbeAgent: &priorityProbeAgent{card: source, maxCasts: 99, lifeFrom: tc.lifeFrom, target: opt.Val(target)}}
				life := g.Players[game.Player1].Life

				attempts := 0
				if route == "priority" {
					e.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})
					attempts = agent.casts
				} else {
					cast := action.CastSpellAction{CardID: source, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
						Targets: []game.Target{target}}
					attempts++
					if outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}); outcome != actionApplied {
						t.Fatalf("outcome=%v, want a witnessed cast without retry", outcome)
					}
					e.resolveTopOfStack(g, &TurnLog{})
					e.applyStateBasedActions(g)
				}

				spellCasts := 0
				for _, event := range g.Events {
					if event.Kind == game.EventSpellCast {
						spellCasts++
						if event.SourceID != source {
							t.Fatalf("cast event=%+v, want only the Dismember cast", event)
						}
					}
				}
				if attempts != tc.wantAttempts || g.Execution.InvalidChoices != tc.wantInvalid || g.Execution.InvalidActions != 0 {
					t.Fatalf("attempts=%d invalid choices=%d actions=%d steps=%v, want %d attempts and %d invalid choices",
						attempts, g.Execution.InvalidChoices, g.Execution.InvalidActions, agent.steps, tc.wantAttempts, tc.wantInvalid)
				}
				if spellCasts != 1 || g.Players[game.Player1].Life != life-4 || !g.Stack.IsEmpty() {
					t.Fatalf("casts=%d life=%d stack=%d, want Dismember cast once for 4 life (%d) and resolved",
						spellCasts, g.Players[game.Player1].Life, g.Stack.Size(), life-4)
				}
				if _, ok := g.PermanentByID(scion.ObjectID); ok {
					t.Fatal("Elderwood Scion survived, want Dismember's -5/-5 to have resolved at it")
				}
			})
		}
	}
}

// revisedTargetAgent is a priorityProbeAgent that announces each Phyrexian
// symbol of its casts as mana, or as 2 life when payLife reports so for the
// attempt, and chooses the first of preferred whose targets are all offered.
// It records whether each target prompt offered the Scion.
type revisedTargetAgent struct {
	*priorityProbeAgent
	payLife      func(attempt int) bool
	preferred    [][]game.Target
	scion        game.Target
	scionOffered []bool
	chosen       [][]game.Target
}

func (a *revisedTargetAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case strings.HasPrefix(request.Prompt, "Pay {") && strings.HasSuffix(request.Prompt, "/P}"):
		life := a.payLife != nil && a.payLife(a.casts)
		for _, option := range request.Options {
			if strings.Contains(option.Label, "Pay 2 life") == life {
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	case request.Prompt == "Choose spell targets.":
		a.scionOffered = append(a.scionOffered, slices.ContainsFunc(request.Options, func(option game.ChoiceOption) bool {
			return slices.Equal(option.Targets, []game.Target{a.scion})
		}))
		for _, targets := range a.preferred {
			var selected []int
			for _, target := range targets {
				for _, option := range request.Options {
					if slices.Equal(option.Targets, []game.Target{target}) {
						selected = append(selected, option.Index)
					}
				}
			}
			if len(selected) == len(targets) && len(selected) >= request.MinChoices && len(selected) <= request.MaxChoices {
				a.chosen = append(a.chosen, targets)
				return selected
			}
		}
		a.chosen = append(a.chosen, nil)
		return request.DefaultSelection
	}
	return a.priorityProbeAgent.ChooseChoice(obs, request)
}

// revisedTargetScenario is a cast of a Dismember variant by Player1, with
// {B}{B}{C} floating, while Player2 controls an Elderwood Scion and bears.
type revisedTargetScenario struct {
	e      *Engine
	g      *game.Game
	scion  *game.Permanent
	bears  []*game.Permanent
	source id.ID
	life   int
}

func newRevisedTargetScenario(t *testing.T, route string, spell *game.CardDef, bears int, sampled bool) *revisedTargetScenario {
	t.Helper()
	e, g := newEffectiveCascadeGame()
	reg := cards.NewDefaultRegistry()
	s := &revisedTargetScenario{e: e, g: g, scion: addCombatPermanent(g, game.Player2, reg.Lookup("Elderwood Scion"))}
	for range bears {
		s.bears = append(s.bears, addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears")))
	}
	result := targetChoicesForSpell(g, game.Player1, spell, nil, game.CastBranch{XValue: opt.Val(0)})
	if result.sampled != sampled || len(result.choices) < maxSampledNumericTargetChoices {
		t.Fatalf("choices=%d sampled=%t, want at least %d choices, sampled=%t",
			len(result.choices), result.sampled, maxSampledNumericTargetChoices, sampled)
	}
	if route == "priority" {
		s.source = addCardToHand(g, game.Player1, spell)
	} else {
		s.source = preparedLargeTargetToken(t, g, spell).ObjectID
	}
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	s.life = g.Players[game.Player1].Life
	return s
}

// run casts the spell with agent until a cast is applied, through the
// priority loop or by redoing the prepared copy's cast with defaultTargets as
// its provisional targets, and resolves it. It returns the attempts made.
func (s *revisedTargetScenario) run(t *testing.T, route string, agent *revisedTargetAgent, defaultTargets ...game.Target) int {
	t.Helper()
	agent.card, agent.maxCasts = s.source, 20
	if route == "priority" {
		s.e.runPriorityLoop(s.g, reboundKickerAgents(agent), &TurnLog{})
		return agent.casts
	}
	cast := action.CastSpellAction{CardID: s.source, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
		Targets: defaultTargets}
	agent.casts++
	if outcome, _ := s.e.attemptCastSpellWithChoices(s.g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}); outcome != actionApplied {
		t.Fatalf("outcome=%v, want a witnessed cast without retry", outcome)
	}
	s.e.resolveTopOfStack(s.g, &TurnLog{})
	s.e.applyStateBasedActions(s.g)
	return agent.casts
}

// spellCasts returns the spells cast, failing unless each was the source.
func (s *revisedTargetScenario) spellCasts(t *testing.T) int {
	t.Helper()
	casts := 0
	for _, event := range s.g.Events {
		if event.Kind == game.EventSpellCast {
			casts++
			if event.SourceID != s.source {
				t.Fatalf("cast event=%+v, want only the Dismember cast", event)
			}
		}
	}
	return casts
}

// The Scion's tax makes the announced all-mana Dismember unpayable at it,
// but every bear remains a completion. Sampled and exhaustive policy domains
// must both exclude the Scion and permit the last bear without a retry or life.
func TestPhyrexianRevisedTargetKeepsAnnouncedPayment(t *testing.T) {
	for _, route := range []string{"priority", "prepared copy"} {
		for _, sampled := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/sampled=%t", route, sampled), func(t *testing.T) {
				bears, wantAttempts, wantOffered := maxSampledNumericTargetChoices, 1, []bool{false}
				if sampled {
					bears = maxSampledNumericTargetChoices + 6
				}
				s := newRevisedTargetScenario(t, route, largeTargetDismember(cards.NewDefaultRegistry(), sampled), bears, sampled)
				scion, bear := game.PermanentTarget(s.scion.ObjectID), game.PermanentTarget(s.bears[len(s.bears)-1].ObjectID)
				agent := &revisedTargetAgent{priorityProbeAgent: &priorityProbeAgent{}, scion: scion,
					preferred: [][]game.Target{{scion}, {bear}}}
				attempts := s.run(t, route, agent, scion)

				player := s.g.Players[game.Player1]
				if attempts != wantAttempts || !slices.Equal(agent.scionOffered, wantOffered) ||
					s.g.Execution.InvalidChoices != 0 || s.g.Execution.InvalidActions != 0 {
					t.Fatalf("attempts=%d Scion offered=%v invalid choices=%d actions=%d, want %d attempts, Scion offered %v, no invalid choices",
						attempts, agent.scionOffered, s.g.Execution.InvalidChoices, s.g.Execution.InvalidActions, wantAttempts, wantOffered)
				}
				if s.spellCasts(t) != 1 || player.Life != s.life || player.ManaPool.Total() != 0 || !s.g.Stack.IsEmpty() {
					t.Fatalf("casts=%d life=%d pool=%d stack=%d, want Dismember cast once for {B}{B}{C} and no life (%d)",
						s.spellCasts(t), player.Life, player.ManaPool.Total(), s.g.Stack.Size(), s.life)
				}
				if _, ok := s.g.PermanentByID(s.bears[len(s.bears)-1].ObjectID); ok {
					t.Fatalf("chosen=%v: the last Grizzly Bears survived, want Dismember resolved at it", agent.chosen)
				}
				if _, ok := s.g.PermanentByID(s.scion.ObjectID); !ok {
					t.Fatal("Elderwood Scion died, want Dismember resolved only at the bear")
				}
			})
		}
	}
}

// Announcing both symbols as life leaves enough mana to target the Scion,
// even among targets beyond the numeric policy sample.
func TestPhyrexianTargetOpenWithLifePayment(t *testing.T) {
	for _, route := range []string{"prepared copy"} {
		t.Run(route, func(t *testing.T) {
			s := newRevisedTargetScenario(t, route, largeTargetDismember(cards.NewDefaultRegistry(), true), maxSampledNumericTargetChoices+6, true)
			scion := game.PermanentTarget(s.scion.ObjectID)
			agent := &revisedTargetAgent{priorityProbeAgent: &priorityProbeAgent{}, scion: scion,
				preferred: [][]game.Target{{scion}}, payLife: func(int) bool { return true }}
			attempts := s.run(t, route, agent, scion)

			player := s.g.Players[game.Player1]
			if attempts != 1 || !slices.Equal(agent.scionOffered, []bool{true}) || s.g.Execution.InvalidChoices != 0 {
				t.Fatalf("attempts=%d Scion offered=%v invalid choices=%d, want one witnessed cast offering the Scion",
					attempts, agent.scionOffered, s.g.Execution.InvalidChoices)
			}
			if s.spellCasts(t) != 1 || player.Life != s.life-4 || player.ManaPool.Total() != 0 || !s.g.Stack.IsEmpty() {
				t.Fatalf("casts=%d life=%d pool=%d, want Dismember cast once for {3} and 4 life (%d)",
					s.spellCasts(t), player.Life, player.ManaPool.Total(), s.life-4)
			}
			if _, ok := s.g.PermanentByID(s.scion.ObjectID); ok {
				t.Fatal("Elderwood Scion survived, want Dismember resolved at it")
			}
		})
	}
}

// A one-or-two-target Dismember must exclude every Scion-containing set
// before selection, while retaining payable bear targets beyond the sample.
func TestPhyrexianTargetSetHasPaymentCompletion(t *testing.T) {
	for _, route := range []string{"priority", "prepared copy"} {
		t.Run(route, func(t *testing.T) {
			spell := largeTargetDismember(cards.NewDefaultRegistry(), true)
			def := *spell
			mode := def.SpellAbility.Val.Modes[0]
			mode.Targets = slices.Clone(mode.Targets)
			mode.Targets[0].MaxTargets = 2
			def.SpellAbility = opt.Val(mode.Ability())
			s := newRevisedTargetScenario(t, route, &def, 10, true)
			scion, bear := game.PermanentTarget(s.scion.ObjectID), game.PermanentTarget(s.bears[0].ObjectID)
			agent := &revisedTargetAgent{priorityProbeAgent: &priorityProbeAgent{}, scion: scion,
				preferred: [][]game.Target{{scion, bear}}}
			attempts := s.run(t, route, agent, bear)

			player := s.g.Players[game.Player1]
			// A Scion-containing set has no completion with mana announced.
			if attempts != 1 || slices.Contains(agent.scionOffered, true) || s.g.Execution.InvalidChoices != 0 {
				t.Fatalf("attempts=%d Scion offered=%v chosen=%v invalid choices=%d, want one witnessed cast without the Scion",
					attempts, agent.scionOffered, agent.chosen, s.g.Execution.InvalidChoices)
			}
			if s.spellCasts(t) != 1 || player.Life != s.life || player.ManaPool.Total() != 0 || !s.g.Stack.IsEmpty() {
				t.Fatalf("casts=%d life=%d pool=%d, want Dismember cast once for {B}{B}{C} and no life (%d)",
					s.spellCasts(t), player.Life, player.ManaPool.Total(), s.life)
			}
			if _, ok := s.g.PermanentByID(s.scion.ObjectID); !ok {
				t.Fatal("Elderwood Scion died, want Dismember cast without it")
			}
		})
	}
}
