package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#3546): registered Howl of the Hunt and Volition Reins use
// SourceAttachedPermanentReference for "enchanted creature/permanent" in their
// enters triggers (Oracle and rulings recorded in
// aura_enters_untap_enchanted_test.go). Blocked queue cards Silken Strength,
// Nature's Chosen, Instill Energy, Pemmin's Aura and Airtight Alibi use the same
// reference from triggered or "{cost}: Untap enchanted creature" activated
// abilities. Official CR effective 2026-09-25 (media.wizards.com
// MagicCompRules 20260925.txt) read: 113.7a (an ability on the stack exists
// independently of its source; if the source is no longer where expected, its
// last known information is used), 608.2h (effects use last known information
// of an object no longer in its expected zone), 702.26b (a phased-out permanent
// is treated as though it does not exist and can't affect or be affected by
// anything), 702.26d (phasing doesn't change zones; the permanent is treated as
// not on the battlefield), 702.26g (attachments of a phasing permanent phase out
// indirectly), 702.26i (an Aura that phased out directly phases in attached to
// what it was attached to when it phased out) and 603.4 (an intervening "if" is
// rechecked on resolution). An Aura that phases out directly in response to its
// ability therefore still names the permanent it enchanted as it phased out,
// and that permanent is affected only while it remains phased in as the same
// object.

func phaseOutForTest(t *testing.T, g *game.Game, permanent *game.Permanent) {
	t.Helper()
	if !phaseOutPermanentTree(g, permanent, effectiveController(g, permanent), make(map[game.ObjectID]bool)) || !permanent.PhasedOut {
		t.Fatalf("failed to phase out permanent %v", permanent.ObjectID)
	}
}

func TestAuraEntersUntapUsesPhasedOutAurasAttachment(t *testing.T) {
	t.Run("Howl of the Hunt", func(t *testing.T) {
		g, engine, auraID, wolf := howlScenario(t, types.Wolf)
		aura, triggered := castAuraEntersUntap(t, g, engine, auraID, wolf)
		if !triggered {
			t.Fatal("Howl of the Hunt did not trigger for an enchanted Wolf")
		}
		phaseOutForTest(t, g, aura)
		if wolf.PhasedOut {
			t.Fatal("enchanted Wolf phased out with an Aura that phased out directly")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if wolf.Tapped {
			t.Fatal("enchanted Wolf stayed tapped after Howl of the Hunt phased out in response")
		}
		if !aura.PhasedOut || !aura.AttachedTo.Exists || aura.AttachedTo.Val != wolf.ObjectID {
			t.Fatal("phased-out Aura lost its attachment")
		}
		if !g.Stack.IsEmpty() {
			t.Fatal("stack not empty after the trigger resolved")
		}
	})
	t.Run("Volition Reins", func(t *testing.T) {
		g, engine, auraID, land := volitionScenario(t, true)
		aura, triggered := castAuraEntersUntap(t, g, engine, auraID, land)
		if !triggered {
			t.Fatal("Volition Reins did not trigger for a tapped enchanted permanent")
		}
		phaseOutForTest(t, g, aura)
		if got := effectiveController(g, land); got != game.Player2 {
			t.Fatalf("land controller with Volition Reins phased out = %v, want Player2", got)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if land.Tapped {
			t.Fatal("enchanted permanent stayed tapped after Volition Reins phased out in response")
		}
		if got := effectiveController(g, land); got != game.Player2 {
			t.Fatalf("land controller after the trigger = %v, want Player2", got)
		}
	})
}

// Phasing out the enchanted creature phases the Aura out indirectly; the
// creature is then treated as though it doesn't exist and isn't untapped.
func TestAuraEntersUntapIgnoresPhasedOutEnchantedCreature(t *testing.T) {
	g, engine, auraID, wolf := howlScenario(t, types.Wolf)
	aura, triggered := castAuraEntersUntap(t, g, engine, auraID, wolf)
	if !triggered {
		t.Fatal("Howl of the Hunt did not trigger for an enchanted Wolf")
	}
	phaseOutForTest(t, g, wolf)
	if !aura.PhasedOut {
		t.Fatal("Aura did not phase out indirectly with its enchanted creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !wolf.Tapped {
		t.Fatal("trigger untapped a phased-out creature")
	}
	if !g.Stack.IsEmpty() {
		t.Fatal("stack not empty after the trigger resolved")
	}
}

// A creature that leaves the battlefield while the Aura is phased out is no
// longer the enchanted object, and no other permanent is substituted.
func TestAuraEntersUntapPhasedOutAuraHostLeftBattlefield(t *testing.T) {
	g, engine, auraID, wolf := howlScenario(t, types.Wolf)
	aura, triggered := castAuraEntersUntap(t, g, engine, auraID, wolf)
	if !triggered {
		t.Fatal("Howl of the Hunt did not trigger for an enchanted Wolf")
	}
	phaseOutForTest(t, g, aura)
	if !movePermanentToZone(g, wolf, zone.Graveyard) {
		t.Fatal("failed to remove the Wolf in response")
	}
	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Phasing bystander Wolf",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Wolf},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	other.Tapped = true
	engine.resolveTopOfStack(g, &TurnLog{})
	if !other.Tapped {
		t.Fatal("trigger untapped a Wolf the phased-out Aura never enchanted")
	}
	if !g.Stack.IsEmpty() {
		t.Fatal("stack not empty after the trigger resolved")
	}
}

func untapEnchantedAuraScenario(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Phasing untap test creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	creature.Tapped = true
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Phasing untap test Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			game.EnchantStaticAbility(&game.TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}),
		},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text: "{0}: Untap enchanted creature.",
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Untap{Object: game.SourceAttachedPermanentReference()},
			}}}.Ability(),
		}},
	}})
	if !attachPermanent(g, aura, creature) {
		t.Fatal("failed to attach the test Aura")
	}
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(aura.ObjectID, 0, nil, 0)) {
		t.Fatal("activating the Aura's untap ability failed")
	}
	if top, ok := g.Stack.Peek(); !ok || top.SourceID != aura.ObjectID {
		t.Fatal("Aura's untap ability is not on top of the stack")
	}
	return g, engine, aura, creature
}

// An activated "Untap enchanted creature" ability resolves using the
// attachment its Aura had as it phased out in response.
func TestActivatedUntapEnchantedCreatureAfterAuraPhasesOut(t *testing.T) {
	g, engine, aura, creature := untapEnchantedAuraScenario(t)
	phaseOutForTest(t, g, aura)
	engine.resolveTopOfStack(g, &TurnLog{})
	if creature.Tapped {
		t.Fatal("enchanted creature stayed tapped after the Aura phased out in response")
	}
}

// Once the phased-out Aura reappears it is the same object (CR 702.26c, 702.26d)
// and its ability uses its current attachment rather than the phase-out
// snapshot.
func TestActivatedUntapEnchantedCreatureUsesReappearedAurasCurrentAttachment(t *testing.T) {
	g, engine, aura, creature := untapEnchantedAuraScenario(t)
	phaseOutForTest(t, g, aura)
	aura.PhasedOut = false
	aura.PhasedOutFor = 0
	aura.PhaseInScheduled = false
	delete(g.LastKnownInformation, aura.ObjectID)
	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Phasing reattach test creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	other.Tapped = true
	if !attachPermanent(g, aura, other) {
		t.Fatal("failed to move the phased-in Aura")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if other.Tapped {
		t.Fatal("currently enchanted creature stayed tapped")
	}
	if !creature.Tapped {
		t.Fatal("formerly enchanted creature was untapped by the phased-in Aura's ability")
	}
}

// Phasing out the enchanted creature phases the Aura out indirectly; the
// creature can't be affected while phased out (CR 702.26b, 702.26g).
func TestActivatedUntapEnchantedCreatureIgnoresPhasedOutCreature(t *testing.T) {
	g, engine, aura, creature := untapEnchantedAuraScenario(t)
	phaseOutForTest(t, g, creature)
	if !aura.PhasedOut {
		t.Fatal("Aura did not phase out indirectly with its enchanted creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !creature.Tapped {
		t.Fatal("ability untapped a phased-out creature")
	}
}
