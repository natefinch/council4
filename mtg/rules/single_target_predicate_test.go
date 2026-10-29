package rules

// Source review: Engine gap #3512. CR 2026-09-25 (MagicCompRules 20260925.txt,
// "effective as of September 25, 2026", verified local copy) 115.9a (a "[spell
// or ability] with [a number of] targets" counts each time any object or player
// was chosen as a target when it was put on the stack, not how many are
// currently legal; the same object or player chosen twice counts twice), 115.6
// (a spell or ability that allows zero targets is targeted only if one or more
// were chosen), 601.2c (the same object may be chosen once per instance of the
// word "target"), 608.2b (targets are rechecked on resolution), 702.47c-d
// (spliced text is part of the host spell and chooses its own targets) and
// 707.10c (a copy's new targets must be legal). Pinned Oracle: Swerve
// c68627b1-025c-48c9-9646-98eb2d268e71 "Change the target of target spell with
// a single target." (Scryfall rulings 2008-10-01: "If a spell targets multiple
// things, you can't target it with Swerve, even if all but one of those targets
// has become illegal." / "If a spell targets the same player or object multiple
// times, you can't target it with Swerve."); Willbender
// 0aae277e-e58e-4115-b5fd-0459451e17ec "When this creature is turned face up,
// change the target of target spell or ability with a single target." (rulings
// 2018-03-16, same two rulings for spells and abilities); Bolt Bend, Redirect
// Lightning and Return the Favor carry the same rulings; Deflection ruling
// 2004-10-04 "This does not check if the current target is legal. It just
// checks if the spell has a single target."; Fork "Copy target instant or
// sorcery spell, except that the copy is red. You may choose new targets for
// the copy."; Prodigal Sorcerer "{T}: This creature deals 1 damage to any
// target."; Desperate Ritual and Glacial Ray (Splice onto Arcane {1}{R}).

import (
	"slices"
	"testing"

	cardd "github.com/natefinch/council4/mtg/cards/d"
	cardg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// pushUpToTwoCreaturesSpell puts an instant with one "up to two target
// creatures" word on the stack, recording only the targets actually chosen.
func pushUpToTwoCreaturesSpell(g *game.Game, controller game.PlayerID, targets ...game.Target) *game.StackObject {
	spec := retargetCreatureSpec(game.Selection{})
	spec.MinTargets, spec.MaxTargets = 0, 2
	def := &game.CardDef{CardFace: game.CardFace{
		Name:         "Up To Two Creatures Spell",
		Types:        []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Targets: []game.TargetSpec{spec}}.Ability()),
	}}
	addImplementationSpellToStack(g, controller, def, targets)
	obj, _ := g.Stack.Peek()
	obj.TargetCounts = []int{len(targets)}
	return obj
}

// castSwerveAt has Player2 cast the real Swerve from hand targeting victim and
// reports whether the cast was accepted.
func castSwerveAt(t *testing.T, g *game.Game, engine *Engine, victim *game.StackObject) bool {
	t.Helper()
	swerveID := addCardToHand(g, game.Player2, productionChoiceCard(t, "Swerve"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Players[game.Player2].ManaPool.Add(mana.R, 1)
	g.Turn.PriorityPlayer = game.Player2
	return engine.applyCastSpell(g, game.Player2, action.CastSpellAction{
		CardID: swerveID, Targets: []game.Target{game.StackObjectTarget(victim.ID)},
	})
}

// TestSwerveTargetsOnlySpellsWithSingleTarget casts the real Swerve at spells
// with differing target instances. A spell with exactly one chosen target is
// legal, including a variable-count spell that chose one; two target words,
// the same creature chosen twice, a spell whose second target became illegal
// and an untargeted spell are not (CR 115.9a, CR 115.6).
func TestSwerveTargetsOnlySpellsWithSingleTarget(t *testing.T) {
	for _, test := range []struct {
		name  string
		push  func(*game.Game, game.Target, game.Target) *game.StackObject
		legal bool
	}{
		{"one target", func(g *game.Game, a, _ game.Target) *game.StackObject {
			return pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}}, a)
		}, true},
		{"up to two targets with one chosen", func(g *game.Game, a, _ game.Target) *game.StackObject {
			return pushUpToTwoCreaturesSpell(g, game.Player1, a)
		}, true},
		{"two target words", func(g *game.Game, a, b game.Target) *game.StackObject {
			return pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {}}, a, b)
		}, false},
		{"up to two targets with two chosen", func(g *game.Game, a, b game.Target) *game.StackObject {
			return pushUpToTwoCreaturesSpell(g, game.Player1, a, b)
		}, false},
		{"same creature targeted twice", func(g *game.Game, a, _ game.Target) *game.StackObject {
			return pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {}}, a, a)
		}, false},
		{"second target became illegal", func(g *game.Game, a, b game.Target) *game.StackObject {
			obj := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {}}, a, b)
			g.Battlefield = slices.DeleteFunc(g.Battlefield, func(p *game.Permanent) bool { return p.ObjectID == b.PermanentID })
			return obj
		}, false},
		{"no targets", func(g *game.Game, _, _ game.Target) *game.StackObject {
			return pushStackSpell(g, game.Player1, "Untargeted", types.Sorcery)
		}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			a := game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID)
			b := game.PermanentTarget(addCreaturePermanent(g, game.Player3).ObjectID)
			victim := test.push(g, a, b)
			if got := castSwerveAt(t, g, NewEngine(nil), victim); got != test.legal {
				t.Fatalf("casting Swerve at %s = %v, want %v", test.name, got, test.legal)
			}
			if top, _ := g.Stack.Peek(); test.legal != (top.ID != victim.ID) {
				t.Fatalf("Swerve on stack = %v, want %v", top.ID != victim.ID, test.legal)
			}
		})
	}
}

// TestSwerveResolvesAgainstSingleTargetSpell proves the resolution recheck
// keeps a still-single-target spell legal and the target then changes (CR
// 608.2b, CR 115.7a).
func TestSwerveResolvesAgainstSingleTargetSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID)
	b := game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID)
	victim := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}}, a)
	if !castSwerveAt(t, g, engine, victim) {
		t.Fatal("Swerve could not target a spell with a single target")
	}
	agent := &retargetRecordingAgent{t: t, want: []game.Target{b}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
	if len(agent.requests) != 1 {
		t.Fatalf("Swerve change prompts = %d, want 1 (it fizzled or skipped)", len(agent.requests))
	}
	if !slices.Equal(victim.Targets, []game.Target{b}) {
		t.Fatalf("victim targets = %v, want %v", victim.Targets, b)
	}
}

// castDesperateRitualWithSplices has Player1 cast a real Desperate Ritual with
// one spliced Glacial Ray per target, each Ray targeting its listed player.
func castDesperateRitualWithSplices(t *testing.T, g *game.Game, engine *Engine, targets ...game.Target) *game.StackObject {
	t.Helper()
	hostID := addCardToHand(g, game.Player1, cardd.DesperateRitual())
	splices := make([]string, len(targets))
	for i := range targets {
		addCardToHand(g, game.Player1, cardg.GlacialRay())
		splices[i] = "Glacial Ray"
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1+len(targets))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1+len(targets))
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{splices: splices, targets: targets}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(hostID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("casting Desperate Ritual with spliced Glacial Rays failed")
	}
	obj, _ := g.Stack.Peek()
	if len(obj.Targets) != 0 || len(obj.SplicedTargets) != len(targets) {
		t.Fatalf("host targets = %v, spliced = %v; want only %d spliced targets", obj.Targets, obj.SplicedTargets, len(targets))
	}
	return obj
}

// TestSwerveCountsSplicedTargets proves spliced text is part of the host spell
// for the single-target count (CR 702.47c-d, CR 115.9a): an untargeted Arcane
// host with one spliced Glacial Ray has a single target that Swerve changes
// (also in a cloned game), while two spliced Glacial Rays make two targets.
func TestSwerveCountsSplicedTargets(t *testing.T) {
	t.Run("one spliced target", func(t *testing.T) {
		g, engine := newSpliceGame()
		host := castDesperateRitualWithSplices(t, g, engine, game.PlayerTarget(game.Player2))
		swerveSpec := productionChoiceCard(t, "Swerve").SpellAbility.Val.Modes[0].Targets[0]
		clone := g.Clone()
		if !targetMatchesSpec(clone, game.Player2, 0, game.Event{}, &swerveSpec, game.StackObjectTarget(host.ID)) {
			t.Fatal("cloned game lost the spliced single target")
		}
		if !castSwerveAt(t, g, engine, host) {
			t.Fatal("Swerve could not target an Arcane spell whose only target is spliced")
		}
		agent := &retargetRecordingAgent{t: t, want: []game.Target{game.PlayerTarget(game.Player3)}}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
		if !slices.Equal(host.SplicedTargets[0], []game.Target{game.PlayerTarget(game.Player3)}) {
			t.Fatalf("spliced targets = %v, want Player3", host.SplicedTargets)
		}
		life2, life3 := g.Players[game.Player2].Life, g.Players[game.Player3].Life
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player2].Life != life2 || g.Players[game.Player3].Life != life3-2 {
			t.Fatalf("life after Glacial Ray = P2 %d, P3 %d; want Player3 dealt 2", g.Players[game.Player2].Life, g.Players[game.Player3].Life)
		}
	})
	t.Run("two spliced targets", func(t *testing.T) {
		g, engine := newSpliceGame()
		host := castDesperateRitualWithSplices(t, g, engine, game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3))
		if castSwerveAt(t, g, engine, host) {
			t.Fatal("Swerve targeted an Arcane spell with two spliced targets")
		}
	})
}

// TestWillbenderTargetsOnlyAbilitiesWithSingleTarget proves the real Willbender
// trigger accepts a real activated ability with one target and rejects
// abilities with two targets or the same target twice, then changes the
// accepted ability's target on resolution (CR 115.9a, CR 115.7a).
func TestWillbenderTargetsOnlyAbilitiesWithSingleTarget(t *testing.T) {
	activatePing := func(g *game.Game, engine *Engine) *game.StackObject {
		t.Helper()
		sorcerer := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Prodigal Sorcerer"))
		g.Turn.PriorityPlayer = game.Player1
		if !engine.applyActivateAbilityWithChoices(g, game.Player1, action.ActivateAbilityAction{
			SourceID: sorcerer.ObjectID, Targets: []game.Target{game.PlayerTarget(game.Player3)},
		}, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatal("Prodigal Sorcerer activation failed")
		}
		ping, _ := g.Stack.Peek()
		return ping
	}

	t.Run("announcement candidates", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		ping := activatePing(g, NewEngine(nil))
		creature := game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID)
		twoTargets := pushStackActivatedAbility(g, game.Player1)
		twoTargets.Targets, twoTargets.TargetCounts = []game.Target{creature, game.PlayerTarget(game.Player3)}, []int{1, 1}
		twice := pushStackTriggeredAbility(g, game.Player1)
		twice.Targets, twice.TargetCounts = []game.Target{creature, creature}, []int{1, 1}
		single := pushStackTriggeredAbility(g, game.Player1)
		single.Targets, single.TargetCounts = []game.Target{creature}, []int{1}

		spec := productionChoiceCard(t, "Willbender").TriggeredAbilities[0].Content.Modes[0].Targets[0]
		candidates := targetCandidatesForSpecChosenBy(g, game.Player2, game.Player2, nil, 0, game.Event{}, &spec)
		for obj, want := range map[*game.StackObject]bool{ping: true, single: true, twoTargets: false, twice: false} {
			if got := slices.Contains(candidates, game.StackObjectTarget(obj.ID)); got != want {
				t.Fatalf("Willbender candidate %d (targets %v) = %v, want %v", obj.ID, obj.Targets, got, want)
			}
		}
	})

	t.Run("changes single-target ability", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		ping := activatePing(g, engine)
		agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(ping.ID), want: []game.Target{game.PlayerTarget(game.Player4)}}
		turnUpWillbender(t, g, engine, agent)
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
		if !slices.Equal(ping.Targets, []game.Target{game.PlayerTarget(game.Player4)}) {
			t.Fatalf("Prodigal Sorcerer ability targets = %v, want Player4", ping.Targets)
		}
		life := g.Players[game.Player4].Life
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player4].Life != life-1 || g.Players[game.Player3].Life != 40 {
			t.Fatal("changed Prodigal Sorcerer ability did not deal its damage to Player4 only")
		}
	})
}

// TestForkCopyOfSwerveNewTargetsMustHaveSingleTarget proves a copy of Swerve
// may choose new targets only among spells with a single target (CR 707.10c,
// CR 115.9a).
func TestForkCopyOfSwerveNewTargetsMustHaveSingleTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID)
	b := game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID)
	first := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}}, a)
	multi := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {}}, a, b)
	second := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}}, b)
	swerve := pushRetargetHostCard(g, game.Player2, productionChoiceCard(t, "Swerve"), game.StackObjectTarget(first.ID))
	pushRetargetHostCard(g, game.Player2, productionChoiceCard(t, "Fork"), game.StackObjectTarget(swerve.ID))

	agent := &retargetRecordingAgent{t: t, want: []game.Target{game.StackObjectTarget(second.ID)}}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})

	offered := retargetOfferedTargets(agent.requests)
	if !slices.ContainsFunc(offered, func(got []game.Target) bool {
		return slices.Equal(got, []game.Target{game.StackObjectTarget(second.ID)})
	}) {
		t.Fatalf("copy new-target options = %v, want the other single-target spell", offered)
	}
	if slices.ContainsFunc(offered, func(got []game.Target) bool { return slices.Contains(got, game.StackObjectTarget(multi.ID)) }) {
		t.Fatalf("copy new-target options = %v offered the two-target spell", offered)
	}
	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy || !slices.Equal(copyObj.Targets, []game.Target{game.StackObjectTarget(second.ID)}) {
		t.Fatalf("Swerve copy = %+v, want a copy targeting the second spell", copyObj)
	}
}
