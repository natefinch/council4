package game

import (
	"testing"

	"github.com/natefinch/council4/opt"
)

func TestCreateTokenExplicitAttackingDefenderValidation(t *testing.T) {
	t.Parallel()
	create := CreateToken{
		Amount:                 Fixed(1),
		Source:                 TokenDef(&CardDef{CardFace: CardFace{Name: "Test Token"}}),
		EntryAttackingDefender: opt.Val(DefendingPlayerReference()),
	}
	if err := create.validatePrimitive(nil, true); err != nil {
		t.Fatalf("explicit attacking defender rejected: %v", err)
	}
	create.EntryAttacking = true
	if err := create.validatePrimitive(nil, true); err == nil {
		t.Fatal("generic and explicit attacking entry modes were both accepted")
	}
}

func TestCreateTokenAttackSameAsObjectValidation(t *testing.T) {
	t.Parallel()
	create := CreateToken{
		Amount:             Fixed(1),
		Source:             TokenDef(&CardDef{CardFace: CardFace{Name: "Test Token"}}),
		AttackSameAsObject: opt.Val(TargetPermanentReference(0)),
	}
	targets := []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent}}
	if err := create.validatePrimitive(targets, true); err != nil {
		t.Fatalf("same-as-object attacking entry rejected: %v", err)
	}
	if err := create.validatePrimitive(nil, true); err == nil {
		t.Fatal("same-as-object attacking entry accepted without its target")
	}
	create.AttackSameAsSource = true
	if err := create.validatePrimitive(targets, true); err == nil {
		t.Fatal("source and object-correlated attacking entry modes were both accepted")
	}
}

func TestCreateTokenEntryBlockingValidation(t *testing.T) {
	t.Parallel()
	create := CreateToken{
		Amount:        Fixed(1),
		Source:        TokenDef(&CardDef{CardFace: CardFace{Name: "Test Token"}}),
		EntryBlocking: opt.Val(TargetPermanentReference(0)),
	}
	targets := []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent}}
	if err := create.validatePrimitive(targets, true); err != nil {
		t.Fatalf("blocking entry rejected: %v", err)
	}
	if err := create.validatePrimitive(nil, true); err == nil {
		t.Fatal("blocking entry accepted without its target")
	}
	attacking := create
	attacking.EntryAttacking = true
	if err := attacking.validatePrimitive(targets, true); err == nil {
		t.Fatal("blocking and attacking entry were both accepted")
	}
	copySource := CreateToken{Amount: Fixed(1), Source: create.Source, EntryBlockingCopySource: true}
	if err := copySource.validatePrimitive(nil, true); err == nil {
		t.Fatal("copy-source blocking entry accepted without a copy of each group member")
	}
	copySource.Source = TokenCopyOf(TokenCopySpec{Source: TokenCopySourceEachInGroup, Group: GroupRef(BattlefieldGroup(Selection{CombatState: CombatStateAttacking}))})
	if err := copySource.validatePrimitive(nil, true); err != nil {
		t.Fatalf("copy-source blocking entry rejected: %v", err)
	}
}
