package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: CR 2026-09-25 (verified official copy) 615.12 says
// unpreventable damage still has prevention effects applied, but they prevent
// nothing, their additional effects still happen, and prevention shields are not
// reduced; 122.1c makes a shield counter's removal such an additional effect.
// Scryfall WotC rulings for Excruciator (2005-10-01), Leyline of Punishment and
// Combust (2010-08-15), and Skullcrack (2024-01-12) confirm protection doesn't
// prevent it, shields aren't used up, and effects that replace or redirect
// damage without the word "prevent" still apply to it.

// unpreventableRouteGame holds a game with a global "prevent all damage" shield
// and three high-toughness creatures for Player2.
type unpreventableRouteGame struct {
	g        *game.Game
	engine   *Engine
	first    *game.Permanent
	second   *game.Permanent
	small    *game.Permanent
	lifeLost func(game.PlayerID) int
}

func newUnpreventableRouteGame() *unpreventableRouteGame {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	r := &unpreventableRouteGame{
		g:      g,
		engine: engine,
		first:  addCombatCreaturePermanentWithPower(g, game.Player2, 10),
		second: addCombatCreaturePermanentWithPower(g, game.Player2, 10),
		small:  addCombatCreaturePermanentWithPower(g, game.Player2, 2),
	}
	var life [game.NumPlayers]int
	for i := range g.Players {
		life[i] = g.Players[i].Life
	}
	r.lifeLost = func(player game.PlayerID) int { return life[player] - g.Players[player].Life }
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{
		All:    true,
		Global: true,
	}, nil)
	return r
}

func (r *unpreventableRouteGame) marked() [3]int {
	return [3]int{r.first.MarkedDamage, r.second.MarkedDamage, r.small.MarkedDamage}
}

// TestDamageCantBePreventedEveryDamageRoute proves the instruction-level
// "the damage can't be prevented" rider reaches every Damage recipient route and
// that the same instruction without the rider is fully prevented.
func TestDamageCantBePreventedEveryDamageRoute(t *testing.T) {
	type outcome struct {
		marked   [3]int
		lifeLost [game.NumPlayers]int
	}
	tests := []struct {
		name    string
		damage  game.Damage
		targets func(*unpreventableRouteGame) []game.Target
		choices [][]int
		want    func(*unpreventableRouteGame) outcome
	}{
		{
			name:    "any target player",
			damage:  game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)},
			targets: func(*unpreventableRouteGame) []game.Target { return []game.Target{game.PlayerTarget(game.Player2)} },
			want: func(*unpreventableRouteGame) outcome {
				return outcome{lifeLost: [game.NumPlayers]int{game.Player2: 3}}
			},
		},
		{
			name:   "any target permanent",
			damage: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)},
			targets: func(r *unpreventableRouteGame) []game.Target {
				return []game.Target{game.PermanentTarget(r.first.ObjectID)}
			},
			want: func(*unpreventableRouteGame) outcome { return outcome{marked: [3]int{3, 0, 0}} },
		},
		{
			name:    "referenced player",
			damage:  game.Damage{Amount: game.Fixed(3), Recipient: game.PlayerDamageRecipient(game.TargetPlayerReference(0))},
			targets: func(*unpreventableRouteGame) []game.Target { return []game.Target{game.PlayerTarget(game.Player2)} },
			want: func(*unpreventableRouteGame) outcome {
				return outcome{lifeLost: [game.NumPlayers]int{game.Player2: 3}}
			},
		},
		{
			name:   "referenced permanent",
			damage: game.Damage{Amount: game.Fixed(3), Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0))},
			targets: func(r *unpreventableRouteGame) []game.Target {
				return []game.Target{game.PermanentTarget(r.second.ObjectID)}
			},
			want: func(*unpreventableRouteGame) outcome { return outcome{marked: [3]int{0, 3, 0}} },
		},
		{
			name:   "divided",
			damage: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0), Divided: true},
			targets: func(r *unpreventableRouteGame) []game.Target {
				return []game.Target{game.PermanentTarget(r.first.ObjectID), game.PermanentTarget(r.second.ObjectID)}
			},
			choices: [][]int{{0, 1, 1}},
			want:    func(*unpreventableRouteGame) outcome { return outcome{marked: [3]int{1, 2, 0}} },
		},
		{
			name:   "each target",
			damage: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0), EachTarget: true},
			targets: func(r *unpreventableRouteGame) []game.Target {
				return []game.Target{game.PermanentTarget(r.first.ObjectID), game.PlayerTarget(game.Player2)}
			},
			want: func(*unpreventableRouteGame) outcome {
				return outcome{marked: [3]int{3, 0, 0}, lifeLost: [game.NumPlayers]int{game.Player2: 3}}
			},
		},
		{
			name: "excess to controller",
			damage: game.Damage{
				Amount:          game.Fixed(5),
				Recipient:       game.AnyTargetDamageRecipient(0),
				ExcessRecipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
			},
			targets: func(r *unpreventableRouteGame) []game.Target {
				return []game.Target{game.PermanentTarget(r.small.ObjectID)}
			},
			want: func(*unpreventableRouteGame) outcome {
				return outcome{marked: [3]int{0, 0, 2}, lifeLost: [game.NumPlayers]int{game.Player2: 3}}
			},
		},
		{
			name: "permanent group",
			damage: game.Damage{
				Amount: game.Fixed(1),
				Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				})),
			},
			want: func(*unpreventableRouteGame) outcome { return outcome{marked: [3]int{1, 1, 1}} },
		},
		{
			name:   "player group",
			damage: game.Damage{Amount: game.Fixed(2), Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference())},
			want: func(*unpreventableRouteGame) outcome {
				var lost [game.NumPlayers]int
				for player := range game.NumPlayers {
					if game.PlayerID(player) != game.Player1 {
						lost[player] = 2
					}
				}
				return outcome{lifeLost: lost}
			},
		},
	}
	for _, test := range tests {
		for _, unpreventable := range []bool{true, false} {
			name := test.name + "/preventable"
			if unpreventable {
				name = test.name + "/unpreventable"
			}
			t.Run(name, func(t *testing.T) {
				r := newUnpreventableRouteGame()
				damage := test.damage
				damage.CantBePrevented = unpreventable
				var targets []game.Target
				if test.targets != nil {
					targets = test.targets(r)
				}
				addEffectSpellToStack(r.g, game.Player1, damage, targets)
				agents := [game.NumPlayers]PlayerAgent{}
				if test.choices != nil {
					agents[game.Player1] = &choiceOnlyAgent{choices: test.choices}
				}
				shields := len(r.g.PreventionShields)
				r.engine.resolveTopOfStackWithChoices(r.g, agents, &TurnLog{})

				want := outcome{}
				if unpreventable {
					want = test.want(r)
				}
				got := outcome{marked: r.marked()}
				for player := range game.NumPlayers {
					got.lifeLost[player] = r.lifeLost(game.PlayerID(player))
				}
				if got != want {
					t.Fatalf("outcome = %+v, want %+v", got, want)
				}
				if len(r.g.PreventionShields) != shields {
					t.Fatalf("prevention shields = %d, want %d unchanged", len(r.g.PreventionShields), shields)
				}
			})
		}
	}
}

// TestDamageCantBePreventedLeavesFiniteShieldUnused proves a "prevent the next
// N damage" shield prevents nothing from unpreventable damage, isn't reduced,
// and still prevents later preventable damage (CR 615.12).
func TestDamageCantBePreventedLeavesFiniteShieldUnused(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player2, Targets: targets}, game.PreventDamage{
		Amount:    game.Fixed(2),
		AnyTarget: game.AnyTargetDamageRecipient(0),
	}, nil)

	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:          game.Fixed(3),
		Recipient:       game.AnyTargetDamageRecipient(0),
		CantBePrevented: true,
	}, targets)
	engine.resolveTopOfStack(g, &TurnLog{})
	if target.MarkedDamage != 3 {
		t.Fatalf("unpreventable marked damage = %d, want 3", target.MarkedDamage)
	}
	if len(g.PreventionShields) != 1 {
		t.Fatalf("prevention shields = %+v, want the 2-damage shield intact", g.PreventionShields)
	}

	addEffectSpellToStack(g, game.Player1, game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}, targets)
	engine.resolveTopOfStack(g, &TurnLog{})
	if target.MarkedDamage != 4 {
		t.Fatalf("marked damage after preventable 3 = %d, want 4 (shield prevents 2)", target.MarkedDamage)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want consumed by preventable damage", g.PreventionShields)
	}
}

// TestDamageCantBePreventedBypassesProtectionAndStaticPrevention proves
// protection's damage prevention and a static prevention replacement prevent
// nothing from unpreventable damage, while preventable damage is still stopped.
func TestDamageCantBePreventedBypassesProtectionAndStaticPrevention(t *testing.T) {
	for _, unpreventable := range []bool{true, false} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		protected := addProtectionFromTypesPermanent(g, game.Player2, types.Sorcery)
		addReplacementPermanent(t, g, game.Player2, damagePreventionCardDef(&game.DamagePreventionSpec{Amount: 20}))
		life := g.Players[game.Player2].Life

		addEffectSpellToStack(g, game.Player1, game.Damage{
			Amount:          game.Fixed(3),
			Recipient:       game.AnyTargetDamageRecipient(0),
			EachTarget:      true,
			CantBePrevented: unpreventable,
		}, []game.Target{game.PermanentTarget(protected.ObjectID), game.PlayerTarget(game.Player2)})
		engine.resolveTopOfStack(g, &TurnLog{})

		want := 0
		if unpreventable {
			want = 3
		}
		if protected.MarkedDamage != want {
			t.Fatalf("unpreventable=%v protected creature damage = %d, want %d", unpreventable, protected.MarkedDamage, want)
		}
		if lost := life - g.Players[game.Player2].Life; lost != want {
			t.Fatalf("unpreventable=%v player damage through static prevention = %d, want %d", unpreventable, lost, want)
		}
	}
}

// TestDamageCantBePreventedStillRemovesShieldCounter proves a shield counter is
// removed by unpreventable damage, which is still dealt (CR 122.1c, 615.12).
func TestDamageCantBePreventedStillRemovesShieldCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	target.Counters.Add(counter.Shield, 1)

	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:          game.Fixed(3),
		Recipient:       game.AnyTargetDamageRecipient(0),
		CantBePrevented: true,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if target.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("shield counters = %d, want removed", target.Counters.Get(counter.Shield))
	}
	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
}

// TestDamageCantBePreventedSurvivesRedirection proves redirected unpreventable
// damage remains unpreventable at its new recipient (Excruciator ruling: it may
// be redirected), while the preventable version is prevented there.
func TestDamageCantBePreventedSurvivesRedirection(t *testing.T) {
	for _, unpreventable := range []bool{true, false} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		redirect := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectRedirectDamageToSource,
			Controller:     game.Player2,
			AffectedPlayer: game.PlayerYou,
			SourceObjectID: redirect.ObjectID,
			Duration:       game.DurationPermanent,
		})
		resolveInstruction(engine, g, &game.StackObject{
			Controller: game.Player2,
			Targets:    []game.Target{game.PermanentTarget(redirect.ObjectID)},
		}, game.PreventDamage{All: true, Object: game.TargetPermanentReference(0)}, nil)
		life := g.Players[game.Player2].Life

		addEffectSpellToStack(g, game.Player1, game.Damage{
			Amount:          game.Fixed(3),
			Recipient:       game.AnyTargetDamageRecipient(0),
			CantBePrevented: unpreventable,
		}, []game.Target{game.PlayerTarget(game.Player2)})
		engine.resolveTopOfStack(g, &TurnLog{})

		want := 0
		if unpreventable {
			want = 3
		}
		if g.Players[game.Player2].Life != life {
			t.Fatalf("unpreventable=%v redirected damage changed player life", unpreventable)
		}
		if redirect.MarkedDamage != want {
			t.Fatalf("unpreventable=%v redirect creature damage = %d, want %d", unpreventable, redirect.MarkedDamage, want)
		}
	}
}

// TestDamageCantBePreventedThisTurnRule proves an unscoped "Damage can't be
// prevented this turn." rule (Skullcrack) covers combat and noncombat damage
// from every source until cleanup, then prevention works again.
func TestDamageCantBePreventedThisTurnRule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	victim := addCombatCreaturePermanentWithPower(g, game.Player1, 10)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageCantBePrevented}},
		Duration:    game.DurationUntilEndOfTurn,
	}, nil)

	if dealt := dealPlayerDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player2, game.Player1, 3, true); dealt != 3 {
		t.Fatalf("combat damage this turn = %d, want 3", dealt)
	}
	if dealt := dealPermanentDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player2, victim, 3, false); dealt != 3 {
		t.Fatalf("noncombat damage this turn = %d, want 3", dealt)
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
	if dealt := dealPlayerDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player2, game.Player1, 3, false); dealt != 0 {
		t.Fatalf("damage after cleanup = %d, want prevented", dealt)
	}
}

func unpreventableStaticCreature(effect game.RuleEffect, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            "Unpreventable Static",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: power}),
		Toughness:       opt.Val(game.PT{Value: power}),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{effect}}},
	}}
}

// TestDamageCantBePreventedStaticRule proves an unscoped static "Damage can't
// be prevented." (Leyline of Punishment) covers every source's combat and
// noncombat damage, including spells, only while its source is on the
// battlefield.
func TestDamageCantBePreventedStaticRule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	leyline := addCombatPermanent(g, game.Player1, unpreventableStaticCreature(game.RuleEffect{Kind: game.RuleEffectDamageCantBePrevented}, 1))
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	victim := addCombatCreaturePermanentWithPower(g, game.Player3, 10)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)

	if dealt := dealPlayerDamage(g, other.CardInstanceID, other.ObjectID, game.Player2, game.Player1, 3, true); dealt != 3 {
		t.Fatalf("opponent creature combat damage = %d, want 3", dealt)
	}
	if dealt := dealPermanentDamage(g, other.CardInstanceID, other.ObjectID, game.Player2, victim, 3, false); dealt != 3 {
		t.Fatalf("opponent creature noncombat damage = %d, want 3", dealt)
	}
	life := g.Players[game.Player3].Life
	addEffectSpellToStack(g, game.Player2, game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)},
		[]game.Target{game.PlayerTarget(game.Player3)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if lost := life - g.Players[game.Player3].Life; lost != 2 {
		t.Fatalf("spell damage = %d, want 2", lost)
	}

	if !movePermanentToZone(g, leyline, zone.Graveyard) {
		t.Fatal("could not remove static source")
	}
	if dealt := dealPlayerDamage(g, other.CardInstanceID, other.ObjectID, game.Player2, game.Player1, 3, true); dealt != 0 {
		t.Fatalf("damage after static source left = %d, want prevented", dealt)
	}
}

// TestDamageCantBePreventedSourceScopedStatic proves "Damage that would be
// dealt by this creature can't be prevented." (Excruciator) covers only its own
// combat and noncombat damage, including through protection, and leaves shields
// in place for other sources.
func TestDamageCantBePreventedSourceScopedStatic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	excruciator := addCombatPermanent(g, game.Player1, unpreventableStaticCreature(game.RuleEffect{
		Kind:           game.RuleEffectDamageCantBePrevented,
		AffectedSource: true,
	}, 7))
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	protected := addProtectionFromTypesPermanent(g, game.Player2, types.Creature)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)

	if dealt := dealPlayerDamage(g, excruciator.CardInstanceID, excruciator.ObjectID, game.Player1, game.Player2, 7, true); dealt != 7 {
		t.Fatalf("source combat damage = %d, want 7", dealt)
	}
	if dealt := dealPermanentDamage(g, excruciator.CardInstanceID, excruciator.ObjectID, game.Player1, protected, 1, false); dealt != 1 {
		t.Fatalf("source damage through protection = %d, want 1", dealt)
	}
	if dealt := dealPlayerDamage(g, other.CardInstanceID, other.ObjectID, game.Player1, game.Player2, 3, true); dealt != 0 {
		t.Fatalf("other creature damage = %d, want prevented", dealt)
	}
	if len(g.PreventionShields) != 1 {
		t.Fatalf("prevention shields = %+v, want the global shield intact", g.PreventionShields)
	}
}

// TestDamageCantBePreventedSelectionAndObjectScopes proves the selection-scoped
// static and the object-scoped resolving rule cover only matching sources.
func TestDamageCantBePreventedSelectionAndObjectScopes(t *testing.T) {
	t.Run("selection", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, unpreventableStaticCreature(game.RuleEffect{
			Kind: game.RuleEffectDamageCantBePrevented,
			AffectedSelection: game.Selection{
				Controller:    game.ControllerYou,
				RequiredTypes: []types.Card{types.Creature},
			},
		}, 1))
		mine := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
		theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)

		if dealt := dealPlayerDamage(g, mine.CardInstanceID, mine.ObjectID, game.Player1, game.Player3, 3, false); dealt != 3 {
			t.Fatalf("matching source damage = %d, want 3", dealt)
		}
		if dealt := dealPlayerDamage(g, theirs.CardInstanceID, theirs.ObjectID, game.Player2, game.Player3, 3, false); dealt != 0 {
			t.Fatalf("nonmatching source damage = %d, want prevented", dealt)
		}
	})

	t.Run("object", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		chosen := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		other := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
		resolveInstruction(engine, g, &game.StackObject{
			Controller: game.Player1,
			Targets:    []game.Target{game.PermanentTarget(chosen.ObjectID)},
		}, game.ApplyRule{
			Object:      opt.Val(game.TargetPermanentReference(0)),
			RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageCantBePrevented}},
			Duration:    game.DurationUntilEndOfTurn,
		}, nil)

		if dealt := dealPlayerDamage(g, chosen.CardInstanceID, chosen.ObjectID, game.Player2, game.Player1, 3, true); dealt != 3 {
			t.Fatalf("chosen source damage = %d, want 3", dealt)
		}
		if dealt := dealPlayerDamage(g, other.CardInstanceID, other.ObjectID, game.Player2, game.Player1, 3, true); dealt != 0 {
			t.Fatalf("other source damage = %d, want prevented", dealt)
		}
	})
}

// TestDamageCantBePreventedAuthoredProgramOperation proves an authored Go
// resolving body that executes an unpreventable Damage operation reaches the
// same shared prevention rule as a declarative instruction.
func TestDamageCantBePreventedAuthoredProgramOperation(t *testing.T) {
	for _, unpreventable := range []bool{true, false} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := choicesProgramEngine(t)
		resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
		addEffectSpellToStack(g, game.Player1, game.GainLife{Amount: game.Fixed(0), Player: game.ControllerReference()},
			[]game.Target{game.PlayerTarget(game.Player2)})
		obj, _ := g.Stack.Peek()
		life := g.Players[game.Player2].Life
		ctx := &programContext{engine: engine, game: g, obj: obj, active: true, log: &TurnLog{},
			specs: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}}}

		if _, err := ctx.Execute(game.Damage{
			Amount:          game.Fixed(4),
			Recipient:       game.AnyTargetDamageRecipient(0),
			CantBePrevented: unpreventable,
		}); err != nil {
			t.Fatal(err)
		}
		want := 0
		if unpreventable {
			want = 4
		}
		if lost := life - g.Players[game.Player2].Life; lost != want {
			t.Fatalf("unpreventable=%v authored damage = %d, want %d", unpreventable, lost, want)
		}
	}
}
