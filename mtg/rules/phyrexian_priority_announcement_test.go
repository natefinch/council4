package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Comprehensive Rules effective 2026-09-25. CR 601.2b: "If a
// cost that will be paid as the spell is being cast includes Phyrexian mana
// symbols, the player announces whether they intend to pay 2 life or a
// corresponding colored mana cost for each of those symbols"; CR 601.2c then
// has the player announce "their choice of an appropriate object or player for
// each target the spell requires"; CR 601.2f: "The total cost is the mana cost
// or alternative cost (as determined in rule 601.2b), plus all additional
// costs and cost increases, and minus all cost reductions"; CR 601.2h pays the
// total cost. CR 733.1: "If a player takes an illegal action or starts to take
// an action but can't legally complete it, the entire action is reversed and
// any payments already made are canceled ... If the action was casting a
// spell, the spell returns to the zone it came from." CR 733.2: "the player
// who had priority retains it and may take another action or pass. The player
// may redo the reversed action in a legal way or take any other action allowed
// by the rules." Scryfall ruling for Vault Skirge (2011-06-01): "you choose how
// to pay for each Phyrexian mana symbol at the same time you would choose
// modes or choose a value for X". Oracle (pinned corpus): Dismember
// fd74f8eb-0253-42dd-8277-186d4934da38 {1}{B/P}{B/P} "Target creature gets
// -5/-5 until end of turn."; Elderwood Scion
// 99c8647f-e08d-4217-b637-3f8f2bffdaa3 "Spells you cast that target this
// creature cost {2} less to cast. Spells your opponents cast that target this
// creature cost {2} more to cast."; Gitaxian Probe
// 1d67f5ff-1fce-45e5-b6a1-416c569351e2 {U/P} "Look at target player's hand.
// Draw a card."; K'rrik, Son of Yawgmoth cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd
// "For each {B} in a cost, you may pay 2 life rather than pay that mana.";
// Vault Skirge 64cf5d59-7bcd-4b0b-a160-c8468d4c0f60 {1}{B/P}; Etherium
// Sculptor "Artifact spells you cast cost {1} less to cast."; Grizzly Bears
// 2/2 vanilla. These scenarios cover Engine gap #3573 on the priority cast and
// prepared copy routes; the free cast route remains #3570.

// TestPriorityPhyrexianAnnouncedBeforeTargetDependentCost takes the registered
// Dismember ({1}{B/P}{B/P}) cast offered with priority while Player1 controls
// the registered Elderwood Scion and has no mana. Only targeting the Scion
// makes the cost payable (CR 601.2f), so that is the only cast offered. Both
// {B/P} payments are still announced before the target is chosen (CR
// 601.2b-c), with the offered target as the default. Announcing 2 life for
// each, or taking the defaults, casts it at the Scion for 4 life. An
// announced {B} is binding (CR 601.2h), so that cast is reversed to hand
// without paying life (CR 733.1). Player2's Grizzly Bears is a legal target,
// but with it the {1} can't be paid after announcing life, so it is not
// offered in the target prompt and the Scion is targeted.
func TestPriorityPhyrexianAnnouncedBeforeTargetDependentCost(t *testing.T) {
	for _, tc := range []struct {
		name        string
		agent       *phyrexianWindowAgent
		chooseBears bool
		wantCast    bool
	}{
		{name: "announcing 2 life", agent: &phyrexianWindowAgent{payLife: true}, wantCast: true},
		{name: "default announcements", agent: &phyrexianWindowAgent{}, wantCast: true},
		{name: "announcing mana", agent: &phyrexianWindowAgent{payMana: true}},
		{name: "announcing 2 life then trying to target the Bears", agent: &phyrexianWindowAgent{payLife: true}, chooseBears: true, wantCast: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			reg := cards.NewDefaultRegistry()
			scion := addCombatPermanent(g, game.Player1, reg.Lookup("Elderwood Scion"))
			bears := addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears"))
			cardID := addCardToHand(g, game.Player1, reg.Lookup("Dismember"))
			var casts []action.Action
			for _, act := range e.legalActions(g, game.Player1) {
				if cast, ok := act.CastSpellPayload(); ok && cast.CardID == cardID {
					casts = append(casts, act)
				}
			}
			offered, _ := casts[0].CastSpellPayload()
			if len(casts) != 1 || !slices.Equal(offered.Targets, []game.Target{game.PermanentTarget(scion.ObjectID)}) {
				t.Fatalf("Dismember casts=%+v, want one cast targeting the Scion", casts)
			}
			agent := &announcedTargetAgent{phyrexianWindowAgent: tc.agent}
			if tc.chooseBears {
				agent.target = opt.Val(game.PermanentTarget(bears.ObjectID))
			}
			life := g.Players[game.Player1].Life

			cast := e.applyActionWithChoices(g, game.Player1, casts[0], reboundKickerAgents(agent), &TurnLog{})

			if !slices.Equal(tc.agent.steps, []string{"phyrexian", "phyrexian", "target"}) || cast != tc.wantCast {
				t.Fatalf("steps=%v cast=%t, want both announcements before the target and cast %t", tc.agent.steps, cast, tc.wantCast)
			}
			if tc.wantCast && (len(agent.offered) != 1 || !slices.Equal(agent.offered[0], []game.Target{game.PermanentTarget(scion.ObjectID)})) {
				t.Fatalf("offered targets=%v, want only the Scion, the one target with which the cost can be paid", agent.offered)
			}
			if !tc.wantCast {
				if g.Players[game.Player1].Life != life || !g.Players[game.Player1].Hand.Contains(cardID) || !g.Stack.IsEmpty() {
					t.Fatalf("life=%d stack=%d, want Dismember reversed to hand and life %d", g.Players[game.Player1].Life, g.Stack.Size(), life)
				}
				return
			}
			obj, _ := g.Stack.Peek()
			if g.Players[game.Player1].Life != life-4 || obj == nil || obj.SourceID != cardID || obj.ManaSpentToCast != 0 ||
				!slices.Equal(obj.Targets, []game.Target{game.PermanentTarget(scion.ObjectID)}) {
				t.Fatalf("life=%d stack top=%+v, want Dismember targeting the Scion for %d life", g.Players[game.Player1].Life, obj, life-4)
			}
		})
	}
}

// announcedTargetAgent is a phyrexianWindowAgent that records the targets
// offered in the spell target prompt and the provisional default there, if
// any, and, when target is set and offered, chooses it there.
type announcedTargetAgent struct {
	*phyrexianWindowAgent
	target      opt.V[game.Target]
	offered     [][]game.Target
	provisional [][]game.Target
}

func (a *announcedTargetAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	selected := a.phyrexianWindowAgent.ChooseChoice(obs, request)
	if request.Prompt == "Choose spell targets." {
		for _, option := range request.Options {
			a.offered = append(a.offered, option.Targets)
		}
		if request.Provisional {
			var provisional []game.Target
			for _, position := range request.DefaultSelection {
				provisional = append(provisional, request.Options[position].Targets...)
			}
			a.provisional = append(a.provisional, provisional)
		}
	}
	if target, ok := a.target.Val, a.target.Exists; ok && request.Prompt == "Choose spell targets." {
		for _, option := range request.Options {
			if slices.Equal(option.Targets, []game.Target{target}) {
				return []int{option.Index}
			}
		}
	}
	return selected
}

// TestPriorityPhyrexianAnnouncedBeforeTargets casts the registered Gitaxian
// Probe from hand with priority, proposing Player2 as its target. The {U/P}
// payment is announced before the target is chosen (CR 601.2b-c), and the
// proposed target remains the default. Announcing 2 life casts it for no
// mana; announcing {U} pays the {U}. Without blue mana the default
// announcement is 2 life. An announced {U} is binding (CR 601.2h): without
// blue mana the cast is reversed (CR 733.1) rather than paid with life.
func TestPriorityPhyrexianAnnouncedBeforeTargets(t *testing.T) {
	for _, tc := range []struct {
		name         string
		payLife      bool
		payMana      bool
		poolU        int
		wantCast     bool
		wantLifeLost int
		wantMana     int
	}{
		{name: "life", payLife: true, wantCast: true, wantLifeLost: 2},
		{name: "mana", payMana: true, poolU: 1, wantCast: true, wantMana: 1},
		{name: "default announcement with blue mana", poolU: 1, wantCast: true, wantMana: 1},
		{name: "default announcement without blue mana", wantCast: true, wantLifeLost: 2},
		{name: "announced mana is binding", payMana: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			cardID := addCardToHand(g, game.Player1, cards.NewDefaultRegistry().Lookup("Gitaxian Probe"))
			g.Players[game.Player1].ManaPool.Add(mana.U, tc.poolU)
			life := g.Players[game.Player1].Life
			agent := &phyrexianWindowAgent{payLife: tc.payLife, payMana: tc.payMana}

			outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, action.CastSpellAction{
				CardID: cardID, SourceZone: zone.Hand, Targets: []game.Target{game.PlayerTarget(game.Player2)},
			}, reboundKickerAgents(agent), &TurnLog{}, nil)
			ok := outcome == actionApplied
			if !ok && outcome != actionReversed {
				t.Fatalf("outcome=%v, want the cast applied or reversed", outcome)
			}

			if !slices.Equal(agent.steps, []string{"phyrexian", "target"}) {
				t.Fatalf("steps=%v, want the {U/P} announced before the target", agent.steps)
			}
			if ok != tc.wantCast || g.Players[game.Player1].Life != life-tc.wantLifeLost {
				t.Fatalf("cast=%t life=%d, want cast=%t and life %d", ok, g.Players[game.Player1].Life, tc.wantCast, life-tc.wantLifeLost)
			}
			if !tc.wantCast {
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Hand.Contains(cardID) || g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatal("Gitaxian Probe left hand although its announced {U} was not paid")
				}
				return
			}
			obj, _ := g.Stack.Peek()
			if obj == nil || obj.SourceID != cardID || obj.ManaSpentToCast != tc.wantMana ||
				!slices.Equal(obj.Targets, []game.Target{game.PlayerTarget(game.Player2)}) {
				t.Fatalf("stack top=%+v, want Gitaxian Probe targeting Player2 with %d mana spent", obj, tc.wantMana)
			}
		})
	}
}

// krrikManaThenLifeAgent announces each Phyrexian symbol as mana, then takes
// K'rrik's 2 life option for it while paying.
type krrikManaThenLifeAgent struct {
	phyrexianWindowAgent
}

func (a *krrikManaThenLifeAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.payLife = slices.Contains(a.steps, "phyrexian")
	return a.phyrexianWindowAgent.ChooseChoice(obs, request)
}

// TestPriorityKrrikPaysLifeForAnnouncedMana casts the registered Vault Skirge
// from hand with priority, Etherium Sculptor removing its {1}. Its {B/P} is
// announced as {B}; K'rrik, Son of Yawgmoth still lets the caster pay 2 life
// rather than that {B} while paying (CR 601.2h), so the {B} stays in the pool.
func TestPriorityKrrikPaysLifeForAnnouncedMana(t *testing.T) {
	reg := cards.NewDefaultRegistry()
	e, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, reg.Lookup("Etherium Sculptor"))
	addCombatPermanent(g, game.Player1, reg.Lookup("K'rrik, Son of Yawgmoth"))
	cardID := addCardToHand(g, game.Player1, reg.Lookup("Vault Skirge"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	life := g.Players[game.Player1].Life
	agent := &krrikManaThenLifeAgent{}

	ok := e.applyCastSpellWithChoices(g, game.Player1, action.CastSpellAction{CardID: cardID, SourceZone: zone.Hand},
		reboundKickerAgents(agent), &TurnLog{})

	obj, found := g.Stack.Peek()
	if !ok || !found || obj.SourceID != cardID || obj.ManaSpentToCast != 0 {
		t.Fatalf("cast=%t stack top=%+v, want Vault Skirge cast without spending mana", ok, obj)
	}
	if !slices.Equal(agent.steps, []string{"phyrexian", "phyrexian"}) || g.Players[game.Player1].Life != life-2 ||
		g.Players[game.Player1].ManaPool.Amount(mana.B) != 1 {
		t.Fatalf("steps=%v life=%d black=%d, want {B} announced, K'rrik's life option taken, life %d and {B} kept",
			agent.steps, g.Players[game.Player1].Life, g.Players[game.Player1].ManaPool.Amount(mana.B), life-2)
	}
}

// phyrexianPreparedToken creates a prepared token whose spell face is {U/P}
// "Look at target player's hand. Draw a card."
func phyrexianPreparedToken(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	return preparedSpellToken(t, g, "Gitaxian Probe")
}

// preparedSpellToken creates a prepared token for Player1 whose spell face has
// the mana cost and spell ability of the registered card named spell.
func preparedSpellToken(t *testing.T, g *game.Game, spell string) *game.Permanent {
	t.Helper()
	spellDef := cards.NewDefaultRegistry().Lookup(spell)
	def := prepareCard()
	alternate := def.Alternate.Val
	alternate.ManaCost = spellDef.ManaCost
	alternate.SpellAbility = spellDef.SpellAbility
	def.Alternate = opt.Val(alternate)
	token, ok := createTokenPermanent(g, game.Player1, copyCardDef(def))
	if !ok || !token.Prepared {
		t.Fatalf("token=%+v, want a prepared token", token)
	}
	return token
}

// TestPhyrexianTargetsOfferedOnlyWithPayableCost casts Dismember
// ({1}{B/P}{B/P}) with Player1 announcing 2 life for both {B/P} and proposing
// Player2's Grizzly Bears as its target, from hand with priority and as the
// spell copy of a prepared token. Player2's Elderwood Scion is also a legal
// target, but spells Player1 casts that target it cost {2} more (CR 601.2f).
// With one colorless mana the cost can't be paid with the Scion as target, so
// the target prompt after the announcements (CR 601.2b-c) offers only the
// Bears, and an agent that would rather target the Scion casts Dismember at
// the Bears instead of starting a cast that must be reversed (CR 733.1). With
// three colorless mana both are offered and choosing the Scion casts it there.
// Either way the proposed Bears is the prompt's provisional default, which an
// agent keeps unless it deliberately changes its target.
func TestPhyrexianTargetsOfferedOnlyWithPayableCost(t *testing.T) {
	for _, route := range []string{"priority", "prepared copy"} {
		for _, tc := range []struct {
			colorless  int
			wantScion  bool
			wantTarget string
		}{
			{colorless: 1, wantTarget: "Bears"},
			{colorless: 3, wantScion: true, wantTarget: "Scion"},
		} {
			t.Run(fmt.Sprintf("%s with %d colorless", route, tc.colorless), func(t *testing.T) {
				e, g := newEffectiveCascadeGame()
				reg := cards.NewDefaultRegistry()
				scion := addCombatPermanent(g, game.Player2, reg.Lookup("Elderwood Scion"))
				bears := addCombatPermanent(g, game.Player2, reg.Lookup("Grizzly Bears"))
				targets := map[string]game.Target{
					"Scion": game.PermanentTarget(scion.ObjectID), "Bears": game.PermanentTarget(bears.ObjectID),
				}
				cast := action.CastSpellAction{Targets: []game.Target{targets["Bears"]}}
				var source id.ID
				if route == "priority" {
					source = addCardToHand(g, game.Player1, reg.Lookup("Dismember"))
					cast.CardID, cast.SourceZone = source, zone.Hand
				} else {
					source = preparedSpellToken(t, g, "Dismember").ObjectID
					cast.CardID, cast.SourceZone, cast.Face = source, zone.Battlefield, game.FaceAlternate
				}
				g.Players[game.Player1].ManaPool.Add(mana.C, tc.colorless)
				life := g.Players[game.Player1].Life
				agent := &announcedTargetAgent{phyrexianWindowAgent: &phyrexianWindowAgent{payLife: true}, target: opt.Val(targets["Scion"])}

				outcome, taken := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, nil)

				offered := func(target game.Target) bool {
					return slices.ContainsFunc(agent.offered, func(option []game.Target) bool {
						return slices.Equal(option, []game.Target{target})
					})
				}
				if !slices.Equal(agent.steps, []string{"phyrexian", "phyrexian", "target"}) ||
					!offered(targets["Bears"]) || offered(targets["Scion"]) != tc.wantScion {
					t.Fatalf("steps=%v offered=%v, want both announcements before the target, the Bears offered and the Scion offered=%t",
						agent.steps, agent.offered, tc.wantScion)
				}
				if len(agent.provisional) != 1 || !slices.Equal(agent.provisional[0], []game.Target{targets["Bears"]}) {
					t.Fatalf("provisional defaults=%v, want the proposed Bears", agent.provisional)
				}
				want := []game.Target{targets[tc.wantTarget]}
				obj, _ := g.Stack.Peek()
				if outcome != actionApplied || !slices.Equal(taken, want) || obj == nil || obj.SourceID != source ||
					!slices.Equal(obj.Targets, want) || obj.ManaSpentToCast != tc.colorless || g.Players[game.Player1].Life != life-4 {
					t.Fatalf("outcome=%v taken=%v stack top=%+v life=%d, want Dismember cast at the %s for %d mana and 4 life",
						outcome, taken, obj, g.Players[game.Player1].Life, tc.wantTarget, tc.colorless)
				}
			})
		}
	}
}

// TestPreparedCopyPhyrexianAnnouncedBeforeTargets casts the spell copy of a
// prepared token whose spell face is {U/P} "Look at target player's hand. Draw
// a card.", proposing Player2 as its target. The {U/P} payment is announced
// before the target is chosen (CR 601.2b-c), and the proposed target remains
// the default. Announcing 2 life casts it. Announcing {U} without blue mana is
// binding, so that cast is reversed (CR 733.1): the token stays prepared and no
// life is paid, and repeating the same attempt is recorded as repeated.
func TestPreparedCopyPhyrexianAnnouncedBeforeTargets(t *testing.T) {
	t.Run("announcing 2 life", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		token := phyrexianPreparedToken(t, g)
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{payLife: true}
		cast := action.CastSpellAction{CardID: token.ObjectID, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
			Targets: []game.Target{game.PlayerTarget(game.Player2)}}

		if !e.applyCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}) {
			t.Fatal("casting the prepared spell copy failed")
		}

		obj, _ := g.Stack.Peek()
		if !slices.Equal(agent.steps, []string{"phyrexian", "target"}) || g.Players[game.Player1].Life != life-2 || token.Prepared ||
			obj == nil || obj.SourceID != token.ObjectID || !slices.Equal(obj.Targets, []game.Target{game.PlayerTarget(game.Player2)}) {
			t.Fatalf("steps=%v life=%d stack top=%+v, want the {U/P} announced before the target, life %d and Player2 targeted",
				agent.steps, g.Players[game.Player1].Life, obj, life-2)
		}
	})
	t.Run("announced mana is binding", func(t *testing.T) {
		e, g := newEffectiveCascadeGame()
		token := phyrexianPreparedToken(t, g)
		life := g.Players[game.Player1].Life
		agent := &phyrexianWindowAgent{payMana: true}
		cast := action.CastSpellAction{CardID: token.ObjectID, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
			Targets: []game.Target{game.PlayerTarget(game.Player2)}}
		var attempts castAttempts

		for attempt := range 2 {
			outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, &attempts)
			if outcome != actionReversed || attempts.repeated != (attempt == 1) {
				t.Fatalf("attempt %d outcome=%v repeated=%t, want reversed, repeated only on the second", attempt, outcome, attempts.repeated)
			}
			if !token.Prepared || !g.Stack.IsEmpty() || g.Players[game.Player1].Life != life {
				t.Fatalf("prepared=%t stack=%d life=%d, want the cast reversed", token.Prepared, g.Stack.Size(), g.Players[game.Player1].Life)
			}
		}
		if !slices.Equal(agent.steps, []string{"phyrexian", "target", "phyrexian", "target"}) {
			t.Fatalf("steps=%v, want each attempt to announce the {U/P} before its target", agent.steps)
		}
	})
}

// priorityProbeAgent takes the offered cast of its card with priority up to
// maxCasts times, announcing each Phyrexian symbol as mana until its
// lifeFrom-th announcement, which it pays with 2 life, and choosing target in
// the target prompt when set. As an observer it records the casts it sees.
type priorityProbeAgent struct {
	card          id.ID
	maxCasts      int
	lifeFrom      int
	target        opt.V[game.Target]
	casts         int
	announcements int
	steps         []string
	observed      []action.CastSpellAction
}

func (a *priorityProbeAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Player != game.Player1 || a.casts >= a.maxCasts {
		return action.Pass()
	}
	for _, act := range legal {
		if cast, ok := act.CastSpellPayload(); ok && cast.CardID == a.card {
			a.casts++
			a.steps = append(a.steps, "cast")
			return act
		}
	}
	return action.Pass()
}

func (a *priorityProbeAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case strings.HasPrefix(request.Prompt, "Pay {") && strings.HasSuffix(request.Prompt, "/P}"):
		a.announcements++
		a.steps = append(a.steps, "phyrexian")
		if a.announcements >= a.lifeFrom {
			for _, option := range request.Options {
				if strings.Contains(option.Label, "Pay 2 life") {
					return []int{option.Index}
				}
			}
		}
		return []int{0}
	case request.Prompt == "Choose spell targets.":
		a.steps = append(a.steps, "target")
		if target, ok := a.target.Val, a.target.Exists; ok {
			for _, option := range request.Options {
				if slices.Equal(option.Targets, []game.Target{target}) {
					return []int{option.Index}
				}
			}
		}
	}
	return request.DefaultSelection
}

func (a *priorityProbeAgent) ObserveAction(_ game.PlayerID, act action.Action, _ PlayerObservation) {
	if cast, ok := act.CastSpellPayload(); ok {
		a.observed = append(a.observed, cast)
	}
}

// TestPriorityLoopReversesUnpayablePhyrexianAnnouncement runs Player1's
// priority with the registered Gitaxian Probe in hand and no blue mana. Its
// {U/P} is announced as {U}, then its target chosen; the {U} can't be paid,
// so the cast is reversed rather than failing the priority loop (CR 733.1).
// Player1 keeps priority (CR 733.2) and may redo the cast announcing 2 life,
// which casts and resolves it, logging only that cast with the target chosen
// in the target prompt, which observers also see. Redoing an offered cast
// with the same announcement and target would only be reversed again, so that
// offer is then withdrawn; once every offer (one per player as the default
// target) is withdrawn, priority moves on with Probe still in hand.
func TestPriorityLoopReversesUnpayablePhyrexianAnnouncement(t *testing.T) {
	for _, tc := range []struct {
		name      string
		maxCasts  int
		lifeFrom  int
		wantCasts int
		wantCast  bool
	}{
		{name: "redo announcing life", maxCasts: 2, lifeFrom: 2, wantCasts: 2, wantCast: true},
		{name: "repeating the reversed cast", maxCasts: 99, lifeFrom: 99, wantCasts: 1 + game.NumPlayers},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, g := newEffectiveCascadeGame()
			probeID := addCardToHand(g, game.Player1, cards.NewDefaultRegistry().Lookup("Gitaxian Probe"))
			addCardToLibrary(g, game.Player1, cards.NewDefaultRegistry().Lookup("Lightning Bolt"))
			life := g.Players[game.Player1].Life
			target := game.PlayerTarget(game.Player3)
			agent := &priorityProbeAgent{card: probeID, maxCasts: tc.maxCasts, lifeFrom: tc.lifeFrom, target: opt.Val(target)}
			observer := &priorityProbeAgent{}
			agents := reboundKickerAgents(agent)
			agents[game.Player2] = observer
			log := &TurnLog{}

			e.runPriorityLoop(g, agents, log)

			var wantSteps []string
			for range tc.wantCasts {
				wantSteps = append(wantSteps, "cast", "phyrexian", "target")
			}
			if !slices.Equal(agent.steps, wantSteps) {
				t.Fatalf("steps=%v, want %v", agent.steps, wantSteps)
			}
			var logged []action.CastSpellAction
			for _, entry := range log.Actions {
				if cast, ok := entry.Action.CastSpellPayload(); ok {
					logged = append(logged, cast)
				}
			}
			spellCasts := 0
			for _, event := range g.Events {
				if event.Kind == game.EventSpellCast {
					spellCasts++
				}
			}
			if g.Execution.InvalidActions != 0 || !g.Stack.IsEmpty() {
				t.Fatalf("invalid=%d stack=%d, want no invalid actions and an empty stack", g.Execution.InvalidActions, g.Stack.Size())
			}
			if !tc.wantCast {
				if len(logged) != 0 || len(observer.observed) != 0 || spellCasts != 0 ||
					g.Players[game.Player1].Life != life || !g.Players[game.Player1].Hand.Contains(probeID) {
					t.Fatalf("logged=%+v observed=%+v casts=%d life=%d, want no cast, Gitaxian Probe in hand and life %d",
						logged, observer.observed, spellCasts, g.Players[game.Player1].Life, life)
				}
				return
			}
			if spellCasts != 1 || g.Players[game.Player1].Life != life-2 || !g.Players[game.Player1].Graveyard.Contains(probeID) {
				t.Fatalf("casts=%d life=%d, want Gitaxian Probe cast once for 2 life (%d) and resolved",
					spellCasts, g.Players[game.Player1].Life, life-2)
			}
			for name, casts := range map[string][]action.CastSpellAction{"turn log": logged, "observer": observer.observed} {
				if len(casts) != 1 || casts[0].CardID != probeID || !slices.Equal(casts[0].Targets, []game.Target{target}) {
					t.Fatalf("%s casts=%+v, want the one cast targeting %+v", name, casts, target)
				}
			}
		})
	}
}
