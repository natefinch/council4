package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestSharedStackActionCompositions(t *testing.T) {
	for _, tt := range []struct {
		name, text        string
		control, retarget int
		target            int
	}{
		{"adjacent", "Gain control of target spell. You may choose new targets for it.", 0, 1, 0},
		{"scry prefix", "Scry 1. Gain control of target spell. You may choose new targets for it.", 1, 2, 0},
		{"life rider", "Gain control of target spell. You gain 2 life. You may choose new targets for it.", 0, 2, 0},
		{"scry rider", "Gain control of target spell. Scry 1. You may choose new targets for it.", 0, 2, 0},
		{"player prefix", "Target player gains 1 life. Gain control of target spell. You gain 2 life. You may choose new targets for it.", 1, 3, 1},
		{"card prefix", "Exile target card from your graveyard. Gain control of target spell. You gain 2 life. You may choose new targets for it.", 1, 3, 1},
		{"stack prefix", "Change the target of target spell with a single target. Gain control of target spell. You gain 2 life. You may choose new targets for it.", 1, 3, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Stack Action", Layout: "normal", TypeLine: "Instant", OracleText: tt.text}
			control := fmt.Sprintf("Sequence[%d].Primitive.(game.ChangeStackObjectController)", tt.control)
			retarget := fmt.Sprintf("Sequence[%d].Primitive.(game.ChooseNewTargets)", tt.retarget)
			assertCardPaths(t, card, control+".Object.kind = game.ObjectReferenceTargetStackObject",
				control+".Controller.kind = game.PlayerReferenceController",
				retarget+".Object.kind = game.ObjectReferenceTargetStackObject",
				fmt.Sprintf("Sequence[%d].Optional = true", tt.retarget))
			if tt.target != 0 {
				assertCardPaths(t, card, fmt.Sprintf("%s.Object.targetIndex = %d", control, tt.target),
					fmt.Sprintf("%s.Object.targetIndex = %d", retarget, tt.target))
			} else {
				assertCardPathsAbsent(t, card, control+".Object.targetIndex", retarget+".Object.targetIndex")
			}
			assertCardPathsAbsent(t, card, fmt.Sprintf("Sequence[%d].Optional = true", tt.control),
				"PublishResult", "ResultGate", "TargetCardResultKey", "Primitive.(game.ApplyContinuous)")
		})
	}
}

func TestSharedStackControlMandatoryAssembly(t *testing.T) {
	for _, text := range []string{
		"Gain control of target spell. Scry 1.",
		"Gain control of target spell. Choose new targets for it.",
		"Choose new targets for target spell. You gain 1 life. Gain control of it.",
	} {
		assertCardPaths(t, &ScryfallCard{Name: "Mandatory Stack Action", Layout: "normal", TypeLine: "Instant", OracleText: text},
			"Primitive.(game.ChangeStackObjectController)", "Sequence[1]")
	}
}

func TestSharedStackActionsAcrossShells(t *testing.T) {
	const body = "Gain control of target spell. You gain 2 life. You may choose new targets for it."
	for _, tt := range []struct{ typeLine, text, path string }{
		{"Artifact", "{0}: " + body, "ActivatedAbilities[0].Content"},
		{"Enchantment", "At the beginning of your upkeep, " + body, "TriggeredAbilities[0].Content"},
		{"Instant", "Choose one \u2014\n\u2022 " + body + "\n\u2022 Draw a card.", "SpellAbility.Val"},
	} {
		t.Run(tt.typeLine, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{Name: "Stack Action Shell", Layout: "normal", TypeLine: tt.typeLine, OracleText: tt.text},
				tt.path+".Modes[0].Sequence[0].Primitive.(game.ChangeStackObjectController)",
				tt.path+".Modes[0].Sequence[2].Primitive.(game.ChooseNewTargets)",
				tt.path+".Modes[0].Sequence[2].Optional = true")
		})
	}
}

func TestSharedStackActionsRefuseUnsafeShapes(t *testing.T) {
	for _, text := range []string{
		"Gain control of two target spells. You may choose new targets for them.",
		"Gain control of target spell until end of turn. You may choose new targets for it.",
		"Gain control of target spell. You may choose new targets for that creature.",
		"Gain control of target spell. You may choose new targets for it until end of turn.",
		"Gain control of target spell. You may choose new targets for it. You may spend mana as though it were mana of any color.",
		"Gain control of target spell or ability. You may choose new targets for it.",
		"You may choose new targets for it. Gain control of target spell.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{Name: "Unsafe Stack Action", Layout: "normal", TypeLine: "Instant", OracleText: text})
		})
	}
}

func TestStackControlRemapKeepsActorProjection(t *testing.T) {
	for _, tt := range []struct{ controller, want game.PlayerReference }{
		{game.ControllerReference(), game.ControllerReference()},
		{game.TargetPlayerReference(0), game.TargetPlayerReference(3)},
		{game.ObjectControllerReference(game.TargetStackObjectReference(0)), game.ObjectControllerReference(game.TargetStackObjectReference(3))},
		{game.ObjectOwnerReference(game.TargetStackObjectReference(0)), game.ObjectOwnerReference(game.TargetStackObjectReference(3))},
	} {
		primitive, ok := rebaseTargetedPrimitive(game.ChangeStackObjectController{
			Object: game.TargetStackObjectReference(0), Controller: tt.controller,
		}, 3, 7)
		change, typed := primitive.(game.ChangeStackObjectController)
		if !ok || !typed || change.Object != game.TargetStackObjectReference(3) {
			t.Fatalf("stack subject was not rebased: %v / %v", primitive, ok)
		}
		if change.Controller != tt.want {
			t.Fatalf("controller projection changed: %v, want %v", change.Controller, tt.want)
		}
	}
}
