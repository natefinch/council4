package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

func addNamedTypedPermanent(g *game.Game, controller game.PlayerID, name string, cardTypes ...types.Card) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: cardTypes,
	}})
}

// recordingKeepAgent records how many keep choices it was asked to make and
// answers each with a fixed selection, so a test can prove which player the
// runtime prompts.
type recordingKeepAgent struct {
	calls  int
	answer []int
}

func (*recordingKeepAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a *recordingKeepAgent) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	a.calls++
	return a.answer
}

func permanentPresent(g *game.Game, permanent *game.Permanent) bool {
	_, ok := permanentByObjectID(g, permanent.ObjectID)
	return ok
}

// TestKeepOnePerTypeOpponentsKeepsOnePerType verifies the core behavior: each
// affected opponent keeps one permanent of each present type and sacrifices the
// rest, while the resolving controller's own permanents are untouched.
func TestKeepOnePerTypeOpponentsKeepsOnePerType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addNamedTypedPermanent(g, game.Player1, "Source", types.Planeswalker)
	obj := linkedSourceObject(source)
	mine := addNamedTypedPermanent(g, game.Player1, "My Creature", types.Creature)
	keptCreature := addNamedTypedPermanent(g, game.Player2, "Opp Creature A", types.Creature)
	sacrificedCreature := addNamedTypedPermanent(g, game.Player2, "Opp Creature B", types.Creature)
	keptLand := addNamedTypedPermanent(g, game.Player2, "Opp Land", types.Land)
	keptArtifact := addNamedTypedPermanent(g, game.Player2, "Opp Artifact", types.Artifact)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.KeepOnePerType{
		Players: game.OpponentsReference(),
		Types:   []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker},
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if permanentPresent(g, sacrificedCreature) {
		t.Error("the unkept second opponent creature was not sacrificed")
	}
	for _, permanent := range []*game.Permanent{keptCreature, keptLand, keptArtifact} {
		if !permanentPresent(g, permanent) {
			t.Errorf("permanent %v was sacrificed but should have been kept as the one of its type", permanent.CardInstanceID)
		}
	}
	if !permanentPresent(g, mine) || !permanentPresent(g, source) {
		t.Error("the controller's own permanents must be untouched under the opponents-only scope")
	}
}

// TestKeepOnePerTypeMultitypeSatisfiesMultipleSlots proves one permanent that has
// several of the named types may be kept for several type slots, so the kept set
// is the union of the chosen permanents. A single artifact creature satisfies both
// the artifact and creature slots, so the separate plain creature is sacrificed.
func TestKeepOnePerTypeMultitypeSatisfiesMultipleSlots(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addNamedTypedPermanent(g, game.Player1, "Source", types.Enchantment)
	obj := linkedSourceObject(source)
	artifactCreature := addNamedTypedPermanent(g, game.Player2, "Juggernaut", types.Artifact, types.Creature)
	plainCreature := addNamedTypedPermanent(g, game.Player2, "Bear", types.Creature)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.KeepOnePerType{
		Players: game.OpponentsReference(),
		Types:   []types.Card{types.Artifact, types.Creature},
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !permanentPresent(g, artifactCreature) {
		t.Error("the artifact creature must survive as the kept artifact and creature")
	}
	if permanentPresent(g, plainCreature) {
		t.Error("the plain creature must be sacrificed once the artifact creature satisfied the creature slot")
	}
}

// TestKeepOnePerTypeNonlandPoolLeavesLandsAlone proves a nonland affected pool
// (Cataclysmic Gearhulk) leaves each player's lands untouched.
func TestKeepOnePerTypeNonlandPoolLeavesLandsAlone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addNamedTypedPermanent(g, game.Player1, "Source", types.Artifact, types.Creature)
	obj := linkedSourceObject(source)
	keptCreature := addNamedTypedPermanent(g, game.Player2, "Creature A", types.Creature)
	sacrificedCreature := addNamedTypedPermanent(g, game.Player2, "Creature B", types.Creature)
	land := addNamedTypedPermanent(g, game.Player2, "Forest", types.Land)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.KeepOnePerType{
		Players:           game.OpponentsReference(),
		Types:             []types.Card{types.Creature},
		AffectedSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !permanentPresent(g, land) {
		t.Error("a land must be untouched by a nonland-pool keep-one-per-type")
	}
	if !permanentPresent(g, keptCreature) {
		t.Error("the kept creature was sacrificed")
	}
	if permanentPresent(g, sacrificedCreature) {
		t.Error("the unkept creature was not sacrificed")
	}
}

// TestKeepOnePerTypePhasedOutExcluded proves a phased-out permanent is neither a
// keep candidate nor a sacrifice victim: it stays on the battlefield untouched
// while a normal unkept permanent of the same type is sacrificed.
func TestKeepOnePerTypePhasedOutExcluded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addNamedTypedPermanent(g, game.Player1, "Source", types.Enchantment)
	obj := linkedSourceObject(source)
	keptCreature := addNamedTypedPermanent(g, game.Player2, "Kept", types.Creature)
	phasedCreature := addNamedTypedPermanent(g, game.Player2, "Phased", types.Creature)
	phasedCreature.PhasedOut = true
	victimCreature := addNamedTypedPermanent(g, game.Player2, "Victim", types.Creature)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.KeepOnePerType{
		Players: game.OpponentsReference(),
		Types:   []types.Card{types.Creature},
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !permanentPresent(g, phasedCreature) {
		t.Error("a phased-out permanent must not be sacrificed")
	}
	if !permanentPresent(g, keptCreature) {
		t.Error("the kept creature was sacrificed")
	}
	if permanentPresent(g, victimCreature) {
		t.Error("the unkept, non-phased creature was not sacrificed")
	}
}

// TestKeepOnePerTypeCantBeSacrificedSurvives proves a permanent that can't be
// sacrificed stays on the battlefield even when it is not the kept one of its
// type, while an unprotected unkept permanent is sacrificed.
func TestKeepOnePerTypeCantBeSacrificedSurvives(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addNamedTypedPermanent(g, game.Player1, "Source", types.Enchantment)
	obj := linkedSourceObject(source)
	keptCreature := addNamedTypedPermanent(g, game.Player2, "Kept", types.Creature)
	cantSacrificeControlNotOwnEnchantment(g, game.Player2)
	protected := makeCreaturePermanent(g, game.Player1, "Borrowed Beast")
	protected.Controller = game.Player2
	victimCreature := addNamedTypedPermanent(g, game.Player2, "Victim", types.Creature)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.KeepOnePerType{
		Players: game.OpponentsReference(),
		Types:   []types.Card{types.Creature},
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !permanentPresent(g, protected) {
		t.Error("a can't-be-sacrificed creature must survive even when it is not the kept one of its type")
	}
	if !permanentPresent(g, keptCreature) {
		t.Error("the kept creature was sacrificed")
	}
	if permanentPresent(g, victimCreature) {
		t.Error("the unprotected unkept creature was not sacrificed")
	}
}

// TestKeepOnePerTypeChooserRouting proves the chooser routing: by default each
// affected player chooses their own kept permanents, while ControllerChoosesForAll
// routes every choice to the resolving controller.
func TestKeepOnePerTypeChooserRouting(t *testing.T) {
	setup := func(controllerChooses bool) (p1Calls, p2Calls int, keptFirst bool) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addNamedTypedPermanent(g, game.Player1, "Source", types.Planeswalker)
		obj := linkedSourceObject(source)
		first := addNamedTypedPermanent(g, game.Player2, "First", types.Creature)
		addNamedTypedPermanent(g, game.Player2, "Second", types.Creature)

		controllerAgent := &recordingKeepAgent{answer: []int{1}}
		playerAgent := &recordingKeepAgent{answer: []int{0}}
		var agents [game.NumPlayers]PlayerAgent
		agents[game.Player1] = controllerAgent
		agents[game.Player2] = playerAgent

		engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.KeepOnePerType{
			Players:                 game.OpponentsReference(),
			Types:                   []types.Card{types.Creature},
			ControllerChoosesForAll: controllerChooses,
		}}, agents, &TurnLog{})

		return controllerAgent.calls, playerAgent.calls, permanentPresent(g, first)
	}

	t.Run("each player chooses their own", func(t *testing.T) {
		p1Calls, p2Calls, keptFirst := setup(false)
		if p1Calls != 0 {
			t.Errorf("controller was prompted %d times, want 0 under per-player choice", p1Calls)
		}
		if p2Calls != 1 {
			t.Errorf("affected player was prompted %d times, want 1", p2Calls)
		}
		if !keptFirst {
			t.Error("affected player's own choice (keep the first creature) was not honored")
		}
	})

	t.Run("controller chooses for all", func(t *testing.T) {
		p1Calls, p2Calls, keptFirst := setup(true)
		if p1Calls != 1 {
			t.Errorf("controller was prompted %d times, want 1 under ControllerChoosesForAll", p1Calls)
		}
		if p2Calls != 0 {
			t.Errorf("affected player was prompted %d times, want 0 under ControllerChoosesForAll", p2Calls)
		}
		if keptFirst {
			t.Error("controller's choice (keep the second creature) was not honored; the first should have been sacrificed")
		}
	})
}
