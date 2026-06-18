package action

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestPass(t *testing.T) {
	got := Pass()
	if got.Kind != ActionPass {
		t.Fatalf("Pass() kind = %v, want %v", got.Kind, ActionPass)
	}
}

func TestPlayLand(t *testing.T) {
	cardID := id.ID(42)

	got := PlayLand(cardID)
	if got.Kind != ActionPlayLand {
		t.Fatalf("PlayLand() kind = %v, want %v", got.Kind, ActionPlayLand)
	}
	playLand, ok := got.PlayLandPayload()
	if !ok {
		t.Fatal("PlayLandPayload() ok = false, want true")
	}
	if playLand.CardID != cardID {
		t.Fatalf("PlayLand() card ID = %v, want %v", playLand.CardID, cardID)
	}
}

func TestCastSpellPreservesTargets(t *testing.T) {
	cardID := id.ID(42)
	targets := []game.Target{
		game.PlayerTarget(game.Player2),
		game.PermanentTarget(id.ID(99)),
	}
	modes := []int{1}

	got := CastSpell(cardID, targets, 0, modes)
	if got.Kind != ActionCastSpell {
		t.Fatalf("CastSpell() kind = %v, want %v", got.Kind, ActionCastSpell)
	}
	cast, ok := got.CastSpellPayload()
	if !ok {
		t.Fatal("CastSpellPayload() ok = false, want true")
	}
	if cast.CardID != cardID {
		t.Fatalf("CastSpell() card ID = %v, want %v", cast.CardID, cardID)
	}
	if !slices.Equal(cast.Targets, targets) {
		t.Fatalf("CastSpell() targets = %+v, want %+v", cast.Targets, targets)
	}
	targets[0] = game.PlayerTarget(game.Player4)
	modes[0] = 2
	cast, _ = got.CastSpellPayload()
	if cast.Targets[0] != game.PlayerTarget(game.Player2) {
		t.Fatalf("CastSpell() targets aliased caller slice: %+v", cast.Targets)
	}
	if cast.ChosenModes[0] != 1 {
		t.Fatalf("CastSpell() chosen modes aliased caller slice: %+v", cast.ChosenModes)
	}
}

func TestActivateAbilityPreservesTargets(t *testing.T) {
	sourceID := id.ID(42)
	targets := []game.Target{
		game.PlayerTarget(game.Player3),
		game.PermanentTarget(id.ID(100)),
	}
	modes := []int{1}
	targetCounts := []int{1, 1}

	got := ActivateAbilityWithModesAndTargetCounts(sourceID, 2, targets, targetCounts, 0, modes)
	if got.Kind != ActionActivateAbility {
		t.Fatalf("ActivateAbility() kind = %v, want %v", got.Kind, ActionActivateAbility)
	}
	activate, ok := got.ActivateAbilityPayload()
	if !ok {
		t.Fatal("ActivateAbilityPayload() ok = false, want true")
	}
	if activate.SourceID != sourceID {
		t.Fatalf("ActivateAbility() source ID = %v, want %v", activate.SourceID, sourceID)
	}
	if !slices.Equal(activate.Targets, targets) {
		t.Fatalf("ActivateAbility() targets = %+v, want %+v", activate.Targets, targets)
	}
	if !slices.Equal(activate.TargetCounts, targetCounts) {
		t.Fatalf("ActivateAbility() target counts = %+v, want %+v", activate.TargetCounts, targetCounts)
	}
	targets[0] = game.PlayerTarget(game.Player4)
	targetCounts[0] = 0
	modes[0] = 0
	activate, _ = got.ActivateAbilityPayload()
	if activate.Targets[0] != game.PlayerTarget(game.Player3) {
		t.Fatalf("ActivateAbility() targets aliased caller slice: %+v", activate.Targets)
	}
	if !slices.Equal(activate.TargetCounts, []int{1, 1}) {
		t.Fatalf("ActivateAbility() target counts aliased caller slice: %+v", activate.TargetCounts)
	}
	if !slices.Equal(activate.ChosenModes, []int{1}) {
		t.Fatalf("ActivateAbility() chosen modes aliased caller slice: %+v", activate.ChosenModes)
	}
}

func TestDeclareAttackersCopiesInputSlice(t *testing.T) {
	attackers := []game.AttackDeclaration{{
		Attacker: id.ID(42),
		Target:   game.AttackTarget{Player: game.Player2},
	}}

	got := DeclareAttackers(attackers)
	attackers[0].Attacker = id.ID(99)
	payload, ok := got.DeclareAttackersPayload()
	if !ok {
		t.Fatal("DeclareAttackersPayload() ok = false, want true")
	}

	if payload.Attackers[0].Attacker != id.ID(42) {
		t.Fatalf("DeclareAttackers() aliased caller slice: %+v", payload.Attackers)
	}
}

func TestDeclareBlockersCopiesInputSlice(t *testing.T) {
	blockers := []game.BlockDeclaration{{
		Blocker:  id.ID(42),
		Blocking: id.ID(99),
	}}

	got := DeclareBlockers(blockers)
	blockers[0].Blocker = id.ID(100)
	payload, ok := got.DeclareBlockersPayload()
	if !ok {
		t.Fatal("DeclareBlockersPayload() ok = false, want true")
	}

	if payload.Blockers[0].Blocker != id.ID(42) {
		t.Fatalf("DeclareBlockers() aliased caller slice: %+v", payload.Blockers)
	}
}

func TestConstructedActionsValidate(t *testing.T) {
	actions := []Action{
		Pass(),
		PlayLand(id.ID(1)),
		CastSpell(id.ID(2), nil, 0, nil),
		CastMutateSpell(id.ID(2), id.ID(8)),
		ActivateAbility(id.ID(3), 0, nil, 0),
		SuspendCard(id.ID(4)),
		DeclareAttackers([]game.AttackDeclaration{{Attacker: id.ID(5), Target: game.AttackTarget{Player: game.Player2}}}),
		DeclareBlockers([]game.BlockDeclaration{{Blocker: id.ID(6), Blocking: id.ID(7)}}),
	}

	for _, act := range actions {
		if err := act.Validate(); err != nil {
			t.Fatalf("%+v Validate() error: %v", act, err)
		}
	}
}

func TestValidateRejectsKindPayloadMismatch(t *testing.T) {
	act := Pass()
	act.castSpell = CastSpellAction{CardID: id.ID(42), SourceZone: zone.Hand}

	if err := act.Validate(); err == nil {
		t.Fatal("Validate() = nil for action with unrelated payload, want error")
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	tests := []Action{
		CastSpell(0, nil, 0, nil),
		ActivateAbility(0, 0, nil, 0),
		SuspendCard(0),
		DeclareAttackers([]game.AttackDeclaration{{Target: game.AttackTarget{Player: game.Player2}}}),
		DeclareBlockers([]game.BlockDeclaration{{Blocker: id.ID(1)}}),
	}

	for _, act := range tests {
		if err := act.Validate(); err == nil {
			t.Fatalf("%+v Validate() = nil, want error", act)
		}
	}
}

func TestRedactedFaceDownClearsHiddenIdentity(t *testing.T) {
	original := CastFaceDown(id.ID(42), game.FaceBack, game.FaceDownDisguise)
	redacted := original.Redacted()

	payload, ok := redacted.CastFaceDownPayload()
	if !ok {
		t.Fatal("redacted action is not a face-down cast")
	}
	if payload.CardID != 0 {
		t.Errorf("redacted CardID = %v, want 0", payload.CardID)
	}
	if payload.Face != game.FaceFront {
		t.Errorf("redacted Face = %v, want FaceFront", payload.Face)
	}
	if payload.FaceDownKind != game.FaceDownDisguise {
		t.Errorf("redacted FaceDownKind = %v, want FaceDownDisguise (public)", payload.FaceDownKind)
	}
	// The original action is not mutated by redaction.
	if originalPayload, _ := original.CastFaceDownPayload(); originalPayload.CardID != id.ID(42) {
		t.Error("Redacted mutated the original action")
	}
}

func TestRedactedPublicActionUnchanged(t *testing.T) {
	playLand := PlayLandFace(id.ID(7), game.FaceFront)
	redacted := playLand.Redacted()
	if redacted.Kind != ActionPlayLand {
		t.Fatalf("redacted Kind = %v, want ActionPlayLand", redacted.Kind)
	}
	payload, ok := redacted.PlayLandPayload()
	if !ok || payload.CardID != id.ID(7) {
		t.Errorf("redacted play-land payload = %+v ok=%v, want CardID 7 unchanged", payload, ok)
	}
}
