package rules

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestLegalActionEnumerationCharacterization(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (*game.Game, game.PlayerID)
		want  []string
	}{
		{
			name: "spell costs modes targets and X values",
			setup: func() (*game.Game, game.PlayerID) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player2, &game.CardDef{
					Name:      "Silvercoat Lion",
					Types:     []game.CardType{game.TypeCreature},
					Power:     optPT(game.PT{Value: 2}),
					Toughness: optPT(game.PT{Value: 2}),
				})
				addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)
				addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)
				addCardToHand(g, game.Player1, artifactTargetSpell())
				addCardToHand(g, game.Player1, xSpell())
				addCardToHand(g, game.Player1, characterizationKickerSpell())
				addCardToHand(g, game.Player1, modalCharm())
				addCardToHand(g, game.Player1, zeroCostSpell())
				addCardToHand(g, game.Player1, noCostSpell())
				addCardToHand(g, game.Player1, &game.CardDef{Name: "Forest", Types: []game.CardType{game.TypeLand}})
				setMainPhasePriority(g, game.Player1)
				return g, game.Player1
			},
			want: []string{
				"play-land Forest face=0",
				"cast No Cost Spell from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"cast Zero Cost Spell from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"cast Test Charm from Hand face=0 x=0 modes=[0] kicked=false targets=[]",
				"cast Test Charm from Hand face=0 x=0 modes=[1] kicked=false targets=[permanent Silvercoat Lion]",
				"cast Characterization Kicker from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"cast Characterization Kicker from Hand face=0 x=0 modes=[] kicked=true targets=[]",
				"cast Characterization X Spell from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"cast Characterization X Spell from Hand face=0 x=1 modes=[] kicked=false targets=[]",
				"pass",
			},
		},
		{
			name: "commander and graveyard casting",
			setup: func() (*game.Game, game.PlayerID) {
				g := newCommanderCastGame(greenCommanderWithCost())
				addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)
				cardID := addCardToHand(g, game.Player1, flashbackSpell())
				g.Players[game.Player1].Hand.Remove(cardID)
				g.Players[game.Player1].Graveyard.Add(cardID)
				setMainPhasePriority(g, game.Player1)
				return g, game.Player1
			},
			want: []string{
				"cast Characterization Flashback from Graveyard face=0 x=0 modes=[] kicked=false targets=[]",
				"cast Castable Commander from Command face=0 x=0 modes=[] kicked=false targets=[]",
				"pass",
			},
		},
		{
			name: "activated ability target enumeration",
			setup: func() (*game.Game, game.PlayerID) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player1, &game.CardDef{
					Name:  "Targeting Rod",
					Types: []game.CardType{game.TypeArtifact},
					Abilities: []game.AbilityDef{{
						Kind:    game.ActivatedAbility,
						Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"}},
						Effects: []game.Effect{{Type: game.EffectDamage, TargetIndex: 0, Amount: 1}},
					}},
				})
				setMainPhasePriority(g, game.Player1)
				return g, game.Player1
			},
			want: []string{
				"activate Targeting Rod ability=0 x=0 targets=[player 2]",
				"activate Targeting Rod ability=0 x=0 targets=[player 3]",
				"activate Targeting Rod ability=0 x=0 targets=[player 4]",
				"pass",
			},
		},
		{
			name: "insufficient mana clamps X choices",
			setup: func() (*game.Game, game.PlayerID) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)
				addCardToHand(g, game.Player1, xSpell())
				setMainPhasePriority(g, game.Player1)
				return g, game.Player1
			},
			want: []string{
				"cast Characterization X Spell from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"pass",
			},
		},
		{
			name: "convoke and delve make otherwise unpayable spells legal",
			setup: func() (*game.Game, game.PlayerID) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatCreaturePermanent(g, game.Player1)
				addCardToGraveyard(g, game.Player1, &game.CardDef{Name: "First Graveyard Card"})
				addCardToGraveyard(g, game.Player1, &game.CardDef{Name: "Second Graveyard Card"})
				addCardToHand(g, game.Player1, delveSpell(mana.Cost{mana.GenericMana(2)}))
				addCardToHand(g, game.Player1, convokeSpell(mana.Cost{mana.GenericMana(1)}))
				setMainPhasePriority(g, game.Player1)
				return g, game.Player1
			},
			want: []string{
				"cast Convoke Spell from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"cast Delve Spell from Hand face=0 x=0 modes=[] kicked=false targets=[]",
				"pass",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, playerID := tt.setup()
			engine := NewEngine(nil)

			actions := engine.legalActions(g, playerID)
			assertActionsValidate(t, actions)
			got := summarizeLegalActions(g, actions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("legal actions:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}

			for i := 0; i < 5; i++ {
				actions = engine.legalActions(g, playerID)
				assertActionsValidate(t, actions)
				again := summarizeLegalActions(g, actions)
				if !reflect.DeepEqual(again, got) {
					t.Fatalf("legal actions changed on repeat %d:\n%s\nwant:\n%s", i, strings.Join(again, "\n"), strings.Join(got, "\n"))
				}
			}
		})
	}
}

func assertActionsValidate(t *testing.T, actions []action.Action) {
	t.Helper()
	for i, act := range actions {
		if err := act.Validate(); err != nil {
			t.Fatalf("legal action %d failed validation: %v (%+v)", i, err, act)
		}
	}
}

func summarizeLegalActions(g *game.Game, actions []action.Action) []string {
	summaries := make([]string, 0, len(actions))
	for _, act := range actions {
		summaries = append(summaries, summarizeLegalAction(g, act))
	}
	return summaries
}

func summarizeLegalAction(g *game.Game, act action.Action) string {
	switch act.Kind {
	case action.ActionPass:
		return "pass"
	case action.ActionPlayLand:
		playLand, ok := act.PlayLandPayload()
		if !ok {
			return "invalid play-land"
		}
		return fmt.Sprintf("play-land %s face=%d", cardName(g, playLand.CardID), playLand.Face)
	case action.ActionCastSpell:
		cast, ok := act.CastSpellPayload()
		if !ok {
			return "invalid cast"
		}
		return fmt.Sprintf(
			"cast %s from %s face=%d x=%d modes=%s kicked=%t targets=%s",
			cardName(g, cast.CardID),
			normalizedCastSourceZone(cast),
			cast.Face,
			cast.XValue,
			intList(cast.ChosenModes),
			cast.KickerPaid,
			targetList(g, cast.Targets),
		)
	case action.ActionActivateAbility:
		activate, ok := act.ActivateAbilityPayload()
		if !ok {
			return "invalid activate"
		}
		return fmt.Sprintf(
			"activate %s ability=%d x=%d targets=%s",
			sourceName(g, activate.SourceID),
			activate.AbilityIndex,
			activate.XValue,
			targetList(g, activate.Targets),
		)
	case action.ActionCastFaceDown:
		cast, ok := act.CastFaceDownPayload()
		if !ok {
			return "invalid cast-face-down"
		}
		return fmt.Sprintf("cast-face-down %s face=%d kind=%d", cardName(g, cast.CardID), cast.Face, cast.FaceDownKind)
	case action.ActionTurnFaceUp:
		turn, ok := act.TurnFaceUpPayload()
		if !ok {
			return "invalid turn-face-up"
		}
		return fmt.Sprintf("turn-face-up %s", sourceName(g, turn.PermanentID))
	default:
		return fmt.Sprintf("action-kind %d", act.Kind)
	}
}

func cardName(g *game.Game, cardID id.ID) string {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return fmt.Sprintf("missing-card-%d", cardID)
	}
	return card.Def.Name
}

func sourceName(g *game.Game, sourceID id.ID) string {
	if permanent, ok := permanentByObjectID(g, sourceID); ok {
		return cardName(g, permanent.CardInstanceID)
	}
	return cardName(g, sourceID)
}

func targetList(g *game.Game, targets []game.Target) string {
	if len(targets) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(targets))
	for _, target := range targets {
		parts = append(parts, targetSummary(g, target))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func targetSummary(g *game.Game, target game.Target) string {
	switch target.Kind {
	case game.TargetPermanent:
		if permanent, ok := permanentByObjectID(g, target.PermanentID); ok {
			return "permanent " + cardName(g, permanent.CardInstanceID)
		}
		return fmt.Sprintf("permanent missing-%d", target.PermanentID)
	case game.TargetPlayer:
		return fmt.Sprintf("player %d", int(target.PlayerID)+1)
	case game.TargetStackObject:
		return fmt.Sprintf("stack-object %d", target.StackObjectID)
	default:
		return fmt.Sprintf("target-kind %d", target.Kind)
	}
}

func intList(values []int) string {
	if len(values) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func noCostSpell() *game.CardDef {
	return &game.CardDef{
		Name:      "No Cost Spell",
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{Kind: game.SpellAbility}},
	}
}

func zeroCostSpell() *game.CardDef {
	return &game.CardDef{
		Name:      "Zero Cost Spell",
		ManaCost:  optCost(mana.Cost{mana.GenericMana(0)}),
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{Kind: game.SpellAbility}},
	}
}

func xSpell() *game.CardDef {
	return &game.CardDef{
		Name:      "Characterization X Spell",
		ManaCost:  optCost(mana.Cost{mana.VariableMana(), mana.ColoredMana(mana.Green)}),
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{Kind: game.SpellAbility}},
	}
}

func characterizationKickerSpell() *game.CardDef {
	return &game.CardDef{
		Name:      "Characterization Kicker",
		Types:     []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{Kind: game.SpellAbility, KickerCost: greenCost()}},
	}
}

func artifactTargetSpell() *game.CardDef {
	return &game.CardDef{
		Name:  "No Legal Artifact Target",
		Types: []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "artifact"}},
		}},
	}
}

func flashbackSpell() *game.CardDef {
	return &game.CardDef{
		Name:     "Characterization Flashback",
		Types:    []game.CardType{game.TypeSorcery},
		ManaCost: optCost(mana.Cost{mana.GenericMana(5)}),
		Abilities: []game.AbilityDef{
			{Kind: game.StaticAbility, Keywords: []game.Keyword{game.Flashback}},
			{
				Kind: game.SpellAbility,
				AlternativeCosts: []game.AlternativeCost{{
					Label:    flashbackAlternativeLabel,
					ManaCost: greenCost(),
				}},
			},
		},
	}
}
