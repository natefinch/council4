package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSearchCastingPermissionRequiresLibraryFunction(t *testing.T) {
	for _, source := range []zone.Type{zone.None, zone.Hand, zone.Library} {
		def := &CardDef{CardFace: CardFace{
			Name: "Search permission", Types: []types.Card{types.Creature},
			StaticAbilities: []StaticAbility{{ZoneOfFunction: source, CastWhileSearchingOwnLibrary: true}},
		}}
		if valid := len(ValidateCardDef(def)) == 0; valid != (source == zone.Library) {
			t.Fatalf("permission validation in %v = %v", source, valid)
		}
	}
}

func TestSearchAndPaymentScopesCloneAndSerialize(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.SearchControlScopes = []SearchControlScope{{Player: Player2, Controller: Player3, Override: true, BeforeFind: true}}
	g.ManaPaymentScopes = []PlayerID{Player2}
	clone := g.Clone()
	clone.SearchControlScopes[0].BeforeFind = false
	clone.ManaPaymentScopes[0] = Player3
	if !g.SearchCastingAllowed(Player2) || g.SearchCastingAllowed(Player3) ||
		!g.ManaPaymentAllowed(Player2) || g.ManaPaymentAllowed(Player3) {
		t.Fatal("clone shared search/payment metadata or changed its actor")
	}
	encoded, err := json.Marshal(g.SearchControlScopes)
	if err != nil {
		t.Fatal(err)
	}
	var restored []SearchControlScope
	if err := json.Unmarshal(encoded, &restored); err != nil || !reflect.DeepEqual(restored, g.SearchControlScopes) {
		t.Fatal("before-find metadata did not serialize by value")
	}
	g.SearchControlScopes = append(g.SearchControlScopes, SearchControlScope{Player: Player2})
	if g.SearchCastingAllowed(Player2) {
		t.Fatal("closed inner search reused the outer search's permission")
	}
	g.Players[Player2].Eliminated = true
	g.SearchControlScopes = g.SearchControlScopes[:1]
	if g.SearchCastingAllowed(Player2) {
		t.Fatal("departed actor retained search casting")
	}
}
