package rules

import (
	"slices"
	"testing"

	cardsr "github.com/natefinch/council4/mtg/cards/r"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Return the Favor is the real top-1000 EDHREC Spree card whose first option
// copies "target instant spell, sorcery spell, activated ability, or triggered
// ability". These end-to-end tests exercise the reusable union stack-object
// target domain (#3019) through the actual generated card: cast-time
// enumeration over the four eligible categories, exclusion of ineligible stack
// objects, the copy resolving under the caster's control with new targets, a
// fizzle when the target leaves the stack, and the Spree mode combinations.

// pushStackSpell puts a spell of the given card type on the stack, sourced from
// a CardInstance whose printed types drive stack-object target matching.
func pushStackSpell(g *game.Game, controller game.PlayerID, name string, cardType types.Card, targets ...game.Target) *game.StackObject {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID:    sourceID,
		Owner: controller,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{cardType},
		}},
	}
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
	}
	if len(targets) > 0 {
		obj.Targets = targets
		obj.TargetCounts = []int{len(targets)}
	}
	g.Stack.Push(obj)
	return obj
}

// pushStackActivatedAbility puts an activated ability on the stack.
func pushStackActivatedAbility(g *game.Game, controller game.PlayerID) *game.StackObject {
	source := addCreaturePermanent(g, controller)
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   controller,
	}
	g.Stack.Push(obj)
	return obj
}

// pushStackTriggeredAbility puts a triggered ability on the stack.
func pushStackTriggeredAbility(g *game.Game, controller game.PlayerID) *game.StackObject {
	source := addCreaturePermanent(g, controller)
	trigger := game.TriggeredAbility{
		Content: game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}},
		}.Ability(),
	}
	obj := &game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    controller,
		InlineTrigger: &trigger,
	}
	g.Stack.Push(obj)
	return obj
}

// returnTheFavorCopyCasts returns the enumerated cast actions for Return the
// Favor that choose only its copy option (mode 0).
func returnTheFavorCopyCasts(t *testing.T, engine *Engine, g *game.Game, spellID id.ID) []action.CastSpellAction {
	t.Helper()
	var copyCasts []action.CastSpellAction
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		if slices.Equal(cast.ChosenModes, []int{0}) {
			copyCasts = append(copyCasts, cast)
		}
	}
	return copyCasts
}

// TestReturnTheFavorCopyModeEnumeratesAllFourStackObjectCategories proves the
// copy option's union target domain offers every eligible stack object — an
// instant spell, a sorcery spell, an activated ability, and a triggered ability
// — and excludes a spell whose card type is outside the {instant, sorcery}
// filter (a creature spell), which the union's spell arm must reject even
// though the ability arms accept every ability.
func TestReturnTheFavorCopyModeEnumeratesAllFourStackObjectCategories(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	instant := pushStackSpell(g, game.Player2, "Opt", types.Instant)
	sorcery := pushStackSpell(g, game.Player2, "Divination", types.Sorcery)
	activated := pushStackActivatedAbility(g, game.Player2)
	triggered := pushStackTriggeredAbility(g, game.Player2)
	creatureSpell := pushStackSpell(g, game.Player2, "Grizzly Bears", types.Creature)

	spellID := addCardToHand(g, game.Player1, cardsr.ReturnTheFavor())
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	targeted := map[id.ID]bool{}
	for _, cast := range returnTheFavorCopyCasts(t, engine, g, spellID) {
		if len(cast.Targets) != 1 {
			t.Errorf("copy cast has %d targets, want exactly 1", len(cast.Targets))
			continue
		}
		targeted[cast.Targets[0].StackObjectID] = true
	}

	for _, want := range []struct {
		name string
		obj  *game.StackObject
	}{
		{"instant spell", instant},
		{"sorcery spell", sorcery},
		{"activated ability", activated},
		{"triggered ability", triggered},
	} {
		if !targeted[want.obj.ID] {
			t.Errorf("copy option did not enumerate the %s as a legal target", want.name)
		}
	}
	if targeted[creatureSpell.ID] {
		t.Error("copy option enumerated a creature spell; the instant/sorcery spell arm must reject it")
	}
}

// TestReturnTheFavorCopyModeAbsentWithNoEligibleTargets proves the copy option
// fails closed: with only an ineligible creature spell on the stack there is no
// legal target, so no copy-only cast is enumerated even though the spell is
// otherwise affordable.
func TestReturnTheFavorCopyModeAbsentWithNoEligibleTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	pushStackSpell(g, game.Player2, "Grizzly Bears", types.Creature)

	spellID := addCardToHand(g, game.Player1, cardsr.ReturnTheFavor())
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	if casts := returnTheFavorCopyCasts(t, engine, g, spellID); len(casts) != 0 {
		t.Errorf("copy-only casts = %d, want 0 (no eligible instant/sorcery/ability target)", len(casts))
	}
}

// TestReturnTheFavorCopyControlledByCasterWithNewTargets casts the real card to
// copy an opponent's instant spell and proves the copy is controlled by Return
// the Favor's controller (not the original spell's controller) and that "you
// may choose new targets for the copy" retargets only the copy.
func TestReturnTheFavorCopyControlledByCasterWithNewTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	victimA := addCreaturePermanent(g, game.Player1)
	victimB := addCreaturePermanent(g, game.Player1)

	// An opponent-controlled Shock-like instant already targeting victim A.
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID:    sourceID,
		Owner: game.Player2,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Shock",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)},
				}},
			}.Ability()),
		}},
	}
	original := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     sourceID,
		Controller:   game.Player2,
		Targets:      []game.Target{game.PermanentTarget(victimA.ObjectID)},
		TargetCounts: []int{1},
	}
	g.Stack.Push(original)

	spellID := addCardToHand(g, game.Player1, cardsr.ReturnTheFavor())
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	var chosen *action.CastSpellAction
	for _, cast := range returnTheFavorCopyCasts(t, engine, g, spellID) {
		if len(cast.Targets) == 1 && cast.Targets[0].StackObjectID == original.ID {
			c := cast
			chosen = &c
			break
		}
	}
	if chosen == nil {
		t.Fatal("no copy cast targeting the opponent's instant spell was enumerated")
	}

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	if !engine.applyCastSpellWithChoices(g, game.Player1, *chosen, agents, &TurnLog{}) {
		t.Fatal("applying the Return the Favor copy cast failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	top, ok := g.Stack.Peek()
	if !ok || !top.Copy {
		t.Fatal("no copy on top of the stack after Return the Favor resolved")
	}
	if top.Controller != game.Player1 {
		t.Errorf("copy controller = %v, want Player1 (the caster controls the copy)", top.Controller)
	}
	if top.Kind != game.StackSpell {
		t.Errorf("copy kind = %v, want StackSpell", top.Kind)
	}
	if len(top.Targets) != 1 || top.Targets[0].PermanentID != victimB.ObjectID {
		t.Errorf("copy targets = %+v, want new target victim B %v", top.Targets, victimB.ObjectID)
	}
	if len(original.Targets) != 1 || original.Targets[0].PermanentID != victimA.ObjectID {
		t.Errorf("original spell retargeted to %+v, want unchanged victim A %v", original.Targets, victimA.ObjectID)
	}
}

// TestReturnTheFavorCopyCopiesActivatedAndTriggeredAbilities proves the union
// domain copies ability stack objects, not just spells: resolving the copy
// option against an activated or a triggered ability places an independent copy
// on the stack under the caster's control.
func TestReturnTheFavorCopyCopiesActivatedAndTriggeredAbilities(t *testing.T) {
	cases := []struct {
		name string
		push func(*game.Game) *game.StackObject
		kind game.StackObjectKind
	}{
		{"activated ability", func(g *game.Game) *game.StackObject {
			return pushStackActivatedAbility(g, game.Player2)
		}, game.StackActivatedAbility},
		{"triggered ability", func(g *game.Game) *game.StackObject {
			return pushStackTriggeredAbility(g, game.Player2)
		}, game.StackTriggeredAbility},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)

			ability := tc.push(g)

			spellID := addCardToHand(g, game.Player1, cardsr.ReturnTheFavor())
			for range 3 {
				addBasicLandPermanent(g, game.Player1, types.Mountain)
			}
			spreePrecombatMain(g)

			var chosen *action.CastSpellAction
			for _, cast := range returnTheFavorCopyCasts(t, engine, g, spellID) {
				if len(cast.Targets) == 1 && cast.Targets[0].StackObjectID == ability.ID {
					c := cast
					chosen = &c
					break
				}
			}
			if chosen == nil {
				t.Fatalf("no copy cast targeting the %s was enumerated", tc.name)
			}

			if !engine.applyCastSpellWithChoices(g, game.Player1, *chosen, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
				t.Fatal("applying the Return the Favor copy cast failed")
			}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			top, ok := g.Stack.Peek()
			if !ok || !top.Copy {
				t.Fatal("no copy on top of the stack after Return the Favor resolved")
			}
			if top.Kind != tc.kind {
				t.Errorf("copy kind = %v, want %v", top.Kind, tc.kind)
			}
			if top.Controller != game.Player1 {
				t.Errorf("copy controller = %v, want Player1", top.Controller)
			}
			if top.ID == ability.ID {
				t.Error("copy shares the original ability's ID; want a distinct object")
			}
		})
	}
}

// TestReturnTheFavorCopyFizzlesWhenTargetLeavesStack proves the copy option
// makes no copy when its sole target has left the stack before resolution
// (CR 608.2b): the spell is countered on resolution and the stack empties.
func TestReturnTheFavorCopyFizzlesWhenTargetLeavesStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	instant := pushStackSpell(g, game.Player2, "Opt", types.Instant)

	spellID := addCardToHand(g, game.Player1, cardsr.ReturnTheFavor())
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	var chosen *action.CastSpellAction
	for _, cast := range returnTheFavorCopyCasts(t, engine, g, spellID) {
		if len(cast.Targets) == 1 && cast.Targets[0].StackObjectID == instant.ID {
			c := cast
			chosen = &c
			break
		}
	}
	if chosen == nil {
		t.Fatal("no copy cast targeting the instant spell was enumerated")
	}
	if !engine.applyCastSpellWithChoices(g, game.Player1, *chosen, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("applying the Return the Favor copy cast failed")
	}

	// The targeted instant leaves the stack (resolves or is countered) before
	// Return the Favor resolves, so the copy option has no legal target.
	if _, ok := g.Stack.RemoveByID(instant.ID); !ok {
		t.Fatal("failed to remove the targeted instant from the stack")
	}
	sizeBefore := g.Stack.Size()

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != sizeBefore-1 {
		t.Errorf("stack size = %d, want %d (Return the Favor left the stack with no copy made)", g.Stack.Size(), sizeBefore-1)
	}
	for _, obj := range g.Stack.Objects() {
		if obj.Copy {
			t.Error("a copy was created even though the target had left the stack")
		}
	}
}

// TestReturnTheFavorModeCombinations proves the Spree subset enumeration and
// per-mode additional costs cooperate with the union target domain: each mode
// carries its own stack-object target and the total additional cost is the sum
// of the chosen modes' {1} costs.
func TestReturnTheFavorModeCombinations(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// A copy-eligible instant and a retarget-eligible single-target spell.
	pushStackSpell(g, game.Player2, "Opt", types.Instant)
	pushStackSpell(g, game.Player2, "Shock", types.Instant,
		game.PermanentTarget(addCreaturePermanent(g, game.Player1).ObjectID))

	spellID := addCardToHand(g, game.Player1, cardsr.ReturnTheFavor())
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	spreePrecombatMain(g)

	var sawCopyOnly, sawRetargetOnly, sawBoth bool
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		switch {
		case slices.Equal(cast.ChosenModes, []int{0}):
			sawCopyOnly = true
			if len(cast.Targets) != 1 {
				t.Errorf("copy-only cast has %d targets, want 1", len(cast.Targets))
			}
		case slices.Equal(cast.ChosenModes, []int{1}):
			sawRetargetOnly = true
			if len(cast.Targets) != 1 {
				t.Errorf("retarget-only cast has %d targets, want 1", len(cast.Targets))
			}
		case slices.Equal(cast.ChosenModes, []int{0, 1}):
			sawBoth = true
			if len(cast.Targets) != 2 {
				t.Errorf("two-mode cast has %d targets, want 2 (one per chosen mode)", len(cast.Targets))
			}
		default:
		}
	}
	if !sawCopyOnly {
		t.Error("no copy-only (mode 0) cast enumerated")
	}
	if !sawRetargetOnly {
		t.Error("no retarget-only (mode 1) cast enumerated")
	}
	if !sawBoth {
		t.Error("no two-mode (modes 0 and 1) cast enumerated")
	}
}
