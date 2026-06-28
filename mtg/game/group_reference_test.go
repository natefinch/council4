package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestObjectReferenceConstructorsAreValid(t *testing.T) {
	refs := []ObjectReference{
		TargetPermanentReference(0),
		SourcePermanentReference(),
		SourceCardPermanentReference(),
		SourceAttachedPermanentReference(),
		TargetAttachedPermanentReference(1),
		LinkedObjectReference("imprint"),
		EventPermanentReference(),
	}
	for _, ref := range refs {
		if problems := ref.Validate(); len(problems) != 0 {
			t.Errorf("ref %+v Validate() = %v, want none", ref, problems)
		}
	}
}

func TestObjectReferenceConstructorsSetExpectedFields(t *testing.T) {
	if got := SourceAttachedPermanentReference(); got.Kind() != ObjectReferenceSourceAttachedPermanent {
		t.Fatalf("SourceAttachedPermanentReference() = %+v", got)
	}
	if got := SourceCardPermanentReference(); got.Kind() != ObjectReferenceSourceCard {
		t.Fatalf("SourceCardPermanentReference() = %+v", got)
	}
	if got := TargetAttachedPermanentReference(2); got.Kind() != ObjectReferenceTargetAttachedPermanent || got.TargetIndex() != 2 {
		t.Fatalf("TargetAttachedPermanentReference(2) = %+v", got)
	}
	if got := LinkedObjectReference("x"); got.Kind() != ObjectReferenceLinkedObject || got.LinkID() != "x" {
		t.Fatalf("LinkedObjectReference(x) = %+v", got)
	}
}

func TestObjectReferenceValidateRejectsStructuralProblems(t *testing.T) {
	cases := map[string]ObjectReference{
		"zero value":                {},
		"target with link":          objectReferenceForTest(ObjectReferenceTargetPermanent, 0, "x"),
		"target negative index":     objectReferenceForTest(ObjectReferenceTargetPermanent, -1, ""),
		"source with index":         objectReferenceForTest(ObjectReferenceSourcePermanent, 3, ""),
		"source card with index":    objectReferenceForTest(ObjectReferenceSourceCard, 3, ""),
		"source attached with link": objectReferenceForTest(ObjectReferenceSourceAttachedPermanent, 0, "x"),
		"target attached with link": objectReferenceForTest(ObjectReferenceTargetAttachedPermanent, 0, "x"),
		"target attached negative":  objectReferenceForTest(ObjectReferenceTargetAttachedPermanent, -5, ""),
		"linked without link":       objectReferenceForTest(ObjectReferenceLinkedObject, 0, ""),
		"linked with index":         objectReferenceForTest(ObjectReferenceLinkedObject, 1, "x"),
		"event with link":           objectReferenceForTest(ObjectReferenceEventPermanent, 0, "x"),
		"unknown kind":              objectReferenceForTest(ObjectReferenceKind(99), 0, ""),
	}
	for name, ref := range cases {
		if problems := ref.Validate(); len(problems) == 0 {
			t.Errorf("%s: Validate() = none, want a problem", name)
		}
	}
}

func TestPlayerReferenceConstructorsAreValid(t *testing.T) {
	refs := []PlayerReference{
		ControllerReference(),
		TargetPlayerReference(0),
		ObjectControllerReference(SourcePermanentReference()),
		ObjectOwnerReference(TargetPermanentReference(0)),
	}
	for _, ref := range refs {
		if problems := ref.Validate(); len(problems) != 0 {
			t.Errorf("ref %+v Validate() = %v, want none", ref, problems)
		}
	}
}

func TestPlayerReferenceValidateRejectsStructuralProblems(t *testing.T) {
	cases := map[string]PlayerReference{
		"zero value":                       {},
		"controller with index":            playerReferenceForTest(PlayerReferenceController, 1, opt.V[ObjectReference]{}),
		"controller with object":           playerReferenceForTest(PlayerReferenceController, 0, opt.Val(SourcePermanentReference())),
		"target with object":               playerReferenceForTest(PlayerReferenceTargetPlayer, 0, opt.Val(SourcePermanentReference())),
		"target negative index":            playerReferenceForTest(PlayerReferenceTargetPlayer, -1, opt.V[ObjectReference]{}),
		"object controller without object": playerReferenceForTest(PlayerReferenceObjectController, 0, opt.V[ObjectReference]{}),
		"object owner without object":      playerReferenceForTest(PlayerReferenceObjectOwner, 0, opt.V[ObjectReference]{}),
		"object controller with index":     playerReferenceForTest(PlayerReferenceObjectController, 1, opt.Val(SourcePermanentReference())),
		"object owner with index":          playerReferenceForTest(PlayerReferenceObjectOwner, 2, opt.Val(SourcePermanentReference())),
		"object controller with invalid object": playerReferenceForTest(
			PlayerReferenceObjectController,
			0,
			opt.Val(objectReferenceForTest(ObjectReferenceTargetPermanent, 0, "x")),
		),
		"unknown kind": playerReferenceForTest(PlayerReferenceKind(99), 0, opt.V[ObjectReference]{}),
	}
	for name, ref := range cases {
		if problems := ref.Validate(); len(problems) == 0 {
			t.Errorf("%s: Validate() = none, want a problem", name)
		}
	}
}

func TestGroupReferenceConstructorsAreValid(t *testing.T) {
	creatures := Selection{RequiredTypes: []types.Card{types.Creature}}
	groups := []GroupReference{
		BattlefieldGroup(creatures),
		BattlefieldGroupExcluding(creatures, TargetPermanentReference(0)),
		AttachedObjectGroup(SourcePermanentReference()),
		ObjectControlledGroup(EventPermanentReference(), creatures),
		ObjectControlledGroupExcluding(EventPermanentReference(), creatures, EventPermanentReference()),
		PlayerControlledGroup(TargetPlayerReference(0), creatures),
		PlayerControlledGroupExcluding(TargetPlayerReference(0), creatures, SourcePermanentReference()),
	}
	for _, group := range groups {
		if !group.Valid() {
			t.Errorf("group %+v Valid() = false, want true: %v", group, group.Validate())
		}
	}
}

func TestGroupReferenceZeroValueInvalid(t *testing.T) {
	var group GroupReference
	if !group.Empty() {
		t.Fatal("zero-value GroupReference is not empty")
	}
	if group.Valid() {
		t.Fatal("zero GroupReference Valid() = true, want false")
	}
	if group.Domain() != groupDomainNone {
		t.Fatalf("zero GroupReference Domain() = %d, want %d", group.Domain(), groupDomainNone)
	}
}

func TestGroupReferenceValidateRejectsStructuralProblems(t *testing.T) {
	creatures := Selection{RequiredTypes: []types.Card{types.Creature}}
	cases := map[string]GroupReference{
		"battlefield with anchor": {
			domain: GroupDomainBattlefield,
			anchor: opt.Val(SourcePermanentReference()),
		},
		"attached without anchor": {domain: GroupDomainAttachedObject},
		"attached with selection": {
			domain:    GroupDomainAttachedObject,
			anchor:    opt.Val(SourcePermanentReference()),
			selection: creatures,
		},
		"attached with exclusion": {
			domain:  GroupDomainAttachedObject,
			anchor:  opt.Val(SourcePermanentReference()),
			exclude: opt.Val(SourcePermanentReference()),
		},
		"object controlled without anchor": {domain: GroupDomainObjectControlled, selection: creatures},
		"player controlled without player": {domain: GroupDomainPlayerControlled, selection: creatures},
		"player controlled with object anchor": {
			domain:       GroupDomainPlayerControlled,
			selection:    creatures,
			playerAnchor: opt.Val(TargetPlayerReference(0)),
			anchor:       opt.Val(SourcePermanentReference()),
		},
		"player anchor on battlefield group": {
			domain:       GroupDomainBattlefield,
			selection:    creatures,
			playerAnchor: opt.Val(TargetPlayerReference(0)),
		},
		"bad player anchor reference": {
			domain:       GroupDomainPlayerControlled,
			selection:    creatures,
			playerAnchor: opt.Val(playerReferenceForTest(PlayerReferenceKind(99), 0, opt.V[ObjectReference]{})),
		},
		"bad anchor reference": {
			domain: GroupDomainObjectControlled,
			anchor: opt.Val(objectReferenceForTest(ObjectReferenceLinkedObject, 0, "")),
		},
	}
	for name, group := range cases {
		if group.Valid() {
			t.Errorf("%s: Valid() = true, want false", name)
		}
	}
}

func TestGroupReferenceAccessors(t *testing.T) {
	creatures := Selection{RequiredTypes: []types.Card{types.Creature}}
	group := ObjectControlledGroupExcluding(EventPermanentReference(), creatures, TargetPermanentReference(1))
	if group.Domain() != GroupDomainObjectControlled {
		t.Fatalf("Domain() = %d", group.Domain())
	}
	if anchor, ok := group.Anchor(); !ok || anchor.Kind() != ObjectReferenceEventPermanent {
		t.Fatalf("Anchor() = %+v, %v", anchor, ok)
	}
	if exclude, ok := group.Exclusion(); !ok || exclude.Kind() != ObjectReferenceTargetPermanent || exclude.TargetIndex() != 1 {
		t.Fatalf("Exclusion() = %+v, %v", exclude, ok)
	}
	if group.Selection().Empty() {
		t.Fatal("Selection() unexpectedly empty")
	}
}

func TestPlayerControlledGroupAccessors(t *testing.T) {
	creatures := Selection{RequiredTypes: []types.Card{types.Creature}}
	group := PlayerControlledGroupExcluding(TargetPlayerReference(0), creatures, SourcePermanentReference())
	if group.Domain() != GroupDomainPlayerControlled {
		t.Fatalf("Domain() = %d, want GroupDomainPlayerControlled", group.Domain())
	}
	player, ok := group.PlayerAnchor()
	if !ok || player.Kind() != PlayerReferenceTargetPlayer || player.TargetIndex() != 0 {
		t.Fatalf("PlayerAnchor() = %+v, %v", player, ok)
	}
	if _, ok := group.Anchor(); ok {
		t.Fatal("Anchor() = ok, want no object anchor for a player-controlled group")
	}
	if exclude, ok := group.Exclusion(); !ok || exclude.Kind() != ObjectReferenceSourcePermanent {
		t.Fatalf("Exclusion() = %+v, %v", exclude, ok)
	}
}

func TestCommonContinuousEffectGroupsAreValid(t *testing.T) {
	cases := []struct {
		name   string
		group  GroupReference
		domain GroupReferenceDomain
	}{
		{
			name:   "all creatures",
			group:  BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Creature}}),
			domain: GroupDomainBattlefield,
		},
		{
			name:   "all creatures except target",
			group:  BattlefieldGroupExcluding(Selection{RequiredTypes: []types.Card{types.Creature}}, TargetPermanentReference(0)),
			domain: GroupDomainBattlefield,
		},
		{
			name:   "equipped creature",
			group:  AttachedObjectGroup(SourcePermanentReference()),
			domain: GroupDomainAttachedObject,
		},
		{
			name:   "other creatures defending player controls",
			group:  ObjectControlledGroupExcluding(EventPermanentReference(), Selection{RequiredTypes: []types.Card{types.Creature}}, EventPermanentReference()),
			domain: GroupDomainObjectControlled,
		},
	}
	for _, tc := range cases {
		if !tc.group.Valid() {
			t.Fatalf("%s invalid: %v", tc.name, tc.group.Validate())
		}
		if tc.group.Domain() != tc.domain {
			t.Errorf("%s domain = %d, want %d", tc.name, tc.group.Domain(), tc.domain)
		}
	}
}

func TestAttachedPermanentReferencesStayDistinct(t *testing.T) {
	source := SourceAttachedPermanentReference()
	target := TargetAttachedPermanentReference(2)
	if source.Kind() != ObjectReferenceSourceAttachedPermanent {
		t.Fatalf("source attached kind = %d", source.Kind())
	}
	if target.Kind() != ObjectReferenceTargetAttachedPermanent || target.TargetIndex() != 2 {
		t.Fatalf("target attached reference = %+v", target)
	}
}
