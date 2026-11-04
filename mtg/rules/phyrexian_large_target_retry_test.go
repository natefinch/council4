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
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Source review: Comprehensive Rules effective 2026-09-25. CR 601.2b: "If a
// cost that will be paid as the spell is being cast includes Phyrexian mana
// symbols, the player announces whether they intend to pay 2 life or a
// corresponding colored mana cost for each of those symbols"; CR 601.2c then
// has the player announce "their choice of an appropriate object or player for
// each target the spell requires"; CR 601.2f totals the cost "minus all cost
// reductions". CR 733.1: "If a player takes an illegal action or starts to
// take an action but can't legally complete it, the entire action is reversed
// and any payments already made are canceled ... If the action was casting a
// spell, the spell returns to the zone it came from." CR 733.2: "the player
// who had priority retains it and may take another action or pass. The player
// may redo the reversed action in a legal way or take any other action allowed
// by the rules." Oracle (pinned corpus): Dismember
// fd74f8eb-0253-42dd-8277-186d4934da38 {1}{B/P}{B/P} "Target creature gets
// -5/-5 until end of turn."; Elderwood Scion
// 99c8647f-e08d-4217-b637-3f8f2bffdaa3 {3}{G}{W} 4/4 "Spells you cast that
// target this creature cost {2} less to cast. Spells your opponents cast that
// target this creature cost {2} more to cast."; Grizzly Bears 2/2 vanilla.
// Engine gap #3576: a redo stays open with 64 or more legal choices of
// targets, whether every choice is enumerated or only a sample is.

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

// TestPhyrexianRedoKeptWithManyLegalTargets gives Player1 an Elderwood Scion,
// no mana and 40 life, and Player2 bears enough for 64 or more legal choices
// of targets, then has Player1 cast Dismember targeting the Scion, announcing
// each {B/P} as {B} until its fifth announcement and 2 life from then on, or
// always announcing {B}. Only targeting the Scion reduces the {1} away (CR
// 601.2f), so only announcing 2 life for both symbols with the Scion targeted
// can be paid. Each unpayable attempt is reversed (CR 733.1) and Player1 may
// redo it in a legal way (CR 733.2), so Dismember is eventually cast at the
// Scion for 4 life on both the priority and prepared copy routes.
//
// With 64 bears plus the Scion the 65 choices are every legal one, so the
// announcements already shown unpayable with every target are withheld from
// the redo as with fewer targets, and an announcement of {B} that isn't
// offered is an invalid choice that takes the offered 2 life: {B}{B}, then
// {B} and 2 life, then 2 life and {B} are each tried once. With {X} added to
// its cost, 70 bears leave a sample of 64 choices, which can't show that an
// announcement is unpayable with every legal target before it is tried, so
// each unpayable announcement stays offered after its first reversal. When it
// is redone, checking each target shows none it could be paid with, so only
// that announcement is then withheld, and the announcements still open,
// ultimately 2 life for both symbols, remain.
func TestPhyrexianRedoKeptWithManyLegalTargets(t *testing.T) {
	for _, route := range []string{"priority", "prepared copy"} {
		for _, tc := range []struct {
			sampled      bool
			lifeFrom     int
			wantAttempts int
			wantInvalid  int
		}{
			{lifeFrom: 5, wantAttempts: 3, wantInvalid: 1},
			{lifeFrom: 99, wantAttempts: 4, wantInvalid: 4},
			{sampled: true, lifeFrom: 5, wantAttempts: 3},
			{sampled: true, lifeFrom: 99, wantAttempts: 7, wantInvalid: 6},
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
					for _, cast := range agent.chosen {
						if castActionKey(cast) != castActionKey(agent.chosen[0]) {
							t.Fatalf("casts chosen=%+v, want every attempt to redo the same cast", agent.chosen)
						}
					}
				} else {
					cast := action.CastSpellAction{CardID: source, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
						Targets: []game.Target{target}}
					var reversals castAttempts
					for attempts < 20 {
						attempts++
						outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, &reversals)
						if outcome == actionApplied {
							break
						}
						if outcome != actionReversed || reversals.exhausted[castActionKey(cast)] {
							t.Fatalf("attempt %d outcome=%v exhausted=%t, want the cast reversed and still open",
								attempts, outcome, reversals.exhausted[castActionKey(cast)])
						}
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
	var reversals castAttempts
	for agent.casts < 20 {
		agent.casts++
		outcome, _ := s.e.attemptCastSpellWithChoices(s.g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, &reversals)
		if outcome == actionApplied {
			break
		}
		if outcome != actionReversed || reversals.exhausted[castActionKey(cast)] {
			t.Fatalf("attempt %d outcome=%v exhausted=%t, want the cast reversed and still open",
				agent.casts, outcome, reversals.exhausted[castActionKey(cast)])
		}
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

// TestPhyrexianRevisedTargetKeepsAnnouncedPayment has Player1, with
// {B}{B}{C} floating, cast Dismember announcing {B} for both {B/P} and
// preferring Player2's Elderwood Scion as its target, then the last Grizzly
// Bears. The Scion raises the cost by {2} (CR 601.2f) beyond the three mana
// Player1 has, so Dismember can't be cast at it with that announcement, but
// can be at any bear. Player1 may redo a reversed cast in any legal way (CR
// 733.2), so withholding the Scion with that announcement must leave every
// bear open with it; the cast must never be pushed to the 2 life it doesn't
// need. With 64 bears every choice of targets is known, so the Scion is never
// offered with that announcement. With {X} added and 70 bears only a sample
// is, so the first attempt at the Scion is reversed, and every later one,
// even after the caster insists on the Scion, offers only the bears with it.
func TestPhyrexianRevisedTargetKeepsAnnouncedPayment(t *testing.T) {
	for _, route := range []string{"priority", "prepared copy"} {
		for _, sampled := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/sampled=%t", route, sampled), func(t *testing.T) {
				bears, wantAttempts, wantOffered := maxSampledNumericTargetChoices, 1, []bool{false}
				if sampled {
					bears, wantAttempts, wantOffered = maxSampledNumericTargetChoices+6, 2, []bool{true, false}
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

// TestPhyrexianWithheldTargetOpenWithRevisedPayment has Player1, with
// {B}{B}{C} floating, cast Dismember with {X} added at Player2's Elderwood
// Scion among 70 Grizzly Bears, so its choices of targets are a sample. It
// first announces {B} for both {B/P}, which the Scion's {2} increase makes
// unpayable, then redoes the cast announcing 2 life for both (CR 733.2). The
// Scion is withheld only with the first announcement: with the second it is
// still offered and Dismember is cast at it for {3} and 4 life. Only the
// prepared copy's cast is covered: from hand, the castability check before the
// payment doesn't consider paying {B/P} with life while generic mana competes
// for the same {B}, so it rejects the cast at the Scion (#3397).
func TestPhyrexianWithheldTargetOpenWithRevisedPayment(t *testing.T) {
	for _, route := range []string{"prepared copy"} {
		t.Run(route, func(t *testing.T) {
			s := newRevisedTargetScenario(t, route, largeTargetDismember(cards.NewDefaultRegistry(), true), maxSampledNumericTargetChoices+6, true)
			scion := game.PermanentTarget(s.scion.ObjectID)
			agent := &revisedTargetAgent{priorityProbeAgent: &priorityProbeAgent{}, scion: scion,
				preferred: [][]game.Target{{scion}}, payLife: func(attempt int) bool { return attempt >= 2 }}
			attempts := s.run(t, route, agent, scion)

			player := s.g.Players[game.Player1]
			if attempts != 2 || !slices.Equal(agent.scionOffered, []bool{true, true}) || s.g.Execution.InvalidChoices != 0 {
				t.Fatalf("attempts=%d Scion offered=%v invalid choices=%d, want 2 attempts offering the Scion each time",
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

// TestPhyrexianWithheldTargetSetReplacedWithOpenDefault casts a Dismember
// variant with {X} added whose target spec takes one or two creatures, so a
// selection is a set of targets, with {B}{B}{C} floating while Player2
// controls an Elderwood Scion and ten Grizzly Bears: 66 choices of targets, of
// which only a sample is enumerated. Player1 announces {B} for both {B/P} and
// selects the Scion and the first bear, which the Scion makes unpayable, so
// the cast is reversed (CR 733.1). The redo offers every creature, since a
// set holding the Scion may still be open with others, but selecting that
// same withheld set again is invalid and replaced with an open default, so
// Dismember is cast without the Scion and without paying life.
func TestPhyrexianWithheldTargetSetReplacedWithOpenDefault(t *testing.T) {
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
			if attempts != 2 || !slices.Equal(agent.scionOffered, []bool{true, true}) || s.g.Execution.InvalidChoices != 1 {
				t.Fatalf("attempts=%d Scion offered=%v chosen=%v invalid choices=%d, want 2 attempts offering the Scion and the repeated set invalid",
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

// pairDismemberScenario is a prepared copy, held by Player1 with {B}{B}{C}
// floating, of a Dismember variant with {X} added whose target spec takes
// exactly two creatures, while Player2 controls scions Elderwood Scions and
// then bears Grizzly Bears. Its choices of targets are only a sample, and the
// proposal announcing {B} for both {B/P} is checked against every choice of
// targets as the cast does (see checkPayment).
type pairDismemberScenario struct {
	e         *Engine
	g         *game.Game
	def       *game.CardDef
	cast      action.CastSpellAction
	scions    []game.Target
	bears     []game.Target
	sample    [][]game.Target
	proposals castProposals
	attempts  castAttempts
	mana      spellCostProposal
}

func newPairDismemberScenario(t *testing.T, scions, bears int) *pairDismemberScenario {
	t.Helper()
	reg := cards.NewDefaultRegistry()
	def := *largeTargetDismember(reg, true)
	mode := def.SpellAbility.Val.Modes[0]
	mode.Targets = slices.Clone(mode.Targets)
	mode.Targets[0].MinTargets, mode.Targets[0].MaxTargets = 2, 2
	def.SpellAbility = opt.Val(mode.Ability())
	e, g := newEffectiveCascadeGame()
	s := &pairDismemberScenario{e: e, g: g, def: &def}
	for range scions {
		s.scions = append(s.scions, game.PermanentTarget(addCombatPermanent(g, game.Player2, reg.Lookup("Elderwood Scion")).ObjectID))
	}
	for range bears {
		s.bears = append(s.bears, game.PermanentTarget(addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears")).ObjectID))
	}
	result := targetChoicesForSpell(g, game.Player1, &def, nil, game.CastBranch{XValue: opt.Val(0)})
	if !result.sampled || len(result.choices) != maxSampledNumericTargetChoices {
		t.Fatalf("choices=%d sampled=%t, want a sample of %d", len(result.choices), result.sampled, maxSampledNumericTargetChoices)
	}
	s.sample = result.choices
	source := preparedLargeTargetToken(t, g, &def).ObjectID
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	s.cast = action.CastSpellAction{CardID: source, SourceZone: zone.Battlefield, Face: game.FaceAlternate}
	request, ok := preparedCopyCostRequest(g, game.Player1, s.cast)
	if !ok {
		t.Fatal("prepared copy has no cost request")
	}
	s.proposals = newCastProposals(&s.attempts, castActionKey(s.cast), nil)
	s.proposals.offerTargets(result)
	e.checkPayment(g, &s.proposals, []payment.SpellRequest{request})
	symbol := cost.PhyrexianMana(mana.B)
	s.mana = spellCostProposal{phyrexian: []phyrexianAnnouncement{{symbol: symbol}, {symbol: symbol}}}
	return s
}

// choose has agent choose the targets of an attempt made with proposal,
// defaulting to preferred, among the choices still open with it.
func (s *pairDismemberScenario) choose(agent PlayerAgent, proposal spellCostProposal, preferred []game.Target) ([]game.Target, bool) {
	return s.e.chooseAnnouncedSpellTargets(s.g, game.Player1, s.def, s.cast, preferred, nil,
		s.proposals.targetsFor(0, proposal), reboundKickerAgents(agent), &TurnLog{})
}

// TestPhyrexianWithheldTargetSetReplacedBeyondSearchedChoices has Player1,
// with {B}{B}{C} floating, cast a Dismember variant with {X} added whose
// target spec takes exactly two creatures, announcing {B} for both {B/P},
// while Player2 controls eleven Elderwood Scions and then two Grizzly Bears:
// 78 choices of targets, of which only the first 64 are sampled. Each Scion
// targeted raises the cost by {2} (CR 601.2f), so with that announcement the
// only payable choice is the two bears, beyond the sample and beyond any
// bounded search. After the first attempt at a Scion and a bear is reversed
// as unpayable (CR 733.1), the caster may redo the cast in any legal way (CR
// 733.2): each later selection of a Scion, whether the caster insists on it or
// keeps it as its provisional default, is invalid and replaced by the bear
// pair, which is never withheld, and the announcement is never withheld with
// every choice. This covers the target choice both the priority and prepared
// copy casts make; from hand, the bear pair's cast isn't proposed with this
// sample at all (#3678).
func TestPhyrexianWithheldTargetSetReplacedBeyondSearchedChoices(t *testing.T) {
	for _, insist := range []bool{true, false} {
		t.Run(fmt.Sprintf("insist=%t", insist), func(t *testing.T) {
			s := newPairDismemberScenario(t, 11, 2)
			first := []game.Target{s.scions[0], s.bears[0]}
			bearPair := slices.Clone(s.bears)
			if slices.ContainsFunc(s.sample, func(choice []game.Target) bool { return slices.Equal(choice, bearPair) }) {
				t.Fatal("sample holds the bear pair, want it beyond the sample")
			}
			for _, choice := range s.sample {
				if s.proposals.payableWithTargets(0, s.mana, choice) {
					t.Fatalf("sampled choice %v payable with {B} announced for both, want none", choice)
				}
			}
			if !s.proposals.payableWithTargets(0, s.mana, bearPair) {
				t.Fatal("bear pair unpayable with {B} announced for both, want it payable")
			}
			agent := &revisedTargetAgent{priorityProbeAgent: &priorityProbeAgent{}, scion: s.scions[0]}
			if insist {
				agent.preferred = [][]game.Target{first}
			}
			// The first attempt, made before anything is withheld, is
			// reversed as unpayable.
			if s.proposals.reverseUnpayable("attempt 1", 0, s.mana, first) != actionReversed {
				t.Fatal("first attempt not reversed")
			}
			for attempt := 2; attempt <= 4; attempt++ {
				selected, ok := s.choose(agent, s.mana, first)
				if !ok || !slices.Equal(selected, bearPair) {
					t.Fatalf("attempt %d selected=%v ok=%t, want the bear pair %v", attempt, selected, ok, bearPair)
				}
				if s.g.Execution.InvalidChoices != attempt-1 {
					t.Fatalf("attempt %d invalid choices=%d, want %d", attempt, s.g.Execution.InvalidChoices, attempt-1)
				}
				if s.proposals.proposalWithheld(0, s.mana) || s.proposals.withheld(0, s.mana, bearPair) {
					t.Fatalf("attempt %d: announcement withheld=%t, bear pair withheld=%t, want both open", attempt,
						s.proposals.proposalWithheld(0, s.mana), s.proposals.withheld(0, s.mana, bearPair))
				}
				if s.attempts.exhausted[castActionKey(s.cast)] {
					t.Fatalf("attempt %d exhausted the cast", attempt)
				}
			}
		})
	}
}

// TestPhyrexianWithheldAnnouncementOnlyWhenNoTargetsOpen is
// TestPhyrexianWithheldTargetSetReplacedBeyondSearchedChoices with twelve
// Elderwood Scions and no bears: 66 choices of targets, every one holding a
// Scion, so none is payable with {B} announced for both {B/P}. Once the first
// attempt is reversed as unpayable, searching every choice finds none open,
// so the redo's choice of targets fails and only that announcement is withheld
// (CR 733.1): the caster may still redo the cast announcing 2 life for both,
// with which a Scion and another Scion are still offered (CR 733.2).
func TestPhyrexianWithheldAnnouncementOnlyWhenNoTargetsOpen(t *testing.T) {
	s := newPairDismemberScenario(t, 12, 0)
	first := s.scions[:2]
	agent := &revisedTargetAgent{priorityProbeAgent: &priorityProbeAgent{}, scion: s.scions[0],
		preferred: [][]game.Target{first}}
	if s.proposals.reverseUnpayable("attempt 1", 0, s.mana, first) != actionReversed {
		t.Fatal("first attempt not reversed")
	}
	if selected, ok := s.choose(agent, s.mana, first); ok {
		t.Fatalf("selected=%v, want no open choice of targets with {B} announced for both", selected)
	}
	if !s.proposals.proposalWithheld(0, s.mana) {
		t.Fatal("announcement still open, want it withheld once no choice of targets is open with it")
	}
	symbol := cost.PhyrexianMana(mana.B)
	life := spellCostProposal{phyrexian: []phyrexianAnnouncement{{symbol: symbol, payLife: true}, {symbol: symbol, payLife: true}}}
	if s.proposals.proposalWithheld(0, life) {
		t.Fatal("2 life announcement withheld, want it still open")
	}
	selected, ok := s.choose(agent, life, first)
	if !ok || !slices.Equal(selected, first) || s.attempts.exhausted[castActionKey(s.cast)] {
		t.Fatalf("selected=%v ok=%t exhausted=%t, want the Scions %v chosen with 2 life announced",
			selected, ok, s.attempts.exhausted[castActionKey(s.cast)], first)
	}
}
