package game

import "testing"

func TestDamageRecipientFullNameAccessors(t *testing.T) {
	objectRecipient := ObjectDamageRecipient(TargetPermanentReference(0))
	if object, ok := objectRecipient.ObjectReference(); !ok || object.Kind() != ObjectReferenceTargetPermanent {
		t.Fatalf("ObjectReference() = %+v, %v", object, ok)
	}

	playerRecipient := PlayerDamageRecipient(TargetPlayerReference(0))
	if player, ok := playerRecipient.PlayerReference(); !ok || player.Kind() != PlayerReferenceTargetPlayer {
		t.Fatalf("PlayerReference() = %+v, %v", player, ok)
	}

	group := BattlefieldGroup(Selection{})
	groupRecipient := GroupDamageRecipient(group)
	if got, ok := groupRecipient.GroupReference(); !ok || got.Domain() != GroupDomainBattlefield {
		t.Fatalf("GroupReference() = %+v, %v", got, ok)
	}

	playerGroupRecipient := PlayerGroupDamageRecipient(OpponentsReference())
	if got, ok := playerGroupRecipient.PlayerGroupReference(); !ok || got.Kind != PlayerGroupReferenceOpponents {
		t.Fatalf("PlayerGroupReference() = %+v, %v", got, ok)
	}

	anyTarget := AnyTargetDamageRecipient(1)
	if object, ok := anyTarget.AnyTargetObjectReference(); !ok || object.TargetIndex() != 1 {
		t.Fatalf("AnyTargetObjectReference() = %+v, %v", object, ok)
	}
	if player, ok := anyTarget.AnyTargetPlayerReference(); !ok || player.TargetIndex() != 1 {
		t.Fatalf("AnyTargetPlayerReference() = %+v, %v", player, ok)
	}
}
