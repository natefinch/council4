package cardprogram

import (
	"encoding/json"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestReplacementProgramCatalogValidationAndDataRoundtrip(t *testing.T) {
	t.Parallel()
	ref := game.ProgramReference{ID: "test/replacement", Version: "1"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Replacement", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{{
			Replacement: game.ReplacementEffect{
				MatchEvent: game.EventCardDrawn, ControllerFilter: game.TriggerControllerYou,
				Duration: game.DurationPermanent, Program: ref, Optional: true,
			},
		}},
	}}
	catalog, err := NewCatalog([]Program{{Reference: ref, Resolve: resolveNoop}})
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateCard(def); err != nil {
		t.Fatal(err)
	}
	if err := (Catalog{}).ValidateCard(def); err == nil {
		t.Fatal("replacement-only card bypassed missing program validation")
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	var restored game.CardDef
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateCard(&restored); err != nil {
		t.Fatal(err)
	}
	replacement := &restored.ReplacementAbilities[0].Replacement
	if replacement.Program != ref || !replacement.Optional ||
		replacement.ControllerFilter != game.TriggerControllerYou {
		t.Fatal("replacement declaration was not portable data")
	}
	replacement.Program.Version = "2"
	if err := catalog.ValidateCard(&restored); err == nil {
		t.Fatal("replacement version mismatch was accepted")
	}
	replacement.Program = ref
	replacement.DrawCardDigLook = 3
	if err := catalog.ValidateCard(&restored); err == nil {
		t.Fatal("mixed replacement actions were silently accepted")
	}
}
