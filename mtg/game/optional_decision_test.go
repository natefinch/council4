package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/opt"
)

func TestValidateOptionalDecisionEnvelope(t *testing.T) {
	publisher := Instruction{
		Primitive: Draw{Player: ControllerReference(), Amount: Fixed(1)},
		Optional:  true, PublishOptionalDecision: "group",
	}
	consumer := Instruction{
		Primitive:            GainLife{Player: ControllerReference(), Amount: Fixed(2)},
		OptionalDecisionGate: "group",
	}
	for _, tt := range []struct {
		name   string
		change func(*Instruction)
		reason string
	}{
		{"valid", func(*Instruction) {}, ""},
		{"not optional", func(i *Instruction) { i.Optional = false }, "requires Optional"},
		{"group offer", func(i *Instruction) { i.OptionalActorGroup = opt.Val(AllPlayersReference()) }, "one deciding player"},
		{"tempting offer", func(i *Instruction) { i.TemptingOffer = true }, "one deciding player"},
		{"for each", func(i *Instruction) { i.ForEachPlayerGroup = opt.Val(AllPlayersReference()) }, "one deciding player"},
		{"self forward gate", func(i *Instruction) { i.OptionalDecisionGate = "group" }, "not yet published"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			first := publisher
			tt.change(&first)
			err := ValidateInstructionSequence([]Instruction{first, consumer})
			if tt.reason == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.reason) {
				t.Fatalf("error=%v, want %q", err, tt.reason)
			}
		})
	}
	for _, tt := range []struct {
		name     string
		sequence []Instruction
		reason   string
	}{
		{"missing", []Instruction{consumer}, "not yet published"},
		{"forward", []Instruction{consumer, publisher}, "not yet published"},
		{"duplicate", []Instruction{publisher, publisher}, "duplicate optional decision"},
		{"direct dispatch body", []Instruction{{Primitive: Draw{Player: ControllerReference(), Amount: Fixed(1)},
			TemptingOfferBody: []Instruction{publisher, consumer}}}, "inside TemptingOfferBody"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstructionSequence(tt.sequence)
			if err == nil || !strings.Contains(err.Error(), tt.reason) {
				t.Fatalf("error=%v, want %q", err, tt.reason)
			}
		})
	}
}
