package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceRoomAbilityTriggerMultiplier(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		card  *ScryfallCard
		wants []string
	}{
		// Dungeon Delver grants the room-ability doubler to the commander creatures
		// its controller owns: the generated source carries a continuous ability
		// grant whose added ability is a static body with the room-ability rule
		// effect, filtered to commander creatures the owner owns regardless of who
		// controls them (Owner: game.OwnerYou over a battlefield-wide group).
		"dungeon delver grants to commander creatures": {
			card: &ScryfallCard{
				Name:       "Dungeon Delver",
				Layout:     "normal",
				ManaCost:   "{1}{B}",
				TypeLine:   "Legendary Enchantment — Background",
				OracleText: "Commander creatures you own have \"Room abilities of dungeons you own trigger an additional time.\"",
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForRoomAbility",
				"MatchCommander: true",
				"Owner: game.OwnerYou",
				"game.BattlefieldGroup",
				"AddAbilities",
			},
		},
		// The printed form (Hama Pashar, Ruin Seeker) lowers the same rule effect as
		// a printed static ability on the source itself, without a grant.
		"printed room ability doubler": {
			card: &ScryfallCard{
				Name:       "Ruin Seeker",
				Layout:     "normal",
				TypeLine:   "Legendary Enchantment",
				OracleText: "Room abilities of dungeons you own trigger an additional time.",
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForRoomAbility",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(tc.card, "r")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			normalized := strings.Join(strings.Fields(source), " ")
			for _, wanted := range tc.wants {
				if !strings.Contains(normalized, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func TestGenerateExecutableCardSourceRoomAbilityMultiplierPrintedNotAGrant(t *testing.T) {
	t.Parallel()
	// The printed form must not emit a continuous ability grant: the rule effect
	// belongs to the source's own printed static ability, not to a granted body.
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Ruin Seeker",
		Layout:     "normal",
		TypeLine:   "Legendary Enchantment",
		OracleText: "Room abilities of dungeons you own trigger an additional time.",
	}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "MatchCommander") || strings.Contains(source, "AddAbilities") {
		t.Fatalf("printed room-ability doubler unexpectedly emitted a grant:\n%s", source)
	}
}
