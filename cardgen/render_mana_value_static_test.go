package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRendererPreservesManaValueBaseCharacteristicData(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Mana-value static",
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerPowerToughnessSet, AffectedSource: true,
				SetPowerToManaValue: true, SetToughnessToManaValue: true,
			}},
		}},
	}}
	source, err := (Renderer{}).RenderCardSource(&ScryfallCard{Name: def.Name, Layout: "normal"}, []*game.CardDef{def}, nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	source = strings.Join(strings.Fields(source), " ")
	for _, fragment := range []string{"SetPowerToManaValue: true", "SetToughnessToManaValue: true"} {
		if !strings.Contains(source, fragment) {
			t.Fatalf("typed setter %q was lost:\n%s", fragment, source)
		}
	}
}
