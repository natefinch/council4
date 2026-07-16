package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestSaddleContributorsGroupValidation(t *testing.T) {
	valid := SaddleContributorsGroup(
		SourcePermanentReference(),
		Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSupertype: types.Legendary},
	)
	if problems := valid.Validate(); len(problems) != 0 {
		t.Fatalf("valid Saddle contributor group problems = %v", problems)
	}
	if valid.Domain() != GroupDomainSaddleContributors {
		t.Fatalf("domain = %v, want Saddle contributors", valid.Domain())
	}
	if _, ok := valid.Anchor(); !ok {
		t.Fatal("Saddle contributor group lost its source anchor")
	}

	invalid := GroupReference{domain: GroupDomainSaddleContributors}
	if problems := invalid.Validate(); len(problems) == 0 {
		t.Fatal("anchorless Saddle contributor group unexpectedly validated")
	}
}

func TestChosenGroupTokenCopyValidationAndClone(t *testing.T) {
	group := SaddleContributorsGroup(
		SourcePermanentReference(),
		Selection{RequiredTypes: []types.Card{types.Creature}},
	)
	create := CreateToken{
		Amount: Fixed(1),
		Source: TokenCopyOf(TokenCopySpec{
			Source:      TokenCopySourceChosenFromGroup,
			Group:       GroupRef(group),
			SetColors:   []color.Color{color.Red},
			SetTypes:    []types.Card{types.Creature},
			SetSubtypes: []types.Sub{types.Human},
			AddColors:   []color.Color{color.Green},
			AddTypes:    []types.Card{types.Artifact},
			AddSubtypes: []types.Sub{types.Goblin},
			AddKeywords: []Keyword{Haste},
		}),
	}

	if err := create.validatePrimitive(nil, true); err != nil {
		t.Fatalf("valid chosen-group token copy: %v", err)
	}

	content := Mode{Sequence: []Instruction{{Primitive: RepeatProcess{
		Times: Fixed(2),
		Body:  Mode{Sequence: []Instruction{{Primitive: create}}}.Ability(),
	}}}}.Ability()
	cloned := cloneAbilityContent(content)
	originalRepeat, ok := content.Modes[0].Sequence[0].Primitive.(RepeatProcess)
	if !ok {
		t.Fatalf("original primitive = %T, want RepeatProcess", content.Modes[0].Sequence[0].Primitive)
	}
	clonedRepeat, ok := cloned.Modes[0].Sequence[0].Primitive.(RepeatProcess)
	if !ok {
		t.Fatalf("cloned primitive = %T, want RepeatProcess", cloned.Modes[0].Sequence[0].Primitive)
	}
	originalCreate, ok := originalRepeat.Body.Modes[0].Sequence[0].Primitive.(CreateToken)
	if !ok {
		t.Fatalf("original repeat primitive = %T, want CreateToken", originalRepeat.Body.Modes[0].Sequence[0].Primitive)
	}
	clonedCreate, ok := clonedRepeat.Body.Modes[0].Sequence[0].Primitive.(CreateToken)
	if !ok {
		t.Fatalf("cloned repeat primitive = %T, want CreateToken", clonedRepeat.Body.Modes[0].Sequence[0].Primitive)
	}
	originalSpec, ok := originalCreate.Source.TokenCopy()
	if !ok {
		t.Fatal("original CreateToken source is not a copy")
	}
	clonedSpec, ok := clonedCreate.Source.TokenCopy()
	if !ok {
		t.Fatal("cloned CreateToken source is not a copy")
	}
	if originalSpec.Group == clonedSpec.Group {
		t.Fatal("chosen-group TokenCopySpec shared its GroupReference pointer across clone")
	}
	*originalSpec.Group = BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Land}})
	if clonedSpec.Group.Domain() != GroupDomainSaddleContributors {
		t.Fatalf("mutating original group changed clone domain to %v", clonedSpec.Group.Domain())
	}
	originalSpec.SetColors[0] = color.Blue
	originalSpec.SetTypes[0] = types.Land
	originalSpec.SetSubtypes[0] = types.Goblin
	originalSpec.AddColors[0] = color.White
	originalSpec.AddTypes[0] = types.Enchantment
	originalSpec.AddSubtypes[0] = types.Human
	originalSpec.AddKeywords[0] = Flying
	if clonedSpec.SetColors[0] != color.Red ||
		clonedSpec.SetTypes[0] != types.Creature ||
		clonedSpec.SetSubtypes[0] != types.Human ||
		clonedSpec.AddColors[0] != color.Green ||
		clonedSpec.AddTypes[0] != types.Artifact ||
		clonedSpec.AddSubtypes[0] != types.Goblin ||
		clonedSpec.AddKeywords[0] != Haste {
		t.Fatal("mutating original TokenCopySpec modifier slices changed clone")
	}

	invalid := CreateToken{
		Amount: Fixed(1),
		Source: TokenCopyOf(TokenCopySpec{Source: TokenCopySourceChosenFromGroup}),
	}
	if err := invalid.validatePrimitive(nil, true); err == nil {
		t.Fatal("chosen-group token copy without a group unexpectedly validated")
	}
}

func TestSaddleIdentitySlicesAreDeepCloned(t *testing.T) {
	permanent := &Permanent{SaddleContributorIDs: []id.ID{1, 2}}
	permanentClone := clonePermanent(permanent)
	permanent.SaddleContributorIDs[0] = 9
	if permanentClone.SaddleContributorIDs[0] != 1 {
		t.Fatalf("permanent clone contributors = %v", permanentClone.SaddleContributorIDs)
	}

	stackObject := &StackObject{
		TappedAsCostIDs:   []id.ID{3, 4},
		CapturedObjectIDs: []id.ID{7, 8},
	}
	stackClone := cloneStackObject(stackObject)
	stackObject.TappedAsCostIDs[0] = 9
	stackObject.CapturedObjectIDs[0] = 9
	if stackClone.TappedAsCostIDs[0] != 3 {
		t.Fatalf("stack clone tapped IDs = %v", stackClone.TappedAsCostIDs)
	}
	if stackClone.CapturedObjectIDs[0] != 7 {
		t.Fatalf("stack clone captured IDs = %v", stackClone.CapturedObjectIDs)
	}

	delayed := []DelayedTrigger{{CapturedObjectIDs: []id.ID{10, 11}}}
	delayedClone := cloneDelayedTriggers(delayed)
	delayed[0].CapturedObjectIDs[0] = 9
	if delayedClone[0].CapturedObjectIDs[0] != 10 {
		t.Fatalf("delayed trigger clone captured IDs = %v", delayedClone[0].CapturedObjectIDs)
	}

	snapshot := ObjectSnapshot{SaddleContributorIDs: []id.ID{5, 6}}
	snapshotClone := cloneObjectSnapshot(snapshot)
	snapshot.SaddleContributorIDs[0] = 9
	if snapshotClone.SaddleContributorIDs[0] != 5 {
		t.Fatalf("snapshot clone contributors = %v", snapshotClone.SaddleContributorIDs)
	}
}
