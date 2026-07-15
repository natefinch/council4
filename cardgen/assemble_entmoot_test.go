package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerAssembleTheEntmoot proves the full card lowers: the static "Creatures
// you control have reach" grant, the sacrifice activation cost, and the ordered
// activated body that creates three tapped X/X (life-gained-this-turn sized)
// Treefolk tokens published under a link group, then puts a reach counter on
// each of them via that group.
func TestLowerAssembleTheEntmoot(t *testing.T) {
	t.Parallel()
	_, diagnostics := compileTestOracle(
		"Sacrifice this enchantment: Create three tapped X/X green Treefolk creature tokens, where X is the amount of life you gained this turn. Put a reach counter on each of them.",
		parser.Context{}, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Assemble the Entmoot",
		Layout:     "normal",
		ManaCost:   "{3}{G}",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control have reach.\nSacrifice this enchantment: Create three tapped X/X green Treefolk creature tokens, where X is the amount of life you gained this turn. Put a reach counter on each of them.",
	})

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want the reach grant", face.StaticAbilities)
	}
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one sacrifice ability", face.ActivatedAbilities)
	}
	activated := face.ActivatedAbilities[0]

	sacrifices := false
	for _, additional := range activated.AdditionalCosts {
		if additional.Kind == cost.AdditionalSacrificeSource {
			sacrifices = true
		}
	}
	if !sacrifices {
		t.Fatalf("activation costs = %#v, want a sacrifice-source additional cost", activated.AdditionalCosts)
	}

	sequence := activated.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v", sequence)
	}
	create, ok := sequence[0].Primitive.(game.CreateToken)
	if !ok || create.Amount.Value() != 3 || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want three tokens publishing a link", sequence[0].Primitive)
	}
	if !create.EntryTapped {
		t.Fatalf("create = %#v, want tapped entry", create)
	}
	if !create.Power.Exists || !create.Power.Val.IsDynamic() {
		t.Fatalf("create power = %#v, want dynamic size", create.Power)
	}
	power := create.Power.Val.DynamicAmount()
	if !power.Exists || power.Val.Kind != game.DynamicAmountLifeGainedThisTurn {
		t.Fatalf("create power dynamic = %#v, want life gained this turn", power)
	}
	if !create.Toughness.Exists || !create.Toughness.Val.IsDynamic() {
		t.Fatalf("create toughness = %#v, want dynamic size", create.Toughness)
	}

	add, ok := sequence[1].Primitive.(game.AddCounter)
	key, linked := add.Group.LinkedKey()
	if !ok || !linked || key != create.PublishLinked ||
		add.Object.Kind() != game.ObjectReferenceNone ||
		add.CounterKind != counter.Reach ||
		add.Amount.Value() != 1 {
		t.Fatalf("counter placement = %#v, want one reach counter on the linked token group", sequence[1].Primitive)
	}
}
