package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt, SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), read
// 702.22a-m, 506.4, 508.1 and 508.1a-m. 508.1e: after choosing attackers and
// whom they attack and checking restrictions and requirements (508.1b-d), if any
// creatures have banding (or bands with other) the active player announces which
// creatures, if any, are banded; 508.1f onward taps attackers and pays costs.
// 702.22c: a band is any number of creatures with banding plus up to one
// creature without banding, all attacking the same player, planeswalker or
// battle, formed only as attackers are declared; in such a band a creature with
// only "bands with other" counts as one without banding ("bands with other"
// bands are covered by combat_bands_with_other_test.go). 702.22e: a band is a
// continuing relation that survives its members losing banding. 702.22f: a
// creature removed from combat is removed from its band, and the band is still a
// band only while at least two creatures remain. 702.22g-k (blocking and damage
// across a band) are separate capabilities. Card faces used: Benalish Hero
// ({W} Human Soldier 1/1, Banding), Mesa Pegasus ({1}{W} Pegasus 1/1, Flying,
// Banding), Merfolk Trickster ({U}{U} 2/2, Flash; "When this creature enters,
// tap target creature an opponent controls. It loses all abilities until end of
// turn."), War Elephant and Kjeldoran Escort (existing banding blockers).

type attackingBandBoard struct {
	g            *game.Game
	hero         *game.Permanent
	pegasus      *game.Permanent
	bear         *game.Permanent
	otherBear    *game.Permanent
	planeswalker *game.Permanent
}

func newAttackingBandBoard(t *testing.T) attackingBandBoard {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	board := attackingBandBoard{
		g:         g,
		hero:      registryCombatPermanent(t, g, game.Player1, "Benalish Hero"),
		pegasus:   registryCombatPermanent(t, g, game.Player1, "Mesa Pegasus"),
		bear:      addCombatCreaturePermanentWithPower(g, game.Player1, 2),
		otherBear: addCombatCreaturePermanentWithPower(g, game.Player1, 2),
		planeswalker: addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Band Test Planeswalker", Types: []types.Card{types.Planeswalker},
		}}),
	}
	for _, creature := range []*game.Permanent{board.hero, board.pegasus, board.bear, board.otherBear} {
		creature.SummoningSick = false
	}
	return board
}

func attackAt(attacker *game.Permanent, target game.AttackTarget) game.AttackDeclaration {
	return game.AttackDeclaration{Attacker: attacker.ObjectID, Target: target}
}

func bandOf(members ...*game.Permanent) game.AttackingBand {
	band := game.AttackingBand{}
	for _, member := range members {
		band.Members = append(band.Members, member.ObjectID)
	}
	return band
}

func sameBands(got, want []game.AttackingBand) bool {
	normalize := func(bands []game.AttackingBand) [][]id.ID {
		var sets [][]id.ID
		for _, band := range bands {
			members := slices.Clone(band.Members)
			slices.Sort(members)
			sets = append(sets, members)
		}
		slices.SortFunc(sets, func(a, b []id.ID) int { return slices.Compare(a, b) })
		return sets
	}
	return slices.EqualFunc(normalize(got), normalize(want), slices.Equal[[]id.ID])
}

// TestAttackingBandDeclaredRecordedObservedAndCloned declares Benalish Hero,
// Mesa Pegasus and a creature without banding as one band through the live
// declare-attackers turn-based action. The band is in the logged action and the
// game's combat state, every seat observes it, and a clone shares no storage.
func TestAttackingBandDeclaredRecordedObservedAndCloned(t *testing.T) {
	board := newAttackingBandBoard(t)
	g := board.g
	player2 := game.AttackTarget{Player: game.Player2}
	band := bandOf(board.hero, board.pegasus, board.bear)
	proposal := action.DeclareAttackersWithBands([]game.AttackDeclaration{
		attackAt(board.hero, player2), attackAt(board.pegasus, player2), attackAt(board.bear, player2),
	}, []game.AttackingBand{band})
	agents := simPassPolicies()
	agents[game.Player1] = &exactNumericPolicy{announcement: proposal}
	log := &TurnLog{}
	combatEngine{NewEngine(nil)}.declareAttackers(g, agents, log)

	if len(log.Actions) != 1 {
		t.Fatalf("logged %d actions, want the attack declaration", len(log.Actions))
	}
	logged, ok := log.Actions[0].Action.DeclareAttackersPayload()
	if !ok || !sameBands(logged.Bands, []game.AttackingBand{band}) {
		t.Fatalf("logged declaration bands = %+v, want %+v", logged.Bands, band)
	}
	if !sameBands(g.Combat.Bands, []game.AttackingBand{band}) || len(g.Combat.Attackers) != 3 {
		t.Fatalf("combat = %+v, want three attackers in one band", g.Combat)
	}
	for _, seat := range []game.PlayerID{game.Player1, game.Player2} {
		view := NewObservation(g, seat).Combat()
		if !sameBands(view.Bands, []game.AttackingBand{band}) {
			t.Fatalf("seat %d observed bands %+v, want %+v", seat, view.Bands, band)
		}
		view.Bands[0].Members[0] = 0
	}
	if g.Combat.Bands[0].Members[0] != board.hero.ObjectID {
		t.Fatal("mutating an observation changed the game's band")
	}

	clone := g.Clone()
	clone.Combat.Bands[0].Members[0] = board.otherBear.ObjectID
	clone.Combat.Bands = append(clone.Combat.Bands, bandOf(board.hero, board.otherBear))
	if !sameBands(g.Combat.Bands, []game.AttackingBand{band}) {
		t.Fatalf("mutating a clone changed the original band: %+v", g.Combat.Bands)
	}
	if !sameBands(g.Clone().Combat.Bands, []game.AttackingBand{band}) {
		t.Fatal("clone did not copy the band")
	}
}

// TestAttackingBandIllegalDeclarationsRejected checks every CR 702.22c rule for
// an announced band: banding membership, a shared attack target, a creature in
// at most one band, at least two members and every member declared.
func TestAttackingBandIllegalDeclarationsRejected(t *testing.T) {
	board := newAttackingBandBoard(t)
	player2 := game.AttackTarget{Player: game.Player2}
	player3 := game.AttackTarget{Player: game.Player3}
	walker := game.AttackTarget{Player: game.Player2, PlaneswalkerID: board.planeswalker.ObjectID}
	all := func(target game.AttackTarget) []game.AttackDeclaration {
		return []game.AttackDeclaration{
			attackAt(board.hero, target), attackAt(board.pegasus, target),
			attackAt(board.bear, target), attackAt(board.otherBear, target),
		}
	}
	for _, test := range []struct {
		name    string
		attacks []game.AttackDeclaration
		bands   []game.AttackingBand
		legal   bool
	}{
		{"banding plus one without", all(player2), []game.AttackingBand{bandOf(board.hero, board.pegasus, board.bear)}, true},
		{"band attacking a planeswalker", all(walker), []game.AttackingBand{bandOf(board.hero, board.bear)}, true},
		{"two bands", all(player2), []game.AttackingBand{bandOf(board.hero, board.bear), bandOf(board.pegasus, board.otherBear)}, true},
		{"two without banding", all(player2), []game.AttackingBand{bandOf(board.hero, board.bear, board.otherBear)}, false},
		{"no member with banding", all(player2), []game.AttackingBand{bandOf(board.bear, board.otherBear)}, false},
		{"members attack different players", []game.AttackDeclaration{
			attackAt(board.hero, player2), attackAt(board.pegasus, player3),
		}, []game.AttackingBand{bandOf(board.hero, board.pegasus)}, false},
		{"player and that player's planeswalker", []game.AttackDeclaration{
			attackAt(board.hero, player2), attackAt(board.pegasus, walker),
		}, []game.AttackingBand{bandOf(board.hero, board.pegasus)}, false},
		{"creature in two bands", all(player2), []game.AttackingBand{bandOf(board.hero, board.bear), bandOf(board.pegasus, board.bear)}, false},
		{"creature twice in one band", all(player2), []game.AttackingBand{bandOf(board.hero, board.hero)}, false},
		{"band of one", all(player2), []game.AttackingBand{bandOf(board.hero)}, false},
		{"undeclared member", []game.AttackDeclaration{attackAt(board.hero, player2)}, []game.AttackingBand{bandOf(board.hero, board.pegasus)}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := NewEngine(nil)
			proposal := action.DeclareAttackersWithBands(test.attacks, test.bands)
			payload, _ := proposal.DeclareAttackersPayload()
			if got := (combatEngine{engine}).attackersLegal(board.g, game.Player1, payload); got != test.legal {
				t.Fatalf("attackersLegal = %v, want %v", got, test.legal)
			}
			result, ok := engine.Simulator().Apply(board.g, game.Player1, proposal, simPassPolicies())
			if ok != test.legal {
				t.Fatalf("Simulator applied = %v, want %v", ok, test.legal)
			}
			if ok && !sameBands(result.Combat.Bands, test.bands) {
				t.Fatalf("applied bands %+v, want %+v", result.Combat.Bands, test.bands)
			}
			if len(board.g.Combat.Bands) != 0 || len(board.g.Combat.Attackers) != 0 {
				t.Fatal("Simulator changed its source game")
			}
		})
	}
}

// stagedBandPolicy declares attackers through the staged chooser, attacking
// with each creature in targets, and answers band choices from addTo (anchor ->
// creatures to add in order). It checks that every band request offers exactly
// the creatures that can still legally join the anchor's band.
type stagedBandPolicy struct {
	simPassPolicy

	t        *testing.T
	targets  map[id.ID]game.AttackTarget
	banding  map[id.ID]bool
	addTo    map[id.ID][]id.ID
	requests int

	anchor  id.ID
	current []id.ID
	banded  map[id.ID]bool
}

func (p *stagedBandPolicy) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.ChooseAttackers()
}

func (p *stagedBandPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceCombatDeclaration:
		target, attacking := p.targets[request.CombatAttacker]
		for _, option := range request.Options {
			if option.Attack.Exists == attacking && (!attacking || option.Attack.Val.Target == target) {
				return []int{option.Index}
			}
		}
		p.t.Fatalf("wanted attack was not offered: %+v", request)
	case game.ChoiceAttackingBand:
		p.requests++
		if request.CombatAttacker != p.anchor {
			p.commit()
			p.anchor, p.current = request.CombatAttacker, []id.ID{request.CombatAttacker}
		}
		if !p.banding[p.anchor] || p.banded[p.anchor] {
			p.t.Fatalf("band anchored at %d, which lacks banding or is already banded", p.anchor)
		}
		hasWithout := slices.ContainsFunc(p.current, func(member id.ID) bool { return !p.banding[member] })
		var want, offered []id.ID
		for attacker, target := range p.targets {
			if target == p.targets[p.anchor] && !p.banded[attacker] && !slices.Contains(p.current, attacker) &&
				(p.banding[attacker] || !hasWithout) {
				want = append(want, attacker)
			}
		}
		if request.DefaultSelection == nil || request.DefaultSelection[0] != 0 || request.Options[0].Attack.Exists {
			p.t.Fatalf("option 0 is not the default finish: %+v", request)
		}
		for _, option := range request.Options[1:] {
			if !option.Attack.Exists || !option.Card.Exists {
				p.t.Fatalf("band candidate lacks its public attack or card: %+v", option)
			}
			offered = append(offered, option.Attack.Val.Attacker)
		}
		slices.Sort(want)
		slices.Sort(offered)
		if !slices.Equal(want, offered) {
			p.t.Fatalf("anchor %d with %v offered %v, want %v", p.anchor, p.current, offered, want)
		}
		for _, next := range p.addTo[p.anchor] {
			if slices.Contains(p.current, next) {
				continue
			}
			for _, option := range request.Options[1:] {
				if option.Attack.Val.Attacker == next {
					p.current = append(p.current, next)
					return []int{option.Index}
				}
			}
		}
		return []int{0}
	default:
	}
	return request.DefaultSelection
}

func (p *stagedBandPolicy) commit() {
	if p.banded == nil {
		p.banded = make(map[id.ID]bool)
	}
	if len(p.current) >= 2 {
		for _, member := range p.current {
			p.banded[member] = true
		}
	}
}

func newStagedBandPolicy(t *testing.T, board attackingBandBoard, targets map[*game.Permanent]game.AttackTarget, addTo map[*game.Permanent][]*game.Permanent) *stagedBandPolicy {
	policy := &stagedBandPolicy{
		t: t, targets: make(map[id.ID]game.AttackTarget), addTo: make(map[id.ID][]id.ID),
		banding: map[id.ID]bool{board.hero.ObjectID: true, board.pegasus.ObjectID: true},
	}
	for creature, target := range targets {
		policy.targets[creature.ObjectID] = target
	}
	for anchor, members := range addTo {
		for _, member := range members {
			policy.addTo[anchor.ObjectID] = append(policy.addTo[anchor.ObjectID], member.ObjectID)
		}
	}
	return policy
}

// TestStagedAttackingBandChoicesOfferOnlyLegalMembers forms bands through the
// staged chooser. Each request offers exactly the declared creatures that can
// still join: the same attack target, not in another band, and no second
// creature without banding. A scripted choice forms two bands, the default
// forms none, and the logged action replays identically on a fresh copy.
func TestStagedAttackingBandChoicesOfferOnlyLegalMembers(t *testing.T) {
	player2 := game.AttackTarget{Player: game.Player2}
	player3 := game.AttackTarget{Player: game.Player3}
	for _, test := range []struct {
		name     string
		addTo    func(attackingBandBoard) map[*game.Permanent][]*game.Permanent
		want     func(attackingBandBoard) []game.AttackingBand
		requests int
	}{
		{
			name: "scripted bands",
			addTo: func(b attackingBandBoard) map[*game.Permanent][]*game.Permanent {
				return map[*game.Permanent][]*game.Permanent{b.hero: {b.bear, b.otherBear}, b.pegasus: {b.otherBear}}
			},
			want: func(b attackingBandBoard) []game.AttackingBand {
				return []game.AttackingBand{bandOf(b.hero, b.bear), bandOf(b.pegasus, b.otherBear)}
			},
			// Hero adds the bear, then finishes with only Pegasus offered; Pegasus
			// adds the other bear and has no candidate left, so no third ask.
			requests: 3,
		},
		{
			name:     "default declares no bands",
			addTo:    func(attackingBandBoard) map[*game.Permanent][]*game.Permanent { return nil },
			want:     func(attackingBandBoard) []game.AttackingBand { return nil },
			requests: 2,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			board := newAttackingBandBoard(t)
			third := addCombatCreaturePermanentWithPower(board.g, game.Player1, 2)
			third.SummoningSick = false
			before := board.g.Clone()
			policy := newStagedBandPolicy(t, board, map[*game.Permanent]game.AttackTarget{
				board.hero: player2, board.pegasus: player2, board.bear: player2,
				board.otherBear: player2, third: player3,
			}, test.addTo(board))
			agents := simPassPolicies()
			agents[game.Player1] = policy
			log := &TurnLog{}
			engine := NewEngine(nil)
			combatEngine{engine}.declareAttackers(board.g, agents, log)

			want := test.want(board)
			if !sameBands(board.g.Combat.Bands, want) || len(board.g.Combat.Attackers) != 5 {
				t.Fatalf("combat = %+v, want five attackers and bands %+v", board.g.Combat, want)
			}
			if want == nil && board.g.Combat.Bands != nil {
				t.Fatalf("default stored empty bands %#v, want nil", board.g.Combat.Bands)
			}
			if policy.requests != test.requests {
				t.Fatalf("asked %d band choices, want %d", policy.requests, test.requests)
			}
			logged := log.Actions[len(log.Actions)-1].Action
			payload, _ := logged.DeclareAttackersPayload()
			if payload.Staged || !sameBands(payload.Bands, want) {
				t.Fatalf("logged %+v, want the completed declaration with bands %+v", payload, want)
			}
			replayed, ok := engine.Simulator().Apply(before, game.Player1, logged, simPassPolicies())
			if !ok || !sameBands(replayed.Combat.Bands, want) ||
				!slices.Equal(replayed.Combat.Attackers, board.g.Combat.Attackers) {
				t.Fatalf("logged declaration did not replay: %+v", replayed)
			}
		})
	}
}

// TestStagedAttackWithoutBandingAsksNoBandChoice keeps declarations among
// creatures without banding exactly as before: no band request is asked.
func TestStagedAttackWithoutBandingAsksNoBandChoice(t *testing.T) {
	board := newAttackingBandBoard(t)
	player2 := game.AttackTarget{Player: game.Player2}
	policy := newStagedBandPolicy(t, board, map[*game.Permanent]game.AttackTarget{
		board.bear: player2, board.otherBear: player2,
	}, nil)
	agents := simPassPolicies()
	agents[game.Player1] = policy
	combatEngine{NewEngine(nil)}.declareAttackers(board.g, agents, &TurnLog{})
	if policy.requests != 0 || board.g.Combat.Bands != nil || len(board.g.Combat.Attackers) != 2 {
		t.Fatalf("requests %d, combat %+v; want two attackers and no band choice", policy.requests, board.g.Combat)
	}
}

// tricksterDefender flashes in Merfolk Trickster during the declare attackers
// step, targets the chosen attacker with its trigger, and records the bands and
// the attacker's banding seen at end of combat.
type tricksterDefender struct {
	simPassPolicy

	trickster id.ID
	target    id.ID
	cast      bool

	sawEndOfCombat bool
	endBands       []game.AttackingBand
	targetAttacks  bool
	targetBanding  bool
}

func (p *tricksterDefender) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step == game.StepDeclareAttackers && !p.cast {
		for _, candidate := range legal {
			if cast, ok := candidate.CastSpellPayload(); ok && cast.CardID == p.trickster {
				p.cast = true
				return candidate
			}
		}
	}
	if obs.Turn.Step == game.StepEndOfCombat && !p.sawEndOfCombat {
		p.sawEndOfCombat = true
		p.endBands = obs.Combat().Bands
		for _, attacker := range obs.Combat().Attackers {
			if attacker.Attacker.ObjectID == p.target {
				p.targetAttacks = true
				p.targetBanding = attacker.Attacker.HasKeyword(game.Banding)
			}
		}
	}
	return action.Pass()
}

func (p *tricksterDefender) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			if slices.ContainsFunc(option.Targets, func(target game.Target) bool { return target.PermanentID == p.target }) {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// TestAttackingBandPersistsAfterMemberLosesBanding has the defending player
// flash in registered Merfolk Trickster after the band is declared, targeting
// Benalish Hero so it loses all abilities. The band remains through end of
// combat (CR 702.22e) even though it now has no creature with banding.
func TestAttackingBandPersistsAfterMemberLosesBanding(t *testing.T) {
	board := newAttackingBandBoard(t)
	g := board.g
	island := registryCardDef(t, "Island")
	addCombatPermanent(g, game.Player2, island)
	addCombatPermanent(g, game.Player2, island)
	trickster := addCardToHand(g, game.Player2, registryCardDef(t, "Merfolk Trickster"))
	player2 := game.AttackTarget{Player: game.Player2}
	band := bandOf(board.hero, board.bear)
	engine := NewEngine(nil)
	declared, ok := engine.Simulator().Apply(g, game.Player1, action.DeclareAttackersWithBands(
		[]game.AttackDeclaration{attackAt(board.hero, player2), attackAt(board.bear, player2)},
		[]game.AttackingBand{band},
	), simPassPolicies())
	if !ok {
		t.Fatal("legal band declaration was rejected")
	}
	defender := &tricksterDefender{trickster: trickster, target: board.hero.ObjectID}
	agents := simPassPolicies()
	agents[game.Player2] = defender
	combatEngine{engine}.resolveCombatAfterAttackers(declared, agents, &TurnLog{})

	if !defender.cast {
		t.Fatal("defender never flashed in Merfolk Trickster")
	}
	if !defender.sawEndOfCombat || !defender.targetAttacks {
		t.Fatal("defender never saw Benalish Hero attacking at end of combat")
	}
	if defender.targetBanding {
		t.Fatal("Benalish Hero kept banding after Merfolk Trickster's trigger")
	}
	if !sameBands(defender.endBands, []game.AttackingBand{band}) {
		t.Fatalf("end-of-combat bands = %+v, want %+v", defender.endBands, band)
	}
}

// TestAttackingBandMemberRemovedFromCombat removes members of a three-creature
// band from combat. The band keeps its remaining two members, then stops being a
// band when only one is left (CR 702.22f); other bands are untouched.
func TestAttackingBandMemberRemovedFromCombat(t *testing.T) {
	board := newAttackingBandBoard(t)
	g := board.g
	third := bandingTestCreature(g, game.Player1, 1, 1, game.Banding)
	markAttacking(g, game.Player2, board.hero, board.pegasus, board.bear, board.otherBear, third)
	g.Combat.Bands = []game.AttackingBand{bandOf(board.hero, board.pegasus, board.bear), bandOf(third, board.otherBear)}

	removePermanentFromCombat(g, board.pegasus.ObjectID)
	if !sameBands(g.Combat.Bands, []game.AttackingBand{bandOf(board.hero, board.bear), bandOf(third, board.otherBear)}) {
		t.Fatalf("after removing Pegasus bands = %+v", g.Combat.Bands)
	}
	if view := NewObservation(g, game.Player2).Combat(); !sameBands(view.Bands, g.Combat.Bands) {
		t.Fatalf("observation %+v differs from %+v", view.Bands, g.Combat.Bands)
	}
	removePermanentFromCombat(g, board.bear.ObjectID)
	if !sameBands(g.Combat.Bands, []game.AttackingBand{bandOf(third, board.otherBear)}) {
		t.Fatalf("a one-creature band remained: %+v", g.Combat.Bands)
	}
	removePermanentFromCombat(g, third.ObjectID)
	if g.Combat.Bands != nil {
		t.Fatalf("bands = %#v, want none", g.Combat.Bands)
	}
	if len(g.Combat.Attackers) != 2 {
		t.Fatalf("attackers = %+v, want Benalish Hero and the other bear", g.Combat.Attackers)
	}
}

// TestAttackingBandMembersLeavingBattlefield declares two bands and then moves
// members off the battlefield by every shared departure route: destruction,
// sacrifice, exile and return to hand. A permanent that leaves the battlefield is
// removed from combat and from its band (CR 506.4, 702.22f), so each band
// shrinks and then stops being a band with one member left, while the other
// band is untouched.
func TestAttackingBandMembersLeavingBattlefield(t *testing.T) {
	board := newAttackingBandBoard(t)
	thirdBander := bandingTestCreature(board.g, game.Player1, 1, 1, game.Banding)
	fourthBander := bandingTestCreature(board.g, game.Player1, 1, 1, game.Banding)
	thirdBander.SummoningSick, fourthBander.SummoningSick = false, false
	player2 := game.AttackTarget{Player: game.Player2}
	declaration := action.DeclareAttackersWithBands([]game.AttackDeclaration{
		attackAt(board.hero, player2), attackAt(board.pegasus, player2), attackAt(thirdBander, player2),
		attackAt(board.bear, player2), attackAt(fourthBander, player2), attackAt(board.otherBear, player2),
	}, []game.AttackingBand{
		bandOf(board.hero, board.pegasus, thirdBander, board.bear),
		bandOf(fourthBander, board.otherBear),
	})
	g, ok := NewEngine(nil).Simulator().Apply(board.g, game.Player1, declaration, simPassPolicies())
	if !ok {
		t.Fatal("band declaration was rejected")
	}
	member := func(original *game.Permanent) *game.Permanent {
		t.Helper()
		permanent, ok := permanentByObjectID(g, original.ObjectID)
		if !ok {
			t.Fatalf("object %d is not on the battlefield", original.ObjectID)
		}
		return permanent
	}
	expect := func(step string, want ...game.AttackingBand) {
		t.Helper()
		if len(want) == 0 {
			want = nil
		}
		if !sameBands(g.Combat.Bands, want) {
			t.Fatalf("after %s bands = %+v, want %+v", step, g.Combat.Bands, want)
		}
		for _, seat := range []game.PlayerID{game.Player1, game.Player2} {
			if view := NewObservation(g, seat).Combat(); !sameBands(view.Bands, want) {
				t.Fatalf("after %s seat %d observed bands %+v, want %+v", step, seat, view.Bands, want)
			}
		}
	}

	if _, ok := destroyPermanent(g, board.pegasus.ObjectID); !ok {
		t.Fatal("Mesa Pegasus was not destroyed")
	}
	expect("Mesa Pegasus dies",
		bandOf(board.hero, thirdBander, board.bear), bandOf(fourthBander, board.otherBear))

	if !sacrificePermanent(g, member(thirdBander)) {
		t.Fatal("the banding creature was not sacrificed")
	}
	expect("a banding member is sacrificed",
		bandOf(board.hero, board.bear), bandOf(fourthBander, board.otherBear))

	if !movePermanentToZone(g, member(fourthBander), zone.Exile) {
		t.Fatal("the banding creature was not exiled")
	}
	expect("the second band's banding member is exiled", bandOf(board.hero, board.bear))

	if !movePermanentToZone(g, member(board.bear), zone.Hand) {
		t.Fatal("the bear was not returned to hand")
	}
	expect("the bear returns to hand")
	if g.Combat.Bands != nil {
		t.Fatalf("bands = %#v, want none", g.Combat.Bands)
	}
}
