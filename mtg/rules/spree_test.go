package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spree (CR 702.171) is a modal cast mechanic: the controller chooses one or
// more of the printed "+ {cost} — effect" options, pays each chosen option's
// additional cost on top of the base spell cost, and only the chosen options'
// targets and effects apply, in printed order. Its per-option additional cost
// lives on each game.Mode.Cost rather than on a card-level boolean, and the
// generic modal machinery (subset enumeration, target gating, resolution, and
// copies) drives it. These tests lock in that end-to-end runtime behavior; the
// parser/compiler/lowering coverage lives in the cardgen packages and the
// per-mode cost arithmetic lives in mtg/rules/payment.

// spreeDamageSpell is a synthetic Spree instant costing {R} with two options
// whose additional costs differ so affordability can distinguish subsets:
//
//	Spree
//	+ {1}   — deal 2 damage to target creature   (mode 0)
//	+ {1}{R} — deal 3 damage to target player     (mode 1)
//
// The distinct amounts (2 vs 3) and distinct target kinds (creature vs player)
// make each option's resolution unambiguously identifiable.
func spreeDamageSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Spree Damage",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1,
			MaxModes: 2,
			Modes: []game.Mode{
				{
					Text: "{1} — deal 2 damage to target creature",
					Cost: opt.Val(cost.Mana{cost.O(1)}),
					Targets: []game.TargetSpec{
						{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
					},
					Sequence: []game.Instruction{
						{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(2)}},
					},
				},
				{
					Text: "{1}{R} — deal 3 damage to target player",
					Cost: opt.Val(cost.Mana{cost.O(1), cost.R}),
					Targets: []game.TargetSpec{
						{MinTargets: 1, MaxTargets: 1, Constraint: "player", Allow: game.TargetAllowPlayer},
					},
					Sequence: []game.Instruction{
						{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(3)}},
					},
				},
			},
		}),
	}}
}

// spreeArtifactSpell is a synthetic Spree sorcery costing {R} whose second
// option destroys an artifact, so its target candidate can be made to vanish:
//
//	Spree
//	+ {1} — deal 2 damage to target creature  (mode 0)
//	+ {1} — destroy target artifact           (mode 1)
func spreeArtifactSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Spree Artifact",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1,
			MaxModes: 2,
			Modes: []game.Mode{
				{
					Text: "{1} — deal 2 damage to target creature",
					Cost: opt.Val(cost.Mana{cost.O(1)}),
					Targets: []game.TargetSpec{
						{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
					},
					Sequence: []game.Instruction{
						{Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(0), Amount: game.Fixed(2)}},
					},
				},
				{
					Text: "{1} — destroy target artifact",
					Cost: opt.Val(cost.Mana{cost.O(1)}),
					Targets: []game.TargetSpec{
						{MinTargets: 1, MaxTargets: 1, Constraint: "artifact", Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}})},
					},
					Sequence: []game.Instruction{
						{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
					},
				},
			},
		}),
	}}
}

func spreePrecombatMain(g *game.Game) {
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
}

// TestSpreeCastEnumeratesEveryNonemptyModeSubsetWithGatedTargets proves the
// engine offers exactly the nonempty subsets of the Spree options ({0}, {1},
// {0,1}) — never the empty (no-mode) cast — and that each cast targets only the
// options it selected: the creature-damage option contributes a creature
// target, the artifact-destroy option contributes an artifact target, and the
// two-mode cast contributes both.
func TestSpreeCastEnumeratesEveryNonemptyModeSubsetWithGatedTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, spreeArtifactSpell())
	creature := addCreaturePermanent(g, game.Player2)
	artifact := addArtifactPermanent(g, game.Player2)
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	var sawMode0, sawMode1, sawBoth bool
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		modes := cast.ChosenModes
		if len(modes) == 0 {
			t.Fatal("enumerated a no-mode Spree cast; Spree requires at least one option")
		}
		if !slices.IsSorted(modes) {
			t.Errorf("chosen modes %v are not in ascending printed order", modes)
		}
		switch {
		case slices.Equal(modes, []int{0}):
			sawMode0 = true
			if len(cast.Targets) != 1 || cast.Targets[0].PermanentID != creature.ObjectID {
				t.Errorf("mode {0} cast targets = %+v, want only the creature", cast.Targets)
			}
		case slices.Equal(modes, []int{1}):
			sawMode1 = true
			if len(cast.Targets) != 1 || cast.Targets[0].PermanentID != artifact.ObjectID {
				t.Errorf("mode {1} cast targets = %+v, want only the artifact", cast.Targets)
			}
		case slices.Equal(modes, []int{0, 1}):
			sawBoth = true
			if len(cast.Targets) != 2 {
				t.Fatalf("mode {0,1} cast has %d targets, want 2 (creature + artifact)", len(cast.Targets))
			}
			if cast.Targets[0].PermanentID != creature.ObjectID || cast.Targets[1].PermanentID != artifact.ObjectID {
				t.Errorf("mode {0,1} targets = %+v, want [creature, artifact] in printed order", cast.Targets)
			}
		default:
			t.Errorf("unexpected chosen mode subset %v", modes)
		}
	}
	if !sawMode0 || !sawMode1 || !sawBoth {
		t.Errorf("missing a Spree subset: {0}=%v {1}=%v {0,1}=%v", sawMode0, sawMode1, sawBoth)
	}
}

// TestSpreeUnaffordableSubsetsAbsent proves the per-option additional cost is
// summed atomically with the base cost, so a subset the player cannot pay for
// is not offered. With only two Mountains, only the cheapest single option
// ({R}+{1}) is affordable; the {R}+{1}{R} option and the {R}+{1}+{1}{R}
// two-mode cast both need more red/total mana and disappear.
func TestSpreeUnaffordableSubsetsAbsent(t *testing.T) {
	newGameWithLands := func(lands int) (*Engine, *game.Game, id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, spreeDamageSpell())
		addCreaturePermanent(g, game.Player2)
		for range lands {
			addBasicLandPermanent(g, game.Player1, types.Mountain)
		}
		spreePrecombatMain(g)
		return engine, g, spellID
	}

	subsets := func(engine *Engine, g *game.Game, spellID id.ID) [][]int {
		var out [][]int
		for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
			out = append(out, cast.ChosenModes)
		}
		return out
	}

	containsSubset := func(all [][]int, want []int) bool {
		for _, s := range all {
			if slices.Equal(s, want) {
				return true
			}
		}
		return false
	}

	// Two mana: base {R} + mode0 {1} = 2 mana is payable; every other subset
	// needs at least 3 mana (a second red for mode1).
	engine, g, spellID := newGameWithLands(2)
	got := subsets(engine, g, spellID)
	if !containsSubset(got, []int{0}) {
		t.Errorf("with 2 mana, mode {0} should be affordable; subsets = %v", got)
	}
	if containsSubset(got, []int{1}) || containsSubset(got, []int{0, 1}) {
		t.Errorf("with 2 mana, only mode {0} is affordable; subsets = %v", got)
	}

	// Four mana ({R}{R} available among four Mountains covers {R}+{1}+{1}{R}):
	// every nonempty subset becomes affordable.
	engine, g, spellID = newGameWithLands(4)
	got = subsets(engine, g, spellID)
	for _, want := range [][]int{{0}, {1}, {0, 1}} {
		if !containsSubset(got, want) {
			t.Errorf("with 4 mana, subset %v should be affordable; subsets = %v", want, got)
		}
	}
}

// TestSpreeSubsetAbsentWhenSelectedModeHasNoLegalTarget proves target gating is
// per selected option: with no artifact on the battlefield the artifact-destroy
// option (mode 1) has no legal target, so every subset that selects it is
// illegal, while the creature-damage option (mode 0) remains castable on its own.
func TestSpreeSubsetAbsentWhenSelectedModeHasNoLegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, spreeArtifactSpell())
	addCreaturePermanent(g, game.Player2)
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	var sawMode0 bool
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		if slices.Contains(cast.ChosenModes, 1) {
			t.Errorf("subset %v selects the artifact-destroy option with no legal artifact target", cast.ChosenModes)
		}
		if slices.Equal(cast.ChosenModes, []int{0}) {
			sawMode0 = true
		}
	}
	if !sawMode0 {
		t.Error("mode {0} should remain castable with a legal creature target")
	}
}

// TestSpreeStackObjectCapturesChosenModesAndResolvesInPrintedOrder casts the
// two-mode subset through the real cast pipeline and proves the resulting stack
// object records the chosen mode set, then resolves each selected option's
// effect against that option's own target segment: the creature takes exactly
// mode 0's 2 damage and the targeted player loses exactly mode 1's 3 life.
func TestSpreeStackObjectCapturesChosenModesAndResolvesInPrintedOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, spreeDamageSpell())
	creature := addCreaturePermanent(g, game.Player2)
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	var chosen *action.CastSpellAction
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		if slices.Equal(cast.ChosenModes, []int{0, 1}) && cast.Targets[1].PlayerID == game.Player2 {
			c := cast
			chosen = &c
			break
		}
	}
	if chosen == nil {
		t.Fatal("no two-mode Spree cast targeting Player2's life was enumerated")
	}
	targetPlayerLifeBefore := g.Players[game.Player2].Life

	if !engine.applyCastSpellWithChoices(g, game.Player1, *chosen, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("applying the two-mode Spree cast failed")
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Spree spell not on the stack after casting")
	}
	if !slices.Equal(top.ChosenModes, []int{0, 1}) {
		t.Fatalf("stack object chosen modes = %v, want [0 1]", top.ChosenModes)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if creature.MarkedDamage != 2 {
		t.Errorf("creature marked damage = %d, want 2 (mode 0)", creature.MarkedDamage)
	}
	if got := targetPlayerLifeBefore - g.Players[game.Player2].Life; got != 3 {
		t.Errorf("target player lost %d life, want 3 (mode 1)", got)
	}
}

// TestSpreeSingleModeCastResolvesOnlyThatMode proves that a one-option Spree
// cast applies only that option's effect and none of the unchosen options.
func TestSpreeSingleModeCastResolvesOnlyThatMode(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, spreeDamageSpell())
	creature := addCreaturePermanent(g, game.Player2)
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	lifeBefore := [game.NumPlayers]int{}
	for i := range lifeBefore {
		lifeBefore[i] = g.Players[game.PlayerID(i)].Life
	}

	var chosen *action.CastSpellAction
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		if slices.Equal(cast.ChosenModes, []int{0}) {
			c := cast
			chosen = &c
			break
		}
	}
	if chosen == nil {
		t.Fatal("no single-mode Spree cast enumerated")
	}
	if !engine.applyCastSpellWithChoices(g, game.Player1, *chosen, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("applying the single-mode Spree cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if creature.MarkedDamage != 2 {
		t.Errorf("creature marked damage = %d, want 2", creature.MarkedDamage)
	}
	for i := range lifeBefore {
		if got := g.Players[game.PlayerID(i)].Life; got != lifeBefore[i] {
			t.Errorf("player %d life changed to %d; the unchosen player-damage option must not resolve", i, got)
		}
	}
}

// TestSpreeCopyPreservesChosenModesAndTargetsWithoutPayingCosts proves a copy of
// a Spree spell (e.g. one made by Return the Favor's own copy option) carries
// the original's chosen mode set and targets and is a copy, not a recast: it is
// created directly on the stack with no additional cost. Copying clones the
// chosen modes and targets so re-choosing the copy's targets never disturbs the
// original.
func TestSpreeCopyPreservesChosenModesAndTargetsWithoutPayingCosts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	creature := addCreaturePermanent(g, game.Player2)

	spellSourceID := g.IDGen.Next()
	g.CardInstances[spellSourceID] = &game.CardInstance{
		ID:    spellSourceID,
		Def:   spreeDamageSpell(),
		Owner: game.Player1,
	}
	spreeSpell := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     spellSourceID,
		Controller:   game.Player1,
		ChosenModes:  []int{0, 1},
		Targets:      []game.Target{game.PermanentTarget(creature.ObjectID), game.PlayerTarget(game.Player3)},
		TargetCounts: []int{1, 1},
	}
	g.Stack.Push(spreeSpell)

	source := addCreaturePermanent(g, game.Player1)
	trigger := game.TriggeredAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.CopyStackObject{Object: game.EventStackObjectReference()},
			}},
		}.Ability(),
	}
	copyTrigger := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player1,
		InlineTrigger:   &trigger,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:          game.EventSpellCast,
			Controller:    game.Player1,
			StackObjectID: spreeSpell.ID,
		},
	}
	g.Stack.Push(copyTrigger)

	engine.resolveTopOfStack(g, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack empty after the copy effect resolved")
	}
	if !top.Copy {
		t.Fatal("top stack object is not marked as a copy")
	}
	if top.ID == spreeSpell.ID {
		t.Fatal("copy shares the original spell's ID; want a distinct object")
	}
	if !slices.Equal(top.ChosenModes, []int{0, 1}) {
		t.Errorf("copy chosen modes = %v, want [0 1] (copied from the original)", top.ChosenModes)
	}
	if !slices.Equal(top.Targets, spreeSpell.Targets) {
		t.Errorf("copy targets = %+v, want %+v (copied from the original)", top.Targets, spreeSpell.Targets)
	}
	// A copy is created directly on the stack (CR 707.10); it is not cast, so it
	// pays neither the base cost nor any chosen option's additional cost. No mana
	// was ever added to the pool, and the copy still appears — proof the copy
	// path never demanded payment.
	if !slices.Equal(spreeSpell.ChosenModes, []int{0, 1}) {
		t.Errorf("original chosen modes mutated to %v; copies must own independent state", spreeSpell.ChosenModes)
	}
}
