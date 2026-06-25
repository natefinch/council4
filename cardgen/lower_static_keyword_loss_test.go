package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerStaticPowerToughnessKeywordLoss proves a static declaration that
// composes a power/toughness modification with a keyword loss ("Equipped
// creature gets +10/+10 and loses flying.", Colossus Hammer) lowers to one
// static ability carrying two continuous effects: the LayerPowerToughnessModify
// buff over the attached object and the LayerAbility removal of the named
// keyword over the same group.
func TestLowerStaticPowerToughnessKeywordLoss(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		power      int
		toughness  int
		keywords   []game.Keyword
	}{
		"equipped creature loses flying": {
			typeLine:   "Artifact — Equipment",
			oracleText: "Equipped creature gets +10/+10 and loses flying.",
			power:      10,
			toughness:  10,
			keywords:   []game.Keyword{game.Flying},
		},
		"enchanted creature loses flying": {
			typeLine:   "Enchantment — Aura",
			oracleText: "Enchanted creature gets -6/-0 and loses flying.",
			power:      -6,
			toughness:  0,
			keywords:   []game.Keyword{game.Flying},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Loss",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			static, ok := keywordLossStaticAbility(face)
			if !ok {
				t.Fatalf("static abilities = %#v, want one with a keyword-loss effect", face.StaticAbilities)
			}
			effects := static.Body.ContinuousEffects
			if len(effects) != 2 {
				t.Fatalf("continuous effects = %#v, want 2", effects)
			}
			buff := effects[0]
			if buff.Layer != game.LayerPowerToughnessModify ||
				buff.Group.Domain() != game.GroupDomainAttachedObject ||
				buff.PowerDelta != test.power ||
				buff.ToughnessDelta != test.toughness {
				t.Fatalf("buff effect = %#v", buff)
			}
			loss := effects[1]
			if loss.Layer != game.LayerAbility ||
				loss.Group.Domain() != game.GroupDomainAttachedObject {
				t.Fatalf("loss effect = %#v", loss)
			}
			if len(loss.AddKeywords) != 0 {
				t.Fatalf("loss effect AddKeywords = %v, want none", loss.AddKeywords)
			}
			if !slices.Equal(loss.RemoveKeywords, test.keywords) {
				t.Fatalf("loss effect RemoveKeywords = %v, want %v", loss.RemoveKeywords, test.keywords)
			}
		})
	}
}

// keywordLossStaticAbility returns the lowered static ability whose continuous
// effects include a keyword removal, ignoring an Aura's enchant static ability.
func keywordLossStaticAbility(face loweredFaceAbilities) (loweredStaticAbility, bool) {
	for _, ability := range face.StaticAbilities {
		for _, effect := range ability.Body.ContinuousEffects {
			if len(effect.RemoveKeywords) > 0 {
				return ability, true
			}
		}
	}
	return loweredStaticAbility{}, false
}

// TestLowerStaticKeywordLossNonKeywordFailsClosed proves a "loses" clause that
// names a non-keyword payload ("loses all abilities") is not recognized as a
// keyword loss and fails closed rather than silently dropping the rules text.
func TestLowerStaticKeywordLossNonKeywordFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Loss Fail",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+2 and loses 3 life.",
	})
	if _, ok := keywordLossStaticAbility(face); ok {
		t.Fatalf("unexpected keyword-loss static ability: %#v", face.StaticAbilities)
	}
}
