package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerRabbleRousingAttackerCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Rabble Rousing",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{4}{W}",
		OracleText: "Hideaway 5 (When this enchantment enters, look at the top five cards of your library, exile one face down, then put the rest on the bottom in a random order.)\nWhenever you attack with one or more creatures, create that many 1/1 green and white Citizen creature tokens. Then if you control ten or more creatures, you may play the exiled card without paying its mana cost.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want hideaway and attack", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[1].Content.Modes[0]
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	dynamic := create.Amount.DynamicAmount()
	if !ok ||
		!dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountTriggeringAttackerCount ||
		dynamic.Val.Selection == nil {
		t.Fatalf("create = %#v", mode.Sequence[0].Primitive)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.PlayHideawayCard); !ok ||
		!mode.Sequence[1].Optional ||
		!mode.Sequence[1].Condition.Exists {
		t.Fatalf("hideaway play = %#v", mode.Sequence[1])
	}
}

func TestLowerPoeticIngenuityCountsMatchingAttackers(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Poetic Ingenuity",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{R}",
		OracleText: "Whenever one or more Dinosaurs you control attack, create that many Treasure tokens.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	create, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	dynamic := create.Amount.DynamicAmount()
	if !ok ||
		!dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountTriggeringAttackerCount ||
		dynamic.Val.Selection == nil ||
		len(dynamic.Val.Selection.SubtypesAny) != 1 ||
		dynamic.Val.Selection.SubtypesAny[0] != types.Dinosaur {
		t.Fatalf("create = %#v", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerResolvingHideawayPlayEffect(t *testing.T) {
	t.Parallel()
	const source = "Create a Treasure token. Then if you control ten or more creatures, you may play the exiled card without paying its mana cost."
	compilation, diagnostics := compileTestOracle(
		source,
		parser.Context{CardName: "Test Hideaway"},
		compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[1]
	content, diagnostic := lowerHideawayPlayEffect(contentCtx{content: compiler.AbilityContent{
		Effects:    []compiler.CompiledEffect{effect},
		References: effect.References,
	}})
	if diagnostic != nil {
		t.Fatalf("diagnostic = %#v, effect = %#v", diagnostic, effect)
	}
	if len(content.Modes) != 1 ||
		len(content.Modes[0].Sequence) != 1 ||
		!content.Modes[0].Sequence[0].Optional {
		t.Fatalf("content = %#v", content)
	}
	if _, ok := content.Modes[0].Sequence[0].Primitive.(game.PlayHideawayCard); !ok {
		t.Fatalf("primitive = %#v", content.Modes[0].Sequence[0].Primitive)
	}
}
