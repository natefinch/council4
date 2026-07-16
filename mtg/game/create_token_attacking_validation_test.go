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
