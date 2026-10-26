package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt, SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), read
// 702.22a-m, 510.1a-e, 510.2, 702.19b/d and 702.2c. 702.22j: an attacker
// being blocked by a creature with banding has its combat damage divided by the
// defending player among any creatures blocking it (an exception to 510.1c);
// for a trampler that player may instead assign lethal damage to each blocker
// and the excess among the blockers and the attack target (702.19b), and with
// no such choice all of it stays on the blockers. 702.22k: a blocker blocking a
// creature with banding has its damage divided by the active player among the
// creatures it's blocking (an exception to 510.1d). Rulings retrieved
// 2026-10-26: Kjeldoran Escort 0fd7536a-5417-4de0-9a48-82a5f82d9af4, War
// Elephant 709cd715-28ef-42e8-bf8c-fc51a1ca0c6a, Mishra's War Machine
// 0f7b5921-3e10-47e7-98a7-8411a18313bf and Battering Ram
// e7e2857f-f6eb-4091-b758-7bb508544170 each include 2009-10-01 "If a creature
// in combat has banding, its controller assigns damage for creatures blocking
// or blocked by it." Thorn Elemental da901037-20e6-4445-8e7e-1ccd2e8b13ae
// 2018-04-27: "If blocked by a creature with banding, the defending player
// decides whether or not the damage is assigned 'as though it weren't
// blocked'." 510.1: attacking creatures' assignments are announced before
// blocking creatures'. Each division and the as-though-unblocked permission is
// asked of the deciding player as a choice, the protective default used only
// when no valid answer is given. 708.2/708.2a: a face-down permanent has no
// characteristics beyond those the face-down rules list (a nameless,
// colorless 2/2 creature), so decisions describe creatures only by public
// characteristics. Attacking bands (702.22c-i) and "bands with other" are
// separate capabilities not exercised here.

// bandingTestCreature returns a vanilla creature with the given power and
// toughness, optionally with the given keywords.
func bandingTestCreature(g *game.Game, controller game.PlayerID, power, toughness int, keywords ...game.Keyword) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Banding Test Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(keywords...),
		}},
	}})
}

func registryCombatPermanent(t *testing.T, g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	t.Helper()
	return addCombatPermanent(g, controller, registryCardDef(t, name))
}

func doubleBlockCombat(attacker *game.Permanent, blockers ...*game.Permanent) *game.CombatState {
	combat := &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	for _, blocker := range blockers {
		combat.Blockers = append(combat.Blockers, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
	}
	return combat
}

// TestBandingBlockerDefendingPlayerAssignsTramplerDamage runs registered
// Kjeldoran Escort (2/3 banding) blocking a 7/7 trampler alongside a 1/1. Under
// CR 702.22j the defending player chooses the trampler's damage assignment; by
// default all 7 goes to the 1/1, the Escort is unharmed and nothing tramples
// through. The same block with a vanilla 2/3 in the Escort's place lets
// the active player assign lethal to both blockers and trample 3 over.
func TestBandingBlockerDefendingPlayerAssignsTramplerDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 7, 7, game.Trample)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	chump := bandingTestCreature(g, game.Player2, 1, 1)
	g.Combat = doubleBlockCombat(attacker, escort, chump)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if escort.MarkedDamage != 0 || chump.MarkedDamage != 7 {
		t.Fatalf("blocker damage escort/chump = %d/%d, want 0/7 (defending player's division)", escort.MarkedDamage, chump.MarkedDamage)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("defending player life = %d, want 40 (defender's default keeps the damage on the blockers)", got)
	}

	control := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	controlAttacker := bandingTestCreature(control, game.Player1, 7, 7, game.Trample)
	vanilla := bandingTestCreature(control, game.Player2, 2, 3)
	controlChump := bandingTestCreature(control, game.Player2, 1, 1)
	control.Combat = doubleBlockCombat(controlAttacker, vanilla, controlChump)

	NewEngine(nil).resolveCombatDamage(control, &TurnLog{})

	if vanilla.MarkedDamage != 3 || controlChump.MarkedDamage != 1 {
		t.Fatalf("control blocker damage = %d/%d, want active player's lethal 3/1", vanilla.MarkedDamage, controlChump.MarkedDamage)
	}
	if got := control.Players[game.Player2].Life; got != 37 {
		t.Fatalf("control defending player life = %d, want 37 (3 trample)", got)
	}
}

// TestBandingBlockerSpreadsDamageBelowLethal proves the defending player may
// divide the damage in amounts the active player would not choose: a 4-power
// attacker blocked by two 3/3s, one with banding, deals 2 to each and neither
// dies, where without banding the active player kills the first blocker.
func TestBandingBlockerSpreadsDamageBelowLethal(t *testing.T) {
	for _, tc := range []struct {
		name          string
		banding       bool
		first, second int
	}{
		{name: "banding", banding: true, first: 2, second: 2},
		{name: "no banding", first: 3, second: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := bandingTestCreature(g, game.Player1, 4, 4)
			first := bandingTestCreature(g, game.Player2, 3, 3)
			var keywords []game.Keyword
			if tc.banding {
				keywords = append(keywords, game.Banding)
			}
			second := bandingTestCreature(g, game.Player2, 3, 3, keywords...)
			g.Combat = doubleBlockCombat(attacker, first, second)

			NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

			if first.MarkedDamage != tc.first || second.MarkedDamage != tc.second {
				t.Fatalf("blocker damage = %d/%d, want %d/%d", first.MarkedDamage, second.MarkedDamage, tc.first, tc.second)
			}
		})
	}
}

// TestBandingBlockerDeathtouchAttackerLosesOnlyOneBlocker: any nonzero damage
// from deathtouch is lethal (CR 702.2c), so the defending player assigns all of
// it to a single blocker rather than letting the active player spread one point
// to each.
func TestBandingBlockerDeathtouchAttackerLosesOnlyOneBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 3, 3, game.Deathtouch)
	first := bandingTestCreature(g, game.Player2, 3, 3, game.Banding)
	second := bandingTestCreature(g, game.Player2, 2, 3)
	g.Combat = doubleBlockCombat(attacker, first, second)

	engine := NewEngine(nil)
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActions(g)

	if second.MarkedDamage != 3 || first.MarkedDamage != 0 {
		t.Fatalf("blocker damage banding/other = %d/%d, want 0/3", first.MarkedDamage, second.MarkedDamage)
	}
	if !permanentWithCardOnBattlefield(g, first.CardInstanceID) || permanentWithCardOnBattlefield(g, second.CardInstanceID) {
		t.Fatalf("only the lower-power blocker should die to deathtouch")
	}
}

// TestBandingBlockerIndestructibleAbsorbsDamage: the defending player assigns
// all of the attacker's damage to an indestructible blocker.
func TestBandingBlockerIndestructibleAbsorbsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 6, 6, game.Trample)
	bander := bandingTestCreature(g, game.Player2, 1, 1, game.Banding)
	wall := bandingTestCreature(g, game.Player2, 0, 2, game.Indestructible)
	g.Combat = doubleBlockCombat(attacker, bander, wall)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if bander.MarkedDamage != 0 || wall.MarkedDamage != 6 {
		t.Fatalf("blocker damage bander/indestructible = %d/%d, want 0/6", bander.MarkedDamage, wall.MarkedDamage)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("defending player life = %d, want 40", got)
	}
}

// TestBandingBlockerDefendingPlayerDeclinesAsThoughUnblocked follows the Thorn
// Elemental ruling: with a banding blocker the defending player, not the
// attacker's controller, decides whether damage is assigned as though the
// attacker weren't blocked, and keeps it on the blockers.
func TestBandingBlockerDefendingPlayerDeclinesAsThoughUnblocked(t *testing.T) {
	newThorn := func(g *game.Game) *game.Permanent {
		return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:      "Thornlike Attacker",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{{
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectAssignCombatDamageAsThoughUnblocked,
					AffectedSource: true,
				}},
			}},
		}})
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	thorn := newThorn(g)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	g.Combat = doubleBlockCombat(thorn, escort)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if escort.MarkedDamage != 5 || g.Players[game.Player2].Life != 40 {
		t.Fatalf("escort damage %d, life %d; want 5 to the banding blocker and life 40", escort.MarkedDamage, g.Players[game.Player2].Life)
	}

	// The defending player, not the attacker's controller, is asked, and may
	// instead have the damage assigned as though the attacker weren't blocked.
	accepted := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	acceptedThorn := newThorn(accepted)
	acceptedEscort := registryCombatPermanent(t, accepted, game.Player2, "Kjeldoran Escort")
	accepted.Combat = doubleBlockCombat(acceptedThorn, acceptedEscort)
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandingDivisionAgent{answer: func(game.ChoiceRequest) []int { return []int{1} }}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1], agents[game.Player2] = attackerAgent, defender

	combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(accepted, normalCombatDamage, agents, &TurnLog{})

	if len(attackerAgent.requests) != 0 || len(defender.requests) != 1 ||
		defender.requests[0].Kind != game.ChoiceMay || defender.requests[0].CombatAttacker != acceptedThorn.ObjectID {
		t.Fatalf("as-though-unblocked requests attacker/defender = %+v/%+v, want one defender may choice", attackerAgent.requests, defender.requests)
	}
	if acceptedEscort.MarkedDamage != 0 || accepted.Players[game.Player2].Life != 35 {
		t.Fatalf("accepted escort damage %d, life %d; want 0 and 35", acceptedEscort.MarkedDamage, accepted.Players[game.Player2].Life)
	}

	control := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	controlThorn := newThorn(control)
	vanilla := bandingTestCreature(control, game.Player2, 2, 3)
	control.Combat = doubleBlockCombat(controlThorn, vanilla)

	NewEngine(nil).resolveCombatDamage(control, &TurnLog{})

	if vanilla.MarkedDamage != 0 || control.Players[game.Player2].Life != 35 {
		t.Fatalf("control blocker damage %d, life %d; want 0 and 35", vanilla.MarkedDamage, control.Players[game.Player2].Life)
	}
}

// TestBandingInvalidDivisionFallsBackToDefault: the defending player's division
// must assign all of the attacker's damage among its blockers (CR 702.22j,
// 510.1e). An answer leaving 2 of a 6-power trampler's damage unassigned (to
// trample over) is not a legal division, so the engine counts it invalid and
// uses the default, which puts all 6 on the lower-power banding 2/2.
func TestBandingInvalidDivisionFallsBackToDefault(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 6, 6, game.Trample)
	first := bandingTestCreature(g, game.Player2, 2, 2, game.Banding)
	second := bandingTestCreature(g, game.Player2, 4, 4)
	g.Combat = doubleBlockCombat(attacker, first, second)
	defender := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{first.CardInstanceID: 2, second.CardInstanceID: 2})
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player2] = defender

	combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, &TurnLog{})

	if len(defender.requests) != 1 {
		t.Fatalf("defender division requests = %d, want 1", len(defender.requests))
	}
	if first.MarkedDamage != 6 || second.MarkedDamage != 0 || g.Players[game.Player2].Life != 40 {
		t.Fatalf("invalid answer fallback = %d/%d life %d, want 6/0 life 40", first.MarkedDamage, second.MarkedDamage, g.Players[game.Player2].Life)
	}
	if g.Execution.InvalidChoices != 1 || g.Execution.ChoiceFallbacks != 1 {
		t.Fatalf("invalid/fallback counts = %d/%d, want 1/1", g.Execution.InvalidChoices, g.Execution.ChoiceFallbacks)
	}
}

// TestGrantedAndRemovedBandingControlDamageDivision proves the rule reads the
// blocker's current abilities when damage is assigned: banding granted by a
// continuous effect transfers the division, and banding removed by one (as
// Tolaria does) leaves the active player dividing the damage.
func TestGrantedAndRemovedBandingControlDamageDivision(t *testing.T) {
	addAbilityEffect := func(g *game.Game, effect game.ContinuousEffect) {
		effect.Layer = game.LayerAbility
		effect.Group = game.ObjectControlledGroup(
			game.SourcePermanentReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
		)
		addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name:            "Banding Ability Effect",
			Types:           []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{effect}}},
		}})
	}

	granted := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(granted, game.Player1, 4, 4, game.Trample)
	first := bandingTestCreature(granted, game.Player2, 3, 3)
	second := bandingTestCreature(granted, game.Player2, 3, 3)
	addAbilityEffect(granted, game.ContinuousEffect{AddKeywords: []game.Keyword{game.Banding}})
	granted.Combat = doubleBlockCombat(attacker, first, second)
	NewEngine(nil).resolveCombatDamage(granted, &TurnLog{})
	if first.MarkedDamage != 2 || second.MarkedDamage != 2 {
		t.Fatalf("granted banding division = %d/%d, want 2/2", first.MarkedDamage, second.MarkedDamage)
	}

	removed := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker = bandingTestCreature(removed, game.Player1, 4, 4, game.Trample)
	escort := registryCombatPermanent(t, removed, game.Player2, "Kjeldoran Escort")
	second = bandingTestCreature(removed, game.Player2, 1, 1)
	addAbilityEffect(removed, game.ContinuousEffect{RemoveKeywords: []game.Keyword{game.Banding}})
	removed.Combat = doubleBlockCombat(attacker, escort, second)
	NewEngine(nil).resolveCombatDamage(removed, &TurnLog{})
	if escort.MarkedDamage != 3 || second.MarkedDamage != 1 {
		t.Fatalf("removed banding division = %d/%d, want active player's lethal 3/1", escort.MarkedDamage, second.MarkedDamage)
	}
}

// TestBlockerOfBandingAttackerDamageDividedByActivePlayer covers CR 702.22k
// with registered War Elephant: a 3/3 blocking both War Elephant (2/2) and a
// 4/4 has its damage divided by the active player, who puts all 3 on the 4/4
// so neither attacker dies. Without banding, the defending player's
// deterministic division kills the first attacker.
func TestBlockerOfBandingAttackerDamageDividedByActivePlayer(t *testing.T) {
	run := func(t *testing.T, elephantName string) (*game.Permanent, *game.Permanent) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		var elephant *game.Permanent
		if elephantName != "" {
			elephant = registryCombatPermanent(t, g, game.Player1, elephantName)
		} else {
			elephant = bandingTestCreature(g, game.Player1, 2, 2, game.Trample)
		}
		big := bandingTestCreature(g, game.Player1, 4, 4)
		blocker := bandingTestCreature(g, game.Player2, 3, 3)
		g.Combat = &game.CombatState{
			Attackers: []game.AttackDeclaration{
				{Attacker: elephant.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				{Attacker: big.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			},
			Blockers: []game.BlockDeclaration{
				{Blocker: blocker.ObjectID, Blocking: elephant.ObjectID},
				{Blocker: blocker.ObjectID, Blocking: big.ObjectID},
			},
		}
		NewEngine(nil).resolveCombatDamage(g, &TurnLog{})
		if blocker.MarkedDamage != 6 {
			t.Fatalf("blocker damage = %d, want 6 (2 from the elephant, 4 from the 4/4)", blocker.MarkedDamage)
		}
		return elephant, big
	}

	elephant, big := run(t, "War Elephant")
	if elephant.MarkedDamage != 0 || big.MarkedDamage != 3 {
		t.Fatalf("banding attacker damage elephant/4-4 = %d/%d, want active player's 0/3", elephant.MarkedDamage, big.MarkedDamage)
	}
	elephant, big = run(t, "")
	if elephant.MarkedDamage != 2 || big.MarkedDamage != 1 {
		t.Fatalf("control attacker damage = %d/%d, want defending player's lethal 2/1", elephant.MarkedDamage, big.MarkedDamage)
	}
}

// TestBandingDivisionAccountsForFirstStrikeDamage: in the regular damage step
// the defending player's division counts damage already marked in the
// first-strike step (CR 510.4, 702.22j). A double striker's 3 first-strike
// damage spreads 2/1 onto two 3/3s; its regular 3 must then kill one of them,
// so all of it goes to a single blocker and the other survives.
func TestBandingDivisionAccountsForFirstStrikeDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 3, 10, game.DoubleStrike)
	first := bandingTestCreature(g, game.Player2, 3, 3, game.Banding)
	second := bandingTestCreature(g, game.Player2, 2, 3)
	g.Combat = doubleBlockCombat(attacker, first, second)
	engine := NewEngine(nil)

	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	if first.MarkedDamage != 2 || second.MarkedDamage != 1 {
		t.Fatalf("first-strike division = %d/%d, want 2/1", first.MarkedDamage, second.MarkedDamage)
	}
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	if first.MarkedDamage != 2 || second.MarkedDamage != 4 {
		t.Fatalf("regular division = %d/%d, want 2/4 (only the lower-power blocker dealt lethal)", first.MarkedDamage, second.MarkedDamage)
	}
}

// bandingDivisionAgent records the combat damage decisions banding routes to its
// player (divisions, the as-though-unblocked choice and the trample excess) and
// answers them with
// answer, or the default. It declares blocks like simultaneousCombatPolicy.
type bandingDivisionAgent struct {
	simultaneousCombatPolicy
	requests     []game.ChoiceRequest
	observations []PlayerObservation
	answer       func(game.ChoiceRequest) []int
}

func (a *bandingDivisionAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceCombatDeclaration || request.CombatAttacker == 0 && request.CombatBlocker == 0 {
		return request.DefaultSelection
	}
	a.requests = append(a.requests, request)
	a.observations = append(a.observations, obs)
	if a.answer != nil {
		return a.answer(request)
	}
	return request.DefaultSelection
}

// divisionSelection encodes a division, keyed by recipient card instance, as a
// ChoiceCombatDamageDivision selection.
func divisionSelection(request game.ChoiceRequest, amounts map[id.ID]int) []int {
	var selected []int
	for _, option := range request.Options {
		for range amounts[option.Card.Val.CardID] {
			selected = append(selected, option.Index)
		}
	}
	return selected
}

// TestBandingDefenderChoosesAttackerDamageInLiveCombat plays a real combat
// through the Simulator: registered Kjeldoran Escort and a 3/3 block a 4/4.
// The defending player, not the active player, is asked to divide the 4/4's
// damage among both blockers (CR 702.22j) and that answer is honored, here all
// 4 on the Escort, a division neither the default nor the active player would
// make.
func TestBandingDefenderChoosesAttackerDamageInLiveCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 4, 4)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	other := bandingTestCreature(g, game.Player2, 3, 3)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{escort.CardInstanceID: 4})
	}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}

	if len(attackerAgent.requests) != 0 {
		t.Fatalf("active player was asked %+v; the defending player divides the damage", attackerAgent.requests)
	}
	if len(defender.requests) != 1 {
		t.Fatalf("defender division requests = %d, want 1", len(defender.requests))
	}
	request := defender.requests[0]
	if request.Kind != game.ChoiceCombatDamageDivision || request.Player != game.Player2 ||
		request.CombatAttacker != attacker.ObjectID || request.CombatBlocker != 0 ||
		request.MinChoices != 4 || request.MaxChoices != 4 || len(request.Options) != 2 {
		t.Fatalf("division request = %+v, want Player2 dividing the attacker's 4 among 2 blockers", request)
	}
	if obs := defender.observations[0]; obs.Player != game.Player2 || obs.Turn.Step != game.StepCombatDamage {
		t.Fatalf("decision observation player/step = %v/%v, want Player2 in the combat damage step", obs.Player, obs.Turn.Step)
	}
	afterOther, ok := permanentByObjectID(after, other.ObjectID)
	if !ok || afterOther.MarkedDamage != 0 {
		t.Fatalf("other blocker = %+v, want unharmed", afterOther)
	}
	if permanentWithCardOnBattlefield(after, escort.CardInstanceID) {
		t.Fatal("Kjeldoran Escort should die to the 4 damage its controller assigned to it")
	}
}

// TestBandingActivePlayerChoosesBlockerDamageInLiveCombat plays a real combat
// through the Simulator: registered War Elephant and a 4/4 attack, and a 3/3
// that can block an additional creature blocks both. The active player, not
// the defending player, is asked to divide the 3/3's damage (CR 702.22k) and
// that answer is honored, here all 3 on War Elephant, unlike the default.
func TestBandingActivePlayerChoosesBlockerDamageInLiveCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
	big := bandingTestCreature(g, game.Player1, 4, 4)
	blocker := bandingTestCreature(g, game.Player2, 3, 3)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: blocker.ObjectID, AffectedSource: true, AdditionalBlockCount: 1})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackerAgent := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{elephant.CardInstanceID: 3})
	}}
	defender := &bandingDivisionAgent{}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		{Attacker: elephant.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: big.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}

	if len(defender.requests) != 0 {
		t.Fatalf("defending player was asked %+v; the active player divides the damage", defender.requests)
	}
	if len(attackerAgent.requests) != 1 {
		t.Fatalf("active player division requests = %d, want 1", len(attackerAgent.requests))
	}
	request := attackerAgent.requests[0]
	if request.Kind != game.ChoiceCombatDamageDivision || request.Player != game.Player1 ||
		request.CombatBlocker != blocker.ObjectID || request.CombatAttacker != 0 ||
		request.MinChoices != 3 || request.MaxChoices != 3 || len(request.Options) != 2 {
		t.Fatalf("division request = %+v, want Player1 dividing the blocker's 3 among 2 attackers", request)
	}
	afterBig, ok := permanentByObjectID(after, big.ObjectID)
	if !ok || afterBig.MarkedDamage != 0 {
		t.Fatalf("4/4 attacker = %+v, want unharmed", afterBig)
	}
	if permanentWithCardOnBattlefield(after, elephant.CardInstanceID) {
		t.Fatal("War Elephant should die to the 3 damage its controller assigned to it")
	}
}

// TestBandingDivisionsAskedInAnnouncementOrder: in one damage pass with both
// banding rules applying, each division goes to its player and attacking
// creatures' assignments are announced before blocking creatures' (CR 510.1).
// The defending player divides the 4/4's damage among the Escort, a 1/1 and a
// 3/3 (CR 702.22j); then the active player divides that 3/3's damage between
// War Elephant and the 4/4 it also blocks (CR 702.22k).
func TestBandingDivisionsAskedInAnnouncementOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
	big := bandingTestCreature(g, game.Player1, 4, 4)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	chump := bandingTestCreature(g, game.Player2, 1, 1)
	wide := bandingTestCreature(g, game.Player2, 3, 3)
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: elephant.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: big.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: wide.ObjectID, Blocking: elephant.ObjectID},
			{Blocker: escort.ObjectID, Blocking: big.ObjectID},
			{Blocker: chump.ObjectID, Blocking: big.ObjectID},
			{Blocker: wide.ObjectID, Blocking: big.ObjectID},
		},
	}
	attackerAgent := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{big.CardInstanceID: 3})
	}}
	defender := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		return divisionSelection(request, map[id.ID]int{chump.CardInstanceID: 1, wide.CardInstanceID: 3})
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1], agents[game.Player2] = attackerAgent, defender
	log := &TurnLog{}

	combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, log)

	if len(log.Choices) != 2 {
		t.Fatalf("logged choices = %+v, want the two divisions", log.Choices)
	}
	first, second := log.Choices[0], log.Choices[1]
	if first.Request.Player != game.Player2 || first.Request.CombatAttacker != big.ObjectID || first.UsedFallback ||
		second.Request.Player != game.Player1 || second.Request.CombatBlocker != wide.ObjectID || second.UsedFallback {
		t.Fatalf("decision order = %+v then %+v, want defender's attacker division then active player's blocker division", first.Request, second.Request)
	}
	if escort.MarkedDamage != 0 || chump.MarkedDamage != 1 || wide.MarkedDamage != 5 {
		t.Fatalf("blocker damage escort/1-1/3-3 = %d/%d/%d, want 0/1/5 (2 from the elephant)", escort.MarkedDamage, chump.MarkedDamage, wide.MarkedDamage)
	}
	if elephant.MarkedDamage != 0 || big.MarkedDamage != 6 {
		t.Fatalf("attacker damage elephant/4-4 = %d/%d, want 0/6 (3 from the 3/3, 2 from the Escort, 1 from the 1/1)", elephant.MarkedDamage, big.MarkedDamage)
	}
}

// trampleExcessAnswer answers the trample-excess number with excess and each
// division with amounts, keyed by recipient card instance.
func trampleExcessAnswer(excess int, amounts map[id.ID]int) func(game.ChoiceRequest) []int {
	return func(request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceNumber {
			return []int{excess}
		}
		return divisionSelection(request, amounts)
	}
}

// TestBandingDefenderAssignsTrampleExcess covers CR 702.22j with 702.19b: the
// defending player who chooses a trampler's damage assignment may keep all of
// it on the blockers, or assign lethal damage to each blocker and the excess
// among the blockers and the attack target, in any amounts.
func TestBandingDefenderAssignsTrampleExcess(t *testing.T) {
	for _, tc := range []struct {
		name         string
		power        int
		deathtouch   bool
		withChump    bool
		answer       func(escort, chump *game.Permanent) func(game.ChoiceRequest) []int
		maxExcess    int
		divisions    []int
		escortDamage int
		chumpDamage  int
		life         int
	}{
		{
			name: "default keeps all damage on the Escort", power: 7, maxExcess: 4,
			escortDamage: 7, life: 40,
		},
		{
			name: "lethal to the Escort and 4 to the player", power: 7, maxExcess: 4,
			answer:       func(_, _ *game.Permanent) func(game.ChoiceRequest) []int { return trampleExcessAnswer(4, nil) },
			escortDamage: 3, life: 36,
		},
		{
			name: "excess split between the Escort and the player", power: 7, maxExcess: 4,
			answer:       func(_, _ *game.Permanent) func(game.ChoiceRequest) []int { return trampleExcessAnswer(2, nil) },
			escortDamage: 5, life: 38,
		},
		{
			name: "lethal to both blockers, the rest divided and 2 to the player", power: 7, withChump: true, maxExcess: 3,
			answer: func(escort, _ *game.Permanent) func(game.ChoiceRequest) []int {
				return trampleExcessAnswer(2, map[id.ID]int{escort.CardInstanceID: 1})
			},
			divisions:    []int{1},
			escortDamage: 4, chumpDamage: 1, life: 38,
		},
		{
			name: "no trample excess divides freely among both blockers", power: 7, withChump: true, maxExcess: 3,
			answer: func(escort, chump *game.Permanent) func(game.ChoiceRequest) []int {
				return trampleExcessAnswer(0, map[id.ID]int{escort.CardInstanceID: 2, chump.CardInstanceID: 5})
			},
			divisions:    []int{7},
			escortDamage: 2, chumpDamage: 5, life: 40,
		},
		{
			name: "deathtouch makes one damage lethal to each blocker", power: 5, deathtouch: true, withChump: true, maxExcess: 3,
			answer:       func(_, _ *game.Permanent) func(game.ChoiceRequest) []int { return trampleExcessAnswer(3, nil) },
			escortDamage: 1, chumpDamage: 1, life: 37,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			keywords := []game.Keyword{game.Trample}
			if tc.deathtouch {
				keywords = append(keywords, game.Deathtouch)
			}
			attacker := bandingTestCreature(g, game.Player1, tc.power, tc.power, keywords...)
			escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
			blockers := []*game.Permanent{escort}
			var chump *game.Permanent
			if tc.withChump {
				chump = bandingTestCreature(g, game.Player2, 1, 1)
				blockers = append(blockers, chump)
			}
			g.Turn.ActivePlayer = game.Player1
			g.Combat = doubleBlockCombat(attacker, blockers...)
			defender := &bandingDivisionAgent{}
			if tc.answer != nil {
				defender.answer = tc.answer(escort, chump)
			}
			attackerAgent := &bandingDivisionAgent{}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1], agents[game.Player2] = attackerAgent, defender

			combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, &TurnLog{})

			if len(attackerAgent.requests) != 0 {
				t.Fatalf("active player was asked %+v; the defending player assigns the damage", attackerAgent.requests)
			}
			if len(defender.requests) != 1+len(tc.divisions) {
				t.Fatalf("defender requests = %+v, want the trample excess then %d division(s)", defender.requests, len(tc.divisions))
			}
			number := defender.requests[0]
			if number.Kind != game.ChoiceNumber || number.Player != game.Player2 || number.CombatAttacker != attacker.ObjectID ||
				!number.NumberRange.Exists || number.NumberRange.Val.Minimum != 0 || number.NumberRange.Val.Maximum != opt.Val(tc.maxExcess) {
				t.Fatalf("trample excess request = %+v, want Player2 choosing 0..%d", number, tc.maxExcess)
			}
			for i, total := range tc.divisions {
				division := defender.requests[1+i]
				if division.Kind != game.ChoiceCombatDamageDivision || division.MinChoices != total || division.CombatAttacker != attacker.ObjectID {
					t.Fatalf("division request %d = %+v, want %d to divide", i, division, total)
				}
			}
			if escort.MarkedDamage != tc.escortDamage {
				t.Fatalf("Escort damage = %d, want %d", escort.MarkedDamage, tc.escortDamage)
			}
			if chump != nil && chump.MarkedDamage != tc.chumpDamage {
				t.Fatalf("1/1 damage = %d, want %d", chump.MarkedDamage, tc.chumpDamage)
			}
			if got := g.Players[game.Player2].Life; got != tc.life {
				t.Fatalf("defending player life = %d, want %d", got, tc.life)
			}
			if g.Execution.InvalidChoices != 0 {
				t.Fatalf("invalid choices = %d, want 0", g.Execution.InvalidChoices)
			}
		})
	}
}

// TestBandingTrampleExcessAboveMaximumFallsBack: assigning more to the player
// than the damage left after lethal to each blocker is illegal (CR 702.19b), so
// the engine counts the answer invalid and keeps the damage on the blockers.
func TestBandingTrampleExcessAboveMaximumFallsBack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 7, 7, game.Trample)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	g.Turn.ActivePlayer = game.Player1
	g.Combat = doubleBlockCombat(attacker, escort)
	defender := &bandingDivisionAgent{answer: trampleExcessAnswer(5, nil)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player2] = defender

	combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, &TurnLog{})

	if escort.MarkedDamage != 7 || g.Players[game.Player2].Life != 40 {
		t.Fatalf("Escort damage %d, life %d; want the default 7 and 40", escort.MarkedDamage, g.Players[game.Player2].Life)
	}
	if g.Execution.InvalidChoices != 1 || g.Execution.ChoiceFallbacks != 1 {
		t.Fatalf("invalid/fallback counts = %d/%d, want 1/1", g.Execution.InvalidChoices, g.Execution.ChoiceFallbacks)
	}
}

// TestBandingDefenderAssignsTrampleExcessInLiveCombat plays a real combat
// through the Simulator: registered Kjeldoran Escort and a 1/1 block a 7/7
// trampler, and the defending player assigns lethal damage to each and the
// remaining 3 to themself.
func TestBandingDefenderAssignsTrampleExcessInLiveCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := bandingTestCreature(g, game.Player1, 7, 7, game.Trample)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	chump := bandingTestCreature(g, game.Player2, 1, 1)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackerAgent := &bandingDivisionAgent{}
	defender := &bandingDivisionAgent{answer: trampleExcessAnswer(3, nil)}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = attackerAgent, defender
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}

	if len(attackerAgent.requests) != 0 || len(defender.requests) != 1 || defender.requests[0].Kind != game.ChoiceNumber {
		t.Fatalf("requests attacker/defender = %+v/%+v, want only the defender's trample excess", attackerAgent.requests, defender.requests)
	}
	if got := after.Players[game.Player2].Life; got != 37 {
		t.Fatalf("defending player life = %d, want 37", got)
	}
	if permanentWithCardOnBattlefield(after, escort.CardInstanceID) || permanentWithCardOnBattlefield(after, chump.CardInstanceID) {
		t.Fatal("both blockers were assigned lethal damage and should die")
	}
}

// TestBandingDecisionsHideFaceDownCharacteristics: a face-down creature has no
// name and is a colorless 2/2 (CR 708.2a), so the damage decisions banding
// routes to a player, and the logged requests, describe a face-down attacker
// and a face-down blocker only by those public characteristics.
func TestBandingDecisionsHideFaceDownCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := registryCombatPermanent(t, g, game.Player1, "Exalted Angel")
	attacker.FaceDown, attacker.FaceDownKind = true, game.FaceDownMorph
	attacker.Counters.Add(counter.PlusOnePlusOne, 2)
	escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	hidden := registryCombatPermanent(t, g, game.Player2, "Exalted Angel")
	hidden.FaceDown, hidden.FaceDownKind = true, game.FaceDownMorph
	g.Turn.ActivePlayer = game.Player1
	g.Combat = doubleBlockCombat(attacker, escort, hidden)
	defender := &bandingDivisionAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player2] = defender
	log := &TurnLog{}

	combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, log)

	if len(defender.requests) != 1 || len(log.Choices) != 1 {
		t.Fatalf("requests/logged = %d/%d, want one division", len(defender.requests), len(log.Choices))
	}
	for _, request := range []game.ChoiceRequest{defender.requests[0], log.Choices[0].Request} {
		if request.Kind != game.ChoiceCombatDamageDivision || request.MinChoices != 4 {
			t.Fatalf("request = %+v, want a division of the face-down 4/4's damage", request)
		}
		texts := []string{request.Prompt}
		infos := []opt.V[game.ChoiceCardInfo]{request.Subject}
		for _, option := range request.Options {
			texts = append(texts, option.Label)
			infos = append(infos, option.Card)
		}
		for _, text := range texts {
			if strings.Contains(text, "Exalted Angel") {
				t.Fatalf("decision text %q reveals a face-down card", text)
			}
		}
		for i, info := range infos {
			if !info.Exists {
				t.Fatalf("decision card info %d missing", i)
			}
			if info.Val.CardID == escort.CardInstanceID {
				if info.Val.Name != "Kjeldoran Escort" {
					t.Fatalf("face-up Escort info = %+v", info.Val)
				}
				continue
			}
			if info.Val.Name != "" || len(info.Val.Colors) != 0 || info.Val.ManaValue != 0 || !slices.Equal(info.Val.Types, []types.Card{types.Creature}) {
				t.Fatalf("face-down info %d = %+v, want a nameless colorless creature with mana value 0", i, info.Val)
			}
		}
	}
}
