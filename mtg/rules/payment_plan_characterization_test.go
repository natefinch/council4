package rules

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestSpellPaymentPlanCharacterization(t *testing.T) {
	// This pins current planner behavior before the Phase 2 cost-planner
	// decomposition. Keep the coverage when the overload-chain entry point is
	// collapsed; update only the call seam, not the behavior being summarized.
	tests := []struct {
		name       string
		setup      func() (*game.Game, *game.CardDef, id.ID, int)
		sourceZone game.ZoneType // defaults to ZoneHand when zero
		kickerPaid bool
		want       []string
	}{
		{
			name: "convoke pays colored symbols before generic symbols",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player1, namedCreature("Green Convoke Creature", mana.Green))
				addCombatPermanent(g, game.Player1, namedCreature("Plain Convoke Creature"))
				return g, convokeSpell(mana.Cost{mana.ColoredMana(mana.Green), mana.GenericMana(1)}), 0, 0
			},
			want: []string{
				"option=0:Normal cost",
				"manaTaps=[]",
				"convoke=[Green Convoke Creature, Plain Convoke Creature]",
				"delve=[]",
				"additional=[]",
				"sacrifices=[]",
				"life=0",
			},
		},
		{
			name: "delve exiles from graveyard top first",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCardToGraveyard(g, game.Player1, &game.CardDef{Name: "First Graveyard Card"})
				addCardToGraveyard(g, game.Player1, &game.CardDef{Name: "Second Graveyard Card"})
				return g, delveSpell(mana.Cost{mana.GenericMana(2)}), 0, 0
			},
			want: []string{
				"option=0:Normal cost",
				"manaTaps=[]",
				"convoke=[]",
				"delve=[Second Graveyard Card, First Graveyard Card]",
				"additional=[]",
				"sacrifices=[]",
				"life=0",
			},
		},
		{
			name: "X payment uses colored source before generic source",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, "Forest")
				addBasicLandPermanent(g, game.Player1, "Island")
				return g, xSpell(), 0, 1
			},
			want: []string{
				"option=0:Normal cost",
				"manaTaps=[Forest:G:1, Island:U:1]",
				"convoke=[]",
				"delve=[]",
				"additional=[]",
				"sacrifices=[]",
				"life=0",
			},
		},
		{
			name: "cost modifier increases generic payment",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.CostModifiers = append(g.CostModifiers, game.CostModifier{
					Kind:            game.CostModifierSpell,
					GenericIncrease: 1,
				})
				addBasicLandPermanent(g, game.Player1, "Forest")
				addBasicLandPermanent(g, game.Player1, "Island")
				return g, genericCostSpell(1), 0, 0
			},
			want: []string{
				"option=0:Normal cost",
				"manaTaps=[Island:U:1, Forest:G:1]",
				"convoke=[]",
				"delve=[]",
				"additional=[]",
				"sacrifices=[]",
				"life=0",
			},
		},
		{
			name: "sacrifice additional cost excludes sacrificed permanent from mana payment",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player1, namedCreature("Offering Creature"))
				addBasicLandPermanent(g, game.Player1, "Forest")
				return g, sacrificeCostSpell(), 0, 0
			},
			want: []string{
				"option=0:Normal cost",
				"manaTaps=[Forest:G:1]",
				"convoke=[]",
				"delve=[]",
				"additional=[Sacrifice a creature]",
				"sacrifices=[Offering Creature]",
				"life=0",
			},
		},
		{
			name: "kicker paid combines base and kicker mana in plan",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, "Forest")
				addBasicLandPermanent(g, game.Player1, "Forest")
				return g, kickerSpell(), 0, 0
			},
			kickerPaid: true,
			want: []string{
				"option=0:Normal cost",
				"manaTaps=[Forest:G:1, Forest:G:1]",
				"convoke=[]",
				"delve=[]",
				"additional=[]",
				"sacrifices=[]",
				"life=0",
			},
		},
		{
			name: "flashback alternative cost replaces base cost when cast from graveyard",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, "Forest")
				return g, flashbackSpell(), 0, 0
			},
			sourceZone: game.ZoneGraveyard,
			want: []string{
				"option=1:Flashback",
				"manaTaps=[Forest:G:1]",
				"convoke=[]",
				"delve=[]",
				"additional=[]",
				"sacrifices=[]",
				"life=0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, card, cardID, xValue := tt.setup()

			zone := tt.sourceZone
			if zone == game.ZoneNone {
				zone = game.ZoneHand
			}
			plan, ok := paymentOrch.buildSpellCostPlan(g, spellPaymentRequest{playerID: game.Player1, cardID: cardID, sourceZone: zone, card: card, xValue: xValue, kickerPaid: tt.kickerPaid})
			if !ok {
				t.Fatal("buildSpellCostPlan() = false, want true")
			}
			got := summarizeSpellCostPlan(g, plan)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("payment plan:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func summarizeSpellCostPlan(g *game.Game, plan spellCostPlan) []string {
	return []string{
		fmt.Sprintf("option=%d:%s", plan.option.index, plan.option.label),
		"manaTaps=" + manaTapList(g, plan.mana.manaTaps),
		"convoke=" + permanentList(g, plan.mana.convokeTaps),
		"delve=" + cardList(g, plan.mana.delveExiles),
		"additional=" + stringList(plan.additional.paid),
		"sacrifices=" + permanentList(g, plan.additional.sacrifices),
		fmt.Sprintf("life=%d", plan.additional.lifePaid+plan.mana.lifePayment),
	}
}

func manaTapList(g *game.Game, taps []manaTap) string {
	if len(taps) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(taps))
	for _, tap := range taps {
		parts = append(parts, fmt.Sprintf("%s:%s:%d", sourceName(g, tap.permanent.ObjectID), tap.color, tap.amount))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func permanentList(g *game.Game, permanents []*game.Permanent) string {
	if len(permanents) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(permanents))
	for _, permanent := range permanents {
		parts = append(parts, cardName(g, permanent.CardInstanceID))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func cardList(g *game.Game, cardIDs []id.ID) string {
	if len(cardIDs) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		parts = append(parts, cardName(g, cardID))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func stringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	return "[" + strings.Join(values, ", ") + "]"
}

func namedCreature(name string, colors ...mana.Color) *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{
		Name:      name,
		Types:     []game.CardType{game.TypeCreature},
		Colors:    colors,
		Power:     optPT(pt),
		Toughness: optPT(pt),
	}
}

func genericCostSpell(generic int) *game.CardDef {
	return &game.CardDef{
		Name:      "Generic Cost Spell",
		ManaCost:  optCost(mana.Cost{mana.GenericMana(generic)}),
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{Kind: game.SpellAbility}},
	}
}

func sacrificeCostSpell() *game.CardDef {
	return &game.CardDef{
		Name:     "Sacrifice Cost Spell",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			AdditionalCosts: []game.AdditionalCost{{
				Kind:               game.AdditionalCostSacrifice,
				Text:               "Sacrifice a creature",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      game.TypeCreature,
			}},
		}},
	}
}

// kickerSpell returns a sorcery with a {G} base cost and a {G} kicker cost.
// When kicked (kickerPaid=true) the combined cost is {G}{G}.
func kickerSpell() *game.CardDef {
	return &game.CardDef{
		Name:     "Kicker Spell",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{
			Kind:       game.SpellAbility,
			KickerCost: greenCost(),
		}},
	}
}
