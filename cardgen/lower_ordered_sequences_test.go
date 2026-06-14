package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerAndJoinedLifeSequence(t *testing.T) {
	t.Parallel()
	// "X and you gain/lose N life" sequences (Sign in Blood / Ambition's Cost
	// family) lower as ordered instructions rather than being rejected wholesale.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Cost",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "You draw three cards and you lose 3 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 3 {
		t.Fatalf("first primitive = %+v, want draw three", mode.Sequence[0].Primitive)
	}
	lose, ok := mode.Sequence[1].Primitive.(game.LoseLife)
	if !ok || lose.Amount.Value() != 3 {
		t.Fatalf("second primitive = %+v, want lose 3 life", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffects(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Draw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Destroy); !ok {
		t.Fatalf("first primitive = %T, want game.Destroy", mode.Sequence[0].Primitive)
	}
	draw, ok := mode.Sequence[1].Primitive.(game.Draw)
	if !ok || draw.Amount.Value() != 1 {
		t.Fatalf("second primitive = %+v, want draw one", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffectsWithMultipleTargets(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want two targets and two instructions", mode)
	}
	destroy, ok := mode.Sequence[0].Primitive.(game.Destroy)
	if !ok || destroy.Object.TargetIndex() != 0 {
		t.Fatalf("first primitive = %+v, want target 0 destroy", mode.Sequence[0].Primitive)
	}
	tap, ok := mode.Sequence[1].Primitive.(game.Tap)
	if !ok || tap.Object.TargetIndex() != 1 {
		t.Fatalf("second primitive = %+v, want target 1 tap", mode.Sequence[1].Primitive)
	}
}

func TestLowerOrderedSpellEffectsRebasesEveryTargetClause(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Tap target creature. Target player mills three cards.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 3 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want three targets and three instructions", mode)
	}
	destroy, destroyOK := mode.Sequence[0].Primitive.(game.Destroy)
	tap, tapOK := mode.Sequence[1].Primitive.(game.Tap)
	mill, millOK := mode.Sequence[2].Primitive.(game.Mill)
	if !destroyOK || !tapOK || !millOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Destroy, game.Tap, game.Mill",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if destroy.Object.TargetIndex() != 0 ||
		tap.Object.TargetIndex() != 1 ||
		mill.Player.TargetIndex() != 2 {
		t.Fatalf(
			"target indices = %d, %d, %d; want 0, 1, 2",
			destroy.Object.TargetIndex(),
			tap.Object.TargetIndex(),
			mill.Player.TargetIndex(),
		)
	}
}

func TestLowerThenJoinedSpellSequence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		typeLine    string
		oracleText  string
		checkFirst  func(*testing.T, game.Instruction)
		checkSecond func(*testing.T, game.Instruction)
	}{
		{
			name:       "draw then discard spell",
			typeLine:   "Sorcery",
			oracleText: "Draw two cards, then discard a card.",
			checkFirst: func(t *testing.T, inst game.Instruction) {
				draw, ok := inst.Primitive.(game.Draw)
				if !ok || draw.Amount.Value() != 2 || draw.Player != game.ControllerReference() {
					t.Fatalf("first = %+v, want controller draws 2", inst.Primitive)
				}
			},
			checkSecond: func(t *testing.T, inst game.Instruction) {
				discard, ok := inst.Primitive.(game.Discard)
				if !ok || discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
					t.Fatalf("second = %+v, want controller discards 1", inst.Primitive)
				}
			},
		},
		{
			name:       "scry then draw spell",
			typeLine:   "Sorcery",
			oracleText: "Scry 2, then draw a card.",
			checkFirst: func(t *testing.T, inst game.Instruction) {
				scry, ok := inst.Primitive.(game.Scry)
				if !ok || scry.Amount.Value() != 2 || scry.Player != game.ControllerReference() {
					t.Fatalf("first = %+v, want controller scries 2", inst.Primitive)
				}
			},
			checkSecond: func(t *testing.T, inst game.Instruction) {
				draw, ok := inst.Primitive.(game.Draw)
				if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
					t.Fatalf("second = %+v, want controller draws 1", inst.Primitive)
				}
			},
		},
		{
			name:       "discard then draw spell",
			typeLine:   "Sorcery",
			oracleText: "Discard a card, then draw a card.",
			checkFirst: func(t *testing.T, inst game.Instruction) {
				discard, ok := inst.Primitive.(game.Discard)
				if !ok || discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
					t.Fatalf("first = %+v, want controller discards 1", inst.Primitive)
				}
			},
			checkSecond: func(t *testing.T, inst game.Instruction) {
				draw, ok := inst.Primitive.(game.Draw)
				if !ok || draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
					t.Fatalf("second = %+v, want controller draws 1", inst.Primitive)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Spell",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			if !face.SpellAbility.Exists {
				t.Fatal("spell ability not lowered")
			}
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
				t.Fatalf("mode = %+v, want no targets and two instructions", mode)
			}
			test.checkFirst(t, mode.Sequence[0])
			test.checkSecond(t, mode.Sequence[1])
		})
	}
}

func TestLowerThenJoinedEnterTriggerSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Looting Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, draw a card, then discard a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want no targets and two instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	if !drawOK || !discardOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Draw, game.Discard",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
}

func TestLowerThenJoinedSharedTargetSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mill",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player mills three cards, then draws a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	mill, millOK := mode.Sequence[0].Primitive.(game.Mill)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !millOK || !drawOK {
		t.Fatalf(
			"primitives = %T, %T; want game.Mill, game.Draw",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
		)
	}
	if mill.Amount.Value() != 3 || mill.Player.TargetIndex() != 0 {
		t.Fatalf("mill = %+v, want target player mills 3", mill)
	}
	if draw.Amount.Value() != 1 || draw.Player.TargetIndex() != 0 {
		t.Fatalf("draw = %+v, want target player draws 1", draw)
	}
}

// TestLowerThenJoinedThreeEffectSequence is a regression for a bug where
// 3-effect then-joined chains would assign the wrong clause start for
// effects after the first in the group, causing middle clauses to
// incorrectly include previous effects' tokens.
func TestLowerThenJoinedThreeEffectSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chain",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card, then discard a card, then proliferate.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want no targets and three instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	_, prolifOK := mode.Sequence[2].Primitive.(game.Proliferate)
	if !drawOK || !discardOK || !prolifOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Draw, game.Discard, game.Proliferate",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
}

// TestLowerThenJoinedNonTargetFinalClause is a regression for the case where
// a then-joined sentence is followed by a separate sentence: the final
// clause of the then-group must be bounded to its own sentence and must not
// spill into subsequent-sentence tokens.
func TestLowerThenJoinedNonTargetFinalClause(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Multi",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card, then discard a card. You gain 3 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %+v, want no targets and three instructions", mode)
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	discard, discardOK := mode.Sequence[1].Primitive.(game.Discard)
	gain, gainOK := mode.Sequence[2].Primitive.(game.GainLife)
	if !drawOK || !discardOK || !gainOK {
		t.Fatalf(
			"primitives = %T, %T, %T; want game.Draw, game.Discard, game.GainLife",
			mode.Sequence[0].Primitive,
			mode.Sequence[1].Primitive,
			mode.Sequence[2].Primitive,
		)
	}
	if draw.Amount.Value() != 1 || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %+v, want controller draws 1", draw)
	}
	if discard.Amount.Value() != 1 || discard.Player != game.ControllerReference() {
		t.Fatalf("discard = %+v, want controller discards 1", discard)
	}
	if gain.Amount.Value() != 3 || gain.Player != game.ControllerReference() {
		t.Fatalf("gain = %+v, want controller gains 3", gain)
	}
}

// TestLowerThenJoinedSharedTargetNoExtraSpec is a regression for the target
// deduplication requirement: a shared-subject then-joined sequence
// (e.g. "Target player mills N, then draws M") must produce exactly one
// game.TargetSpec, and both instructions must reference TargetIndex 0.
func TestLowerThenJoinedSharedTargetNoExtraSpec(t *testing.T) {
	t.Parallel()
	// Verify that compound-mill produces exactly one target spec and both
	// instructions reference the same target player at index 0.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Shared Target Test",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player mills three cards, then draws a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want exactly 1 (no duplicate target spec)", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	mill, millOK := mode.Sequence[0].Primitive.(game.Mill)
	draw, drawOK := mode.Sequence[1].Primitive.(game.Draw)
	if !millOK || !drawOK {
		t.Fatalf("primitives = %T, %T, want game.Mill, game.Draw",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive)
	}
	if mill.Player.TargetIndex() != 0 {
		t.Fatalf("mill.Player target index = %d, want 0", mill.Player.TargetIndex())
	}
	if draw.Player.TargetIndex() != 0 {
		t.Fatalf("draw.Player target index = %d, want 0 (reusing existing target)", draw.Player.TargetIndex())
	}
}

// TestLowerThenJoinedSharedTargetAfterEarlierTarget is the exact regression for
// the inherited-target rebase-offset bug. When a then-joined sentence follows an
// earlier sentence that already contributed a target spec, the shared target in
// the then-group is NOT at accumulated-target index 0 — it is at the index where
// the owning clause placed it. Before the fix, allSharedTargets always rebased
// with offset 0, causing the draw to reference the wrong game target (the
// artifact at 0 instead of the player at 1).
//
// Requirements:
//   - Two game.TargetSpec entries: artifact at index 0, target player at index 1.
//   - Destroy references TargetPermanentReference(0).
//   - Mill references TargetPlayerReference(1).
//   - Draw (inherited shared) references TargetPlayerReference(1), not (0).
func TestLowerThenJoinedSharedTargetAfterEarlierTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target artifact. Target player mills three cards, then draws a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 {
		t.Fatalf("targets = %d, want 2 (artifact at 0, target player at 1)", len(mode.Targets))
	}
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %d, want 3 (destroy, mill, draw)", len(mode.Sequence))
	}
	destroy, destroyOK := mode.Sequence[0].Primitive.(game.Destroy)
	mill, millOK := mode.Sequence[1].Primitive.(game.Mill)
	draw, drawOK := mode.Sequence[2].Primitive.(game.Draw)
	if !destroyOK || !millOK || !drawOK {
		t.Fatalf("primitives = %T, %T, %T; want Destroy, Mill, Draw",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive, mode.Sequence[2].Primitive)
	}
	if destroy.Object.TargetIndex() != 0 {
		t.Fatalf("destroy target index = %d, want 0 (artifact)", destroy.Object.TargetIndex())
	}
	if mill.Player.TargetIndex() != 1 {
		t.Fatalf("mill target index = %d, want 1 (target player)", mill.Player.TargetIndex())
	}
	if draw.Player.TargetIndex() != 1 {
		t.Fatalf("draw target index = %d, want 1 (shared target player, NOT 0)", draw.Player.TargetIndex())
	}
}

// TestLowerThenJoinedFightChain is the exact regression for the mixed
// inherited+owned target composition gap. "Target creature fights target
// creature, then fights target creature." requires the second fight to receive
// the inherited subject (T0, already at game index 0) together with its own new
// target (T2, appended at game index 2). Before the fix, inheritedTargets was
// only computed when clauseTargets was empty, so the second effect saw only T2
// and lowerFightSpell (which expects two targets) returned unsupported.
//
// Requirements:
//   - Three game.TargetSpec entries (T0, T1, T2 — all "target creature").
//   - Fight 1: Object=TargetPermanentReference(0), Related=TargetPermanentReference(1).
//   - Fight 2: Object=TargetPermanentReference(0) (inherited T0), Related=TargetPermanentReference(2) (owned T2).
func TestLowerThenJoinedFightChain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature fights target creature, then fights target creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 3 {
		t.Fatalf("targets = %d, want 3 (T0, T1, T2 — one per creature chosen)", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	fight0, ok0 := mode.Sequence[0].Primitive.(game.Fight)
	fight1, ok1 := mode.Sequence[1].Primitive.(game.Fight)
	if !ok0 || !ok1 {
		t.Fatalf("primitives = %T, %T; want game.Fight, game.Fight",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive)
	}
	// Fight 0: T0 fights T1.
	if fight0.Object.TargetIndex() != 0 || fight0.RelatedObject.TargetIndex() != 1 {
		t.Fatalf("fight0 = Object(%d) RelatedObject(%d), want Object(0) RelatedObject(1)",
			fight0.Object.TargetIndex(), fight0.RelatedObject.TargetIndex())
	}
	// Fight 1: inherited T0 fights new T2.
	if fight1.Object.TargetIndex() != 0 || fight1.RelatedObject.TargetIndex() != 2 {
		t.Fatalf("fight1 = Object(%d) RelatedObject(%d), want Object(0) RelatedObject(2)",
			fight1.Object.TargetIndex(), fight1.RelatedObject.TargetIndex())
	}
}

// where the second effect does not use the shared target (proliferate has no
// target) correctly discards the spurious shared target via the fallback
// path, producing one target spec for destroy and a standalone proliferate.
func TestLowerThenJoinedDestroyThenProliferate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Spread",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature, then proliferate.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 (destroy target only, no duplicate)", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	destroy, destroyOK := mode.Sequence[0].Primitive.(game.Destroy)
	_, prolifOK := mode.Sequence[1].Primitive.(game.Proliferate)
	if !destroyOK || !prolifOK {
		t.Fatalf("primitives = %T, %T, want game.Destroy, game.Proliferate",
			mode.Sequence[0].Primitive, mode.Sequence[1].Primitive)
	}
	if destroy.Object.TargetIndex() != 0 {
		t.Fatalf("destroy.Object target index = %d, want 0", destroy.Object.TargetIndex())
	}
}
