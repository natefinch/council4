package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCapturedConditionIsNotAutomaticMana(t *testing.T) {
	t.Parallel()
	instruction := game.Instruction{ConditionGate: "scope"}
	if unconditionalPaymentInstruction(&instruction) {
		t.Fatal("condition-gated mana must not be considered unconditional")
	}
}
