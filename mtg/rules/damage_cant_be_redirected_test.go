package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: recipient-scoped damage that can't be prevented or redirected
// (#3639). Comprehensive Rules effective 2026-11-06 (local verified copy
// .cardwork/MagicCompRules-20261106.txt):
//   - 615.12: prevention effects still apply to unpreventable damage but
//     prevent nothing; shields are not reduced.
//   - 614.9: a redirection effect deals damage that would be dealt to one
//     battle, creature, planeswalker or player to another instead.
//   - 616.2: after one replacement or prevention effect applies, the modified
//     event is checked again, so damage redirected onto a recipient meets the
//     effects that apply to that recipient.
// Scryfall Oracle (pinned corpus 6da9ced0...):
//   - Whippoorwill 84050a10 "{G}{G}, {T}: Target creature can't be regenerated
//     this turn. Damage that would be dealt to that creature this turn can't be
//     prevented or dealt instead to another permanent or player. When the
//     creature dies this turn, exile the creature." (no rulings)
//   - Lava Burst f7be3da5 "Lava Burst deals X damage to any target. If Lava
//     Burst would deal damage to a creature, that damage can't be prevented or
//     dealt instead to another permanent or player." 2008-10-01 rulings:
//     prevention and redirection have no effect on its damage to a creature but
//     work normally on damage it would deal to a player.
// Jade Monolith and Beacon of Destiny (see prevent_next_source_redirect_test.go)
// supply the real one-shot redirection shields.

func lockedDamageJadeMonolith(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, redirectTestCard("Jade Monolith", []types.Card{types.Artifact}, 0, 0,
		redirectAbility("{1}: The next time a source of your choice would deal damage to target creature this turn, that source deals that damage to you instead.",
			opt.Val(cost.Mana{cost.O(1)}), nil,
			[]game.TargetSpec{redirectCreatureTarget("creature", game.ControllerAny, game.TargetChooserController)},
			game.PreventDamage{
				Object: game.TargetPermanentReference(0), All: true, OneShot: true, Redirect: game.DamageRedirectToController,
			})))
}

func lavaBurstOnStack(g *game.Game, controller game.PlayerID, target game.Target) id.ID {
	card := addEffectSpellToStack(g, controller, game.Damage{
		Amount:    game.Fixed(5),
		Recipient: game.AnyTargetDamageRecipient(0),
		CreatureDamageCantBePreventedOrRedirected: true,
	}, []game.Target{target})
	def := g.CardInstances[card].Def
	def.Name, def.Types = "Lava Burst", []types.Card{types.Sorcery}
	return card
}

func preventAllDamageEverywhere(engine *Engine, g *game.Game) {
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
}

// TestLavaBurstCreatureDamageIgnoresPreventionAndRedirection plays Lava Burst
// at a creature whose controller has a "prevent all damage" shield and
// activates Jade Monolith choosing Lava Burst: the creature takes all the
// damage, the controller loses no life, and both shields remain unused.
func TestLavaBurstCreatureDamageIgnoresPreventionAndRedirection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	monolith := lockedDamageJadeMonolith(g, game.Player2)
	preventAllDamageEverywhere(engine, g)
	burst := lavaBurstOnStack(g, game.Player1, game.PermanentTarget(creature.ObjectID))
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	target := game.PermanentTarget(creature.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: redirectChoiceAgent(t, burst, target)}

	activateRedirect(t, engine, g, game.Player2, monolith, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)
	if creature.MarkedDamage != 5 || g.Players[game.Player2].Life != 40 || len(g.PreventionShields) != 2 {
		t.Fatalf("after Lava Burst: creature damage %d, life %d, shields %+v; want 5, 40 and both shields unused",
			creature.MarkedDamage, g.Players[game.Player2].Life, g.PreventionShields)
	}
}

// TestLavaBurstPlayerDamageIsPreventedAndRedirectedNormally proves Lava Burst's
// damage to a player is prevented normally, and that damage Beacon of Destiny
// redirects from the player onto itself becomes damage Lava Burst would deal
// to a creature, which then can't be prevented (CR 616.2).
func TestLavaBurstPlayerDamageIsPreventedAndRedirectedNormally(t *testing.T) {
	t.Run("prevented", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		preventAllDamageEverywhere(engine, g)
		lavaBurstOnStack(g, game.Player1, game.PlayerTarget(game.Player2))
		resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
		if g.Players[game.Player2].Life != 40 {
			t.Fatalf("player life = %d, want Lava Burst's player damage prevented", g.Players[game.Player2].Life)
		}
	})

	t.Run("redirected onto a creature", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		beacon := addCombatPermanent(g, game.Player2, redirectTestCard("Beacon of Destiny", []types.Card{types.Creature}, 1, 10,
			redirectAbility("{T}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to this creature instead.",
				opt.V[cost.Mana]{}, cost.Tap, nil, game.PreventDamage{
					Player: game.ControllerReference(), All: true, OneShot: true,
					Redirect: game.DamageRedirectToObject, RedirectObject: game.SourcePermanentReference(),
				})))
		preventAllDamageEverywhere(engine, g)
		burst := lavaBurstOnStack(g, game.Player1, game.PlayerTarget(game.Player2))
		// The player applies Beacon's redirection before the prevention shield
		// (CR 616.1); the redirected damage then meets that shield at Beacon.
		sources := sourceChoiceAgent(t, burst, nil)
		agents := [game.NumPlayers]PlayerAgent{game.Player2: &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind == game.ChoiceReplacement {
				for _, option := range request.Options {
					if strings.Contains(option.Label, "redirection") {
						return []int{option.Index}
					}
				}
				t.Fatalf("redirection was not offered in %+v", request.Options)
			}
			return sources.choice(obs, request)
		}}}

		activateRedirect(t, engine, g, game.Player2, beacon, 0, nil, agents)
		resolveStackWithTriggers(engine, g, agents)
		if g.Players[game.Player2].Life != 40 || beacon.MarkedDamage != 5 || len(g.PreventionShields) != 1 {
			t.Fatalf("after Lava Burst: life %d, Beacon damage %d, shields %+v; want 40, 5 and only the global shield left",
				g.Players[game.Player2].Life, beacon.MarkedDamage, g.PreventionShields)
		}
	})
}

// TestWhippoorwillLocksTargetCreatureDamageThisTurn plays Whippoorwill's
// damage clause on a creature, then Jade Monolith choosing a Bolt aimed at it,
// under a "prevent all damage" shield: the Bolt is dealt to that creature,
// neither shield is used, another creature's damage is still prevented, and
// the lock ends with the turn.
func TestWhippoorwillLocksTargetCreatureDamageThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	monolith := lockedDamageJadeMonolith(g, game.Player2)
	whippoorwill := addCombatPermanent(g, game.Player1, redirectTestCard("Whippoorwill", []types.Card{types.Creature}, 1, 1,
		redirectAbility("{T}: Damage that would be dealt to target creature this turn can't be prevented or dealt instead to another permanent or player.",
			opt.V[cost.Mana]{}, cost.Tap,
			[]game.TargetSpec{redirectCreatureTarget("creature", game.ControllerAny, game.TargetChooserController)},
			game.ApplyRule{
				Object:      opt.Val(game.TargetPermanentReference(0)),
				RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageToCantBePreventedOrRedirected}},
				Duration:    game.DurationUntilEndOfTurn,
			})))
	preventAllDamageEverywhere(engine, g)
	target := game.PermanentTarget(creature.ObjectID)
	bolt := boltOnStack(g, game.Player3, target)
	g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: redirectChoiceAgent(t, bolt, target),
		game.Player2: redirectChoiceAgent(t, bolt, target),
	}

	activateRedirect(t, engine, g, game.Player2, monolith, 0, []game.Target{target}, agents)
	activateRedirect(t, engine, g, game.Player1, whippoorwill, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)
	if creature.MarkedDamage != 3 || g.Players[game.Player2].Life != 40 || len(g.PreventionShields) != 2 {
		t.Fatalf("after Bolt: creature damage %d, life %d, shields %+v; want 3, 40 and both shields unused",
			creature.MarkedDamage, g.Players[game.Player2].Life, g.PreventionShields)
	}
	if dealt := dealPermanentDamage(g, bolt, 0, game.Player3, other, 3, false); dealt != 0 {
		t.Fatalf("damage to another creature = %d, want prevented", dealt)
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	preventAllDamageEverywhere(engine, g)
	if dealt := dealPermanentDamage(g, bolt, 0, game.Player3, creature, 3, false); dealt != 0 {
		t.Fatalf("damage to the creature after cleanup = %d, want prevented", dealt)
	}
}

// TestCreatureLockedDamageAuthoredProgramOperation proves an authored Go
// resolving body executing Lava Burst's Damage operation reaches the same
// shared recipient lock: creature damage is dealt through a prevention shield
// while player damage is prevented.
func TestCreatureLockedDamageAuthoredProgramOperation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	preventAllDamageEverywhere(engine, g)
	addEffectSpellToStack(g, game.Player1, game.GainLife{Amount: game.Fixed(0), Player: game.ControllerReference()},
		[]game.Target{game.PermanentTarget(creature.ObjectID), game.PlayerTarget(game.Player2)})
	obj, _ := g.Stack.Peek()
	anyTarget := game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent | game.TargetAllowPlayer}
	ctx := &programContext{engine: engine, game: g, obj: obj, active: true, log: &TurnLog{},
		specs: []game.TargetSpec{anyTarget, anyTarget}}

	for index := range 2 {
		if _, err := ctx.Execute(game.Damage{
			Amount:    game.Fixed(4),
			Recipient: game.AnyTargetDamageRecipient(index),
			CreatureDamageCantBePreventedOrRedirected: true,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if creature.MarkedDamage != 4 || g.Players[game.Player2].Life != 40 {
		t.Fatalf("authored damage: creature %d, life %d; want 4 and player damage prevented",
			creature.MarkedDamage, g.Players[game.Player2].Life)
	}
}

// TestWhippoorwillLockStopsExcessDamageRedirection plays the registered Flame
// Spill ("Flame Spill deals 4 damage to target creature. Excess damage is dealt
// to that creature's controller instead.") at an indestructible 2/2. Under
// Whippoorwill's lock the self-replacement can't deal any of the damage to the
// controller instead, so the creature is dealt all 4; without the lock the
// excess 2 is redirected normally.
func TestWhippoorwillLockStopsExcessDamageRedirection(t *testing.T) {
	for _, locked := range []bool{true, false} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		creature := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Indestructible)
		target := game.PermanentTarget(creature.ObjectID)
		if locked {
			resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1, Targets: []game.Target{target}}, game.ApplyRule{
				Object:      opt.Val(game.TargetPermanentReference(0)),
				RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectDamageToCantBePreventedOrRedirected}},
				Duration:    game.DurationUntilEndOfTurn,
			}, nil)
		}
		def := cards.NewDefaultRegistry().Lookup("Flame Spill")
		if def == nil || !def.SpellAbility.Exists {
			t.Fatal("registered Flame Spill missing")
		}
		card := addInstructionSpellToStackForController(g, game.Player1, def.SpellAbility.Val.Modes[0].Sequence, []game.Target{target})
		g.CardInstances[card].Def = def
		engine.resolveTopOfStack(g, &TurnLog{})

		wantDamage, wantLife := 2, 38
		if locked {
			wantDamage, wantLife = 4, 40
		}
		if creature.MarkedDamage != wantDamage || g.Players[game.Player2].Life != wantLife {
			t.Fatalf("locked=%v: creature damage %d, controller life %d; want %d and %d",
				locked, creature.MarkedDamage, g.Players[game.Player2].Life, wantDamage, wantLife)
		}
	}
}

// TestCreatureLockedDamageKeepsExcessOnTheCreature proves the same rule for an
// instruction that locks its own damage to creatures (Lava Burst's clause): a
// Damage with an excess redirect deals all of it to the creature.
func TestCreatureLockedDamageKeepsExcessOnTheCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Indestructible)
	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:          game.Fixed(5),
		Recipient:       game.AnyTargetDamageRecipient(0),
		ExcessRecipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
		CreatureDamageCantBePreventedOrRedirected: true,
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if creature.MarkedDamage != 5 || g.Players[game.Player2].Life != 40 {
		t.Fatalf("creature damage %d, controller life %d; want 5 and 40", creature.MarkedDamage, g.Players[game.Player2].Life)
	}
}
