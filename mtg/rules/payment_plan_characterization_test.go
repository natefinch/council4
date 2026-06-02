package rules

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func TestSpellPaymentPlanCharacterization(t *testing.T) {
	// This pins payment planner behavior across structural refactors. Keep the
	// coverage when entry points move; update only the call seam, not the
	// behavior being summarized.
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
				"tapped=[Green Convoke Creature, Plain Convoke Creature]",
				"exile=[]",
				"graveyard=[]",
				"additional=[]",
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
				"tapped=[]",
				"exile=[First Graveyard Card, Second Graveyard Card]",
				"graveyard=[]",
				"additional=[]",
				"life=0",
			},
		},
		{
			name: "X payment uses colored source before generic source",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, types.Forest)
				addBasicLandPermanent(g, game.Player1, types.Island)
				return g, xSpell(), 0, 1
			},
			want: []string{
				"option=0:Normal cost",
				"tapped=[Forest, Island]",
				"exile=[]",
				"graveyard=[]",
				"additional=[]",
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
				addBasicLandPermanent(g, game.Player1, types.Forest)
				addBasicLandPermanent(g, game.Player1, types.Island)
				return g, genericCostSpell(1), 0, 0
			},
			want: []string{
				"option=0:Normal cost",
				"tapped=[Forest, Island]",
				"exile=[]",
				"graveyard=[]",
				"additional=[]",
				"life=0",
			},
		},
		{
			name: "sacrifice additional cost excludes sacrificed permanent from mana payment",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player1, namedCreature("Offering Creature"))
				addBasicLandPermanent(g, game.Player1, types.Forest)
				return g, sacrificeCostSpell(), 0, 0
			},
			want: []string{
				"option=0:Normal cost",
				"tapped=[Forest]",
				"exile=[]",
				"graveyard=[Offering Creature]",
				"additional=[Sacrifice a creature]",
				"life=0",
			},
		},
		{
			name: "kicker paid combines base and kicker mana in plan",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, types.Forest)
				addBasicLandPermanent(g, game.Player1, types.Forest)
				return g, kickerSpell(), 0, 0
			},
			kickerPaid: true,
			want: []string{
				"option=0:Normal cost",
				"tapped=[Forest, Forest]",
				"exile=[]",
				"graveyard=[]",
				"additional=[]",
				"life=0",
			},
		},
		{
			name: "flashback alternative cost replaces base cost when cast from graveyard",
			setup: func() (*game.Game, *game.CardDef, id.ID, int) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, types.Forest)
				return g, flashbackSpell(), 0, 0
			},
			sourceZone: game.ZoneGraveyard,
			want: []string{
				"option=1:Flashback",
				"tapped=[Forest]",
				"exile=[]",
				"graveyard=[]",
				"additional=[]",
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
			req := payment.SpellRequest{PlayerID: game.Player1, CardID: cardID, SourceZone: zone, Card: card, XValue: xValue, KickerPaid: tt.kickerPaid}
			options := paymentOrch.planner(g).PayableSpellOptions(req)
			if len(options) == 0 {
				t.Fatal("PayableSpellOptions() empty, want payable option")
			}
			lifeBefore := g.Players[game.Player1].Life
			additionalPaid, ok := paymentOrch.paySpellCosts(g, req)
			if !ok {
				t.Fatal("PaySpellCosts() = false, want true")
			}
			got := summarizeSpellPayment(g, options[0], additionalPaid, lifeBefore)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("payment summary:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func summarizeSpellPayment(g *game.Game, option payment.SpellOptionSummary, additionalPaid []string, lifeBefore int) []string {
	player := g.Players[game.Player1]
	return []string{
		fmt.Sprintf("option=%d:%s", option.Index, option.Label),
		"tapped=" + tappedPermanentList(g),
		"exile=" + cardList(g, player.Exile.All()),
		"graveyard=" + cardList(g, player.Graveyard.All()),
		"additional=" + stringList(additionalPaid),
		fmt.Sprintf("life=%d", lifeBefore-player.Life),
	}
}

func tappedPermanentList(g *game.Game) string {
	var permanents []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Tapped {
			permanents = append(permanents, permanent)
		}
	}
	return permanentList(g, permanents)
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
		Types:     []types.Card{types.Creature},
		Colors:    colors,
		Power:     optPT(pt),
		Toughness: optPT(pt),
	}
}

func genericCostSpell(generic int) *game.CardDef {
	return &game.CardDef{
		Name:      "Generic Cost Spell",
		ManaCost:  optCost(mana.Cost{mana.GenericMana(generic)}),
		Types:     []types.Card{types.Sorcery},
		Abilities: []game.AbilityDef{{Kind: game.SpellAbility}},
	}
}

func sacrificeCostSpell() *game.CardDef {
	return &game.CardDef{
		Name:     "Sacrifice Cost Spell",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Sorcery},
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			AdditionalCosts: []game.AdditionalCost{{
				Kind:               game.AdditionalCostSacrifice,
				Text:               "Sacrifice a creature",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Creature,
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
		Types:    []types.Card{types.Sorcery},
		Abilities: []game.AbilityDef{{
			Kind:       game.SpellAbility,
			KickerCost: greenCost(),
		}},
	}
}
