package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// checkGiveControlPrimitive validates a give-control ApplyContinuous: a control
// layer whose new controller is the chosen target player (NewControllerRef) and
// whose affected object is the supplied reference.
func checkGiveControlPrimitive(t *testing.T, mode game.Mode, seqIdx int, object game.ObjectReference, duration game.EffectDuration) {
	t.Helper()
	prim, ok := mode.Sequence[seqIdx].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[%d] = %T, want game.ApplyContinuous", seqIdx, mode.Sequence[seqIdx].Primitive)
	}
	if !prim.Object.Exists || prim.Object.Val != object {
		t.Fatalf("ApplyContinuous.Object = %v, want %v", prim.Object, object)
	}
	if len(prim.ContinuousEffects) != 1 {
		t.Fatalf("ContinuousEffects len = %d, want 1", len(prim.ContinuousEffects))
	}
	eff := prim.ContinuousEffects[0]
	if eff.Layer != game.LayerControl {
		t.Fatalf("Layer = %v, want LayerControl", eff.Layer)
	}
	if eff.NewController.Exists {
		t.Fatalf("NewController = %v, want unset (give-control uses NewControllerRef)", eff.NewController)
	}
	if !eff.NewControllerRef.Exists || eff.NewControllerRef.Val != game.TargetPlayerReference(0) {
		t.Fatalf("NewControllerRef = %v, want TargetPlayerReference(0)", eff.NewControllerRef)
	}
	if prim.Duration != duration {
		t.Fatalf("Duration = %v, want %v", prim.Duration, duration)
	}
}

func TestLowerGiveControlTwoTarget(t *testing.T) {
	t.Parallel()
	// Donate: target player gains control of a permanent you control.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Donate",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player gains control of target permanent you control.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %d, want 2 (player + permanent)", len(mode.Targets))
	}
	if mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("target[0] allow = %v, want TargetAllowPlayer", mode.Targets[0].Allow)
	}
	if mode.Targets[1].Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("target[1] controller = %v, want ControllerYou", mode.Targets[1].Selection.Val.Controller)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGiveControlPrimitive(t, mode, 0, game.TargetPermanentReference(1), game.DurationPermanent)
}

func TestLowerGiveControlToOpponentCreature(t *testing.T) {
	t.Parallel()
	// Wrong Turn: target opponent gains control of any target creature.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wrong Turn",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target opponent gains control of target creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %d, want 2", len(mode.Targets))
	}
	if mode.Targets[0].Selection.Val.Player != game.PlayerOpponent {
		t.Fatalf("target[0] player = %v, want PlayerOpponent", mode.Targets[0].Selection.Val.Player)
	}
	checkGiveControlPrimitive(t, mode, 0, game.TargetPermanentReference(1), game.DurationPermanent)
}

func TestLowerGiveControlOfSource(t *testing.T) {
	t.Parallel()
	// Jinxed Idol: an activated ability hands the source artifact to an opponent.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Jinxed Idol",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Sacrifice a creature: Target opponent gains control of this artifact.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 (player only)", len(mode.Targets))
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	checkGiveControlPrimitive(t, mode, 0, game.SourcePermanentReference(), game.DurationPermanent)
}

func TestLowerGiveControlRejectsItPronounSource(t *testing.T) {
	t.Parallel()
	// "target opponent gains control of it" binds the pronoun to a target slot
	// rather than the source, so it must fail closed (Sleeper Agent).
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test It Pronoun",
		Layout:     "normal",
		TypeLine:   "Creature — Minion",
		Power:      new("3"),
		Toughness:  new("1"),
		OracleText: "When this creature enters, target opponent gains control of it.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostic for it-pronoun give-control")
	}
}

func TestLowerGainControlPumpHasteSequence(t *testing.T) {
	t.Parallel()
	// Traitorous Instinct: gain control, untap, then a combined +2/+0-and-haste
	// rider on the controlled creature.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Traitorous",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of target creature until end of turn. Untap that creature. It gets +2/+0 and gains haste until end of turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 4 {
		t.Fatalf("mode targets=%d seq=%d, want 1 target 4 instructions", len(mode.Targets), len(mode.Sequence))
	}
	checkGainControlPrimitive(t, mode, 0, game.DurationUntilEndOfTurn)
	checkUntapPrimitive(t, mode, 1)
	modify, ok := mode.Sequence[2].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("sequence[2] = %T, want game.ModifyPT", mode.Sequence[2].Primitive)
	}
	if modify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("ModifyPT.Object = %v, want TargetPermanentReference(0)", modify.Object)
	}
	checkKeywordGrantPrimitive(t, mode, 3, game.Haste)
}
