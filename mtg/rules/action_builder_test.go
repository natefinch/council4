package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
)

// TestActionBuilderProducesValidActions checks that the builder emits actions
// that pass Action.Validate() for every supported action kind.
func TestActionBuilderProducesValidActions(t *testing.T) {
	cardID := id.ID(1)
	sourceID := id.ID(2)

	t.Run("pass", func(t *testing.T) {
		act := actionBuild.pass()
		if err := act.Validate(); err != nil {
			t.Fatalf("pass() produced invalid action: %v", err)
		}
		if act.Kind != action.ActionPass {
			t.Fatalf("pass() kind = %v, want ActionPass", act.Kind)
		}
	})

	t.Run("playLand", func(t *testing.T) {
		act := actionBuild.playLand(cardID, game.FaceFront)
		if err := act.Validate(); err != nil {
			t.Fatalf("playLand() produced invalid action: %v", err)
		}
		payload, ok := act.PlayLandPayload()
		if !ok {
			t.Fatal("playLand() missing PlayLandPayload")
		}
		if payload.CardID != cardID {
			t.Fatalf("playLand() CardID = %v, want %v", payload.CardID, cardID)
		}
		if payload.Face != game.FaceFront {
			t.Fatalf("playLand() Face = %v, want FaceFront", payload.Face)
		}
	})

	t.Run("castSpell", func(t *testing.T) {
		targets := []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player1}}
		modes := []int{0}
		act := actionBuild.castSpell(cardID, zone.Hand, game.FaceFront, targets, 3, modes)
		if err := act.Validate(); err != nil {
			t.Fatalf("castSpell() produced invalid action: %v", err)
		}
		payload, ok := act.CastSpellPayload()
		if !ok {
			t.Fatal("castSpell() missing CastSpellPayload")
		}
		if payload.CardID != cardID {
			t.Fatalf("castSpell() CardID = %v, want %v", payload.CardID, cardID)
		}
		if payload.KickerPaid {
			t.Fatal("castSpell() KickerPaid = true, want false")
		}
	})

	t.Run("castKickedSpell", func(t *testing.T) {
		act := actionBuild.castKickedSpell(cardID, zone.Hand, game.FaceFront, nil, 0, nil)
		if err := act.Validate(); err != nil {
			t.Fatalf("castKickedSpell() produced invalid action: %v", err)
		}
		payload, ok := act.CastSpellPayload()
		if !ok {
			t.Fatal("castKickedSpell() missing CastSpellPayload")
		}
		if !payload.KickerPaid {
			t.Fatal("castKickedSpell() KickerPaid = false, want true")
		}
	})

	t.Run("activateAbility", func(t *testing.T) {
		act := actionBuild.activateAbility(sourceID, 2, nil, 0)
		if err := act.Validate(); err != nil {
			t.Fatalf("activateAbility() produced invalid action: %v", err)
		}
		payload, ok := act.ActivateAbilityPayload()
		if !ok {
			t.Fatal("activateAbility() missing ActivateAbilityPayload")
		}
		if payload.SourceID != sourceID || payload.AbilityIndex != 2 {
			t.Fatalf("activateAbility() SourceID=%v AbilityIndex=%v, want %v/2", payload.SourceID, payload.AbilityIndex, sourceID)
		}
	})

	t.Run("suspendCard", func(t *testing.T) {
		act := actionBuild.suspendCard(cardID)
		if err := act.Validate(); err != nil {
			t.Fatalf("suspendCard() produced invalid action: %v", err)
		}
		payload, ok := act.SuspendCardPayload()
		if !ok {
			t.Fatal("suspendCard() missing SuspendCardPayload")
		}
		if payload.CardID != cardID {
			t.Fatalf("suspendCard() CardID = %v, want %v", payload.CardID, cardID)
		}
	})

	t.Run("castFaceDown", func(t *testing.T) {
		act := actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownDisguise)
		if err := act.Validate(); err != nil {
			t.Fatalf("castFaceDown() produced invalid action: %v", err)
		}
		payload, ok := act.CastFaceDownPayload()
		if !ok {
			t.Fatal("castFaceDown() missing CastFaceDownPayload")
		}
		if payload.CardID != cardID || payload.Face != game.FaceFront || payload.FaceDownKind != game.FaceDownDisguise {
			t.Fatalf("castFaceDown() payload = %+v, want card %v face front disguise", payload, cardID)
		}
	})

	t.Run("turnFaceUp", func(t *testing.T) {
		act := actionBuild.turnFaceUp(sourceID)
		if err := act.Validate(); err != nil {
			t.Fatalf("turnFaceUp() produced invalid action: %v", err)
		}
		payload, ok := act.TurnFaceUpPayload()
		if !ok {
			t.Fatal("turnFaceUp() missing TurnFaceUpPayload")
		}
		if payload.PermanentID != sourceID {
			t.Fatalf("turnFaceUp() PermanentID = %v, want %v", payload.PermanentID, sourceID)
		}
	})

	t.Run("declareAttackers_nil", func(t *testing.T) {
		act := actionBuild.declareAttackers(nil)
		if err := act.Validate(); err != nil {
			t.Fatalf("declareAttackers(nil) produced invalid action: %v", err)
		}
	})

	t.Run("declareBlockers_nil", func(t *testing.T) {
		act := actionBuild.declareBlockers(nil)
		if err := act.Validate(); err != nil {
			t.Fatalf("declareBlockers(nil) produced invalid action: %v", err)
		}
	})
}

// TestActionBuilderPanicsOnInvalidInput verifies that the builder fails loudly
// when given inputs that would produce an invalid action.
func TestActionBuilderPanicsOnInvalidInput(t *testing.T) {
	cardID := id.ID(1)

	t.Run("playLand_zeroCardID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("playLand(0, FaceFront) did not panic, want panic for zero card ID")
			}
		}()
		actionBuild.playLand(0, game.FaceFront)
	})

	t.Run("castSpell_zeroCardID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("castSpell(0, ...) did not panic, want panic for zero card ID")
			}
		}()
		actionBuild.castSpell(0, zone.Hand, game.FaceFront, nil, 0, nil)
	})

	t.Run("castKickedSpell_zeroCardID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("castKickedSpell(0, ...) did not panic, want panic for zero card ID")
			}
		}()
		actionBuild.castKickedSpell(0, zone.Hand, game.FaceFront, nil, 0, nil)
	})

	t.Run("activateAbility_zeroSourceID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("activateAbility(0, ...) did not panic, want panic for zero source ID")
			}
		}()
		actionBuild.activateAbility(0, 0, nil, 0)
	})

	t.Run("suspendCard_zeroCardID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("suspendCard(0) did not panic, want panic for zero card ID")
			}
		}()
		actionBuild.suspendCard(0)
	})

	t.Run("castFaceDown_zeroCardID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("castFaceDown(0) did not panic, want panic for zero card ID")
			}
		}()
		actionBuild.castFaceDown(0, game.FaceFront, game.FaceDownMorph)
	})

	t.Run("castFaceDown_noKind", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("castFaceDown(..., FaceDownNone) did not panic, want panic for missing face-down kind")
			}
		}()
		actionBuild.castFaceDown(cardID, game.FaceFront, game.FaceDownNone)
	})

	t.Run("turnFaceUp_zeroPermanentID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("turnFaceUp(0) did not panic, want panic for zero permanent ID")
			}
		}()
		actionBuild.turnFaceUp(0)
	})

	t.Run("declareAttackers_zeroAttacker", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("declareAttackers() did not panic, want panic for zero attacker ID")
			}
		}()
		actionBuild.declareAttackers([]game.AttackDeclaration{{Target: game.AttackTarget{Player: game.Player2}}})
	})

	t.Run("declareBlockers_zeroBlocker", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("declareBlockers() did not panic, want panic for zero blocker ID")
			}
		}()
		actionBuild.declareBlockers([]game.BlockDeclaration{{Blocking: id.ID(2)}})
	})
}

// TestActionBuilderSliceIsolation confirms that mutating the slices passed to
// the builder does not affect the resulting action's payload.
func TestActionBuilderSliceIsolation(t *testing.T) {
	cardID := id.ID(1)

	t.Run("castSpell_targets", func(t *testing.T) {
		targets := []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player1}}
		act := actionBuild.castSpell(cardID, zone.Hand, game.FaceFront, targets, 0, nil)
		targets[0] = game.Target{Kind: game.TargetPermanent}
		payload, _ := act.CastSpellPayload()
		if payload.Targets[0].Kind != game.TargetPlayer {
			t.Fatal("castSpell() target was not copied: mutation of input slice changed action payload")
		}
	})

	t.Run("castSpell_modes", func(t *testing.T) {
		modes := []int{1, 2}
		act := actionBuild.castSpell(cardID, zone.Hand, game.FaceFront, nil, 0, modes)
		modes[0] = 99
		payload, _ := act.CastSpellPayload()
		if payload.ChosenModes[0] != 1 {
			t.Fatal("castSpell() modes were not copied: mutation of input slice changed action payload")
		}
	})

	t.Run("activateAbility_targets", func(t *testing.T) {
		sourceID := id.ID(2)
		targets := []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}}
		act := actionBuild.activateAbility(sourceID, 0, targets, 0)
		targets[0] = game.Target{Kind: game.TargetPermanent}
		payload, _ := act.ActivateAbilityPayload()
		if payload.Targets[0].Kind != game.TargetPlayer {
			t.Fatal("activateAbility() targets were not copied: mutation of input slice changed action payload")
		}
	})
}
