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
