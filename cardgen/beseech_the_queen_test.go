package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func beseechTheQueenCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Beseech the Queen",
		Layout:   "normal",
		ManaCost: "{2/B}{2/B}{2/B}",
		TypeLine: "Sorcery",
		OracleText: "({2/B} can be paid with any two mana or with {B}. This card's mana value is 6.)\n" +
			"Search your library for a card with mana value less than or equal to the number of lands you control, reveal it, put it into your hand, then shuffle.",
	}
}

// TestGenerateExecutableCardSourceBeseechTheQueen proves Beseech the Queen
// generates end to end: the monocolored-hybrid {2/B} cost lowers to three
// cost.Twobrid pips, and the dynamic-count mana-value bound ("with mana value
// less than or equal to the number of lands you control") renders as a
// Selection.ManaValueDynamic count over the lands the controller controls.
func TestGenerateExecutableCardSourceBeseechTheQueen(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(beseechTheQueenCard(), "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"cost.Twobrid(mana.B),\n\t\t\t\tcost.Twobrid(mana.B),\n\t\t\t\tcost.Twobrid(mana.B),",
		"game.Search{",
		"SourceZone:  zone.Library,",
		"Destination: zone.Hand,",
		"Filter:      game.Selection{ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{Kind: game.DynamicAmountCountSelector, Multiplier: 1, Group: game.GroupRef(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}))})},",
		"Reveal:      true,",
		"Amount: game.Fixed(1),",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerBeseechTheQueenSearchFilter proves the search lowers to a library
// search whose filter bounds the searched card's mana value by the dynamic count
// of lands the controller controls. The bound is the count-selector kind over a
// you-controlled battlefield land group with multiplier one, and the filtered
// search keeps the default fail-to-find policy so it may legally find nothing.
func TestLowerBeseechTheQueenSearchFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, beseechTheQueenCard())
	if !face.SpellAbility.Exists {
		t.Fatal("Beseech the Queen produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("spell modes = %#v, want one single-instruction mode", modes)
	}
	search, ok := modes[0].Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Search", modes[0].Sequence[0].Primitive)
	}
	if search.Spec.SourceZone != zone.Library || search.Spec.Destination != zone.Hand {
		t.Fatalf("search zones = %v -> %v, want library -> hand", search.Spec.SourceZone, search.Spec.Destination)
	}
	if !search.Spec.Reveal {
		t.Fatal("search does not reveal the found card")
	}
	if search.Spec.FailToFindPolicy != game.SearchFailToFindDefault {
		t.Fatalf("fail-to-find policy = %v, want default (a filtered search may fail to find)", search.Spec.FailToFindPolicy)
	}
	bound := search.Spec.Filter.ManaValueDynamic
	if !bound.Exists {
		t.Fatal("search filter has no dynamic mana-value bound")
	}
	if bound.Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("bound kind = %v, want DynamicAmountCountSelector", bound.Val.Kind)
	}
	if bound.Val.Multiplier != 1 || bound.Val.Addend != 0 {
		t.Fatalf("bound scaling = %+v, want multiplier 1 addend 0", bound.Val)
	}
	if bound.Val.Group == nil {
		t.Fatal("count bound has no group")
	}
	group := *bound.Val.Group
	if group.Domain() != game.GroupDomainBattlefield {
		t.Fatalf("group domain = %v, want battlefield", group.Domain())
	}
	sel := group.Selection()
	if len(sel.RequiredTypes) != 1 || sel.RequiredTypes[0] != types.Land {
		t.Fatalf("group required types = %v, want [land]", sel.RequiredTypes)
	}
	if sel.Controller != game.ControllerYou {
		t.Fatalf("group controller = %v, want you", sel.Controller)
	}
}
