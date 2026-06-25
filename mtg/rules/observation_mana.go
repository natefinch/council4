package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/eval"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
)

// cardFaceManaColors returns the colors of mana the face's mana abilities can
// add, deduplicated and in WUBRG order. Colorless-only production yields no
// colors. It lets an observation report a land's color fixing without exposing
// ability internals to agents. A "{T}: Add one mana of the chosen color" source
// yields no colors here because the chosen color is unknowable until the
// permanent enters; abilitiesManaProduction resolves it for permanents.
// commanderColors resolves a commander-identity source (Command Tower) to the
// controller's commander color identity; pass nil when it is unknown.
func cardFaceManaColors(face *game.CardFace, commanderColors []color.Color) []color.Color {
	var found colorSet
	for i := range face.ManaAbilities {
		manaAbilityColors(&face.ManaAbilities[i], nil, commanderColors, &found)
	}
	return found.ordered()
}

// abilitiesManaProduction reports whether any of the effective abilities is a
// mana ability and the colors those mana abilities can add. Granted mana
// abilities are included because the input is the permanent's effective ability
// set. entryChoices resolves entry-time color choices (CR 614.12) so a "{T}: Add
// one mana of the chosen color" source reports its chosen color. commanderColors
// resolves a commander-identity source (Command Tower) to the controller's
// commander color identity; pass nil when it is unknown.
func abilitiesManaProduction(abilities []game.Ability, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult, commanderColors []color.Color) (bool, []color.Color) {
	var found colorSet
	producesMana := false
	for _, ability := range abilities {
		body, ok := ability.(*game.ManaAbility)
		if !ok {
			continue
		}
		producesMana = true
		manaAbilityColors(body, entryChoices, commanderColors, &found)
	}
	return producesMana, found.ordered()
}

// manaAbilityColors adds to found every color the mana ability can add, reading
// fixed AddMana colors, entry-time chosen colors, and the color options of any
// mana-color choice it makes.
func manaAbilityColors(body *game.ManaAbility, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult, commanderColors []color.Color, found *colorSet) {
	if len(body.Content.Modes) == 0 {
		return
	}
	for m := range body.Content.Modes {
		sequence := body.Content.Modes[m].Sequence
		for i := range sequence {
			addInstructionManaColors(sequence[i].Primitive, entryChoices, commanderColors, found)
		}
	}
}

func addInstructionManaColors(primitive game.Primitive, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult, commanderColors []color.Color, found *colorSet) {
	if primitive == nil {
		return
	}
	switch primitive.Kind() {
	case game.PrimitiveAddMana:
		add, ok := primitive.(game.AddMana)
		if !ok {
			return
		}
		if c, ok := manaColor(add.ManaColor); ok {
			found.add(c)
		}
		if add.EntryChoiceFrom != "" {
			if result, ok := entryChoices[add.EntryChoiceFrom]; ok {
				if c, ok := manaColor(result.Color); ok {
					found.add(c)
				}
			}
		}
	case game.PrimitiveChoose:
		choose, ok := primitive.(game.Choose)
		if !ok || choose.Choice.Kind != game.ResolutionChoiceMana {
			return
		}
		if choose.Choice.ColorSource == game.ResolutionChoiceColorSourceLandsProduce {
			// The colors come from other lands' production. Such a derived
			// ability contributes no colors of its own to this report, matching
			// CR's loop-avoidance ruling and keeping landsProduceMana's scan from
			// recursing into another "lands could produce" source.
			return
		}
		if choose.Choice.ColorSource == game.ResolutionChoiceColorSourceLinkedExileColors {
			// The colors come from a card imprinted on the source permanent,
			// unknown without a specific permanent and its recorded link. At the
			// face level no imprint exists, so this ability contributes no colors
			// to the static report (CR 202.2); the actual colors are computed at
			// activation/resolution from the linked exiled card.
			return
		}
		if choose.Choice.ColorSource == game.ResolutionChoiceColorSourceTriggerLandProduced {
			// The colors come from the type the triggering land just produced,
			// known only at resolution of the firing tapped-for-mana trigger. This
			// derived ability contributes no colors of its own to the static
			// report (CR 605.1a).
			return
		}
		if choose.Choice.ColorSource == game.ResolutionChoiceColorSourceCommanderIdentity {
			// A commander-identity source (Command Tower, Path of Ancestry) adds
			// one mana of any color in the controller's commander color identity
			// (CR 903.4), not any color. Resolve it to that identity; an empty
			// identity (no modeled or colorless commander) contributes no colors.
			for _, c := range commanderColors {
				found.add(c)
			}
			return
		}
		if choose.Choice.ColorSource != game.ResolutionChoiceColorSourceStatic {
			// "any color" and other unresolved dynamic choices can yield any color.
			for _, c := range color.AllColors() {
				found.add(c)
			}
			return
		}
		for _, mc := range choose.Choice.Colors {
			if c, ok := manaColor(mc); ok {
				found.add(c)
			}
		}
	default:
	}
}

// abilitiesProduceColorless reports whether any of the given mana abilities
// could add colorless ({C}) mana. It reads fixed AddMana colorless output,
// entry-chosen colorless, and static mana-color choices whose options include
// colorless. Dynamic color sources contribute no colorless: "any color" and
// commander-identity choices yield only colored mana (CR 106.5), and a
// lands-produce source contributes nothing here to avoid recursion. It supports
// the "any type" lands-produce wording (Reflecting Pool), which offers colorless
// when a scanned land could produce it.
func abilitiesProduceColorless(abilities []game.Ability, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	for _, ability := range abilities {
		body, ok := ability.(*game.ManaAbility)
		if !ok {
			continue
		}
		if manaAbilityProducesColorless(body, entryChoices) {
			return true
		}
	}
	return false
}

// manaAbilityProducesColorless reports whether the mana ability could add
// colorless mana, inspecting each AddMana and static mana-color Choose in its
// instruction sequence.
func manaAbilityProducesColorless(body *game.ManaAbility, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	if len(body.Content.Modes) == 0 {
		return false
	}
	for m := range body.Content.Modes {
		sequence := body.Content.Modes[m].Sequence
		for i := range sequence {
			if instructionProducesColorless(sequence[i].Primitive, entryChoices) {
				return true
			}
		}
	}
	return false
}

func instructionProducesColorless(primitive game.Primitive, entryChoices map[game.ChoiceKey]game.ResolutionChoiceResult) bool {
	if primitive == nil {
		return false
	}
	switch primitive.Kind() {
	case game.PrimitiveAddMana:
		add, ok := primitive.(game.AddMana)
		if !ok {
			return false
		}
		if add.ManaColor == mana.C {
			return true
		}
		if add.EntryChoiceFrom != "" {
			if result, ok := entryChoices[add.EntryChoiceFrom]; ok && result.Color == mana.C {
				return true
			}
		}
		return false
	case game.PrimitiveChoose:
		choose, ok := primitive.(game.Choose)
		if !ok || choose.Choice.Kind != game.ResolutionChoiceMana ||
			choose.Choice.ColorSource != game.ResolutionChoiceColorSourceStatic {
			return false
		}
		return slices.Contains(choose.Choice.Colors, mana.C)
	default:
		return false
	}
}

// manaColor maps a mana color to its card color, reporting false for colorless.
func manaColor(mc mana.Color) (color.Color, bool) {
	switch mc {
	case mana.W:
		return color.White, true
	case mana.U:
		return color.Blue, true
	case mana.B:
		return color.Black, true
	case mana.R:
		return color.Red, true
	case mana.G:
		return color.Green, true
	default:
		return "", false
	}
}

// commanderIdentityColors returns the colors in the player's commander color
// identity, used to resolve a commander-identity mana source (Command Tower,
// Path of Ancestry) to concrete colors. It is empty when the player has no
// modeled commander or a colorless one.
func commanderIdentityColors(g *game.Game, playerID game.PlayerID) []color.Color {
	player, ok := playerByID(g, playerID)
	if !ok || player.CommanderInstanceID == 0 {
		return nil
	}
	card, ok := g.GetCardInstance(player.CommanderInstanceID)
	if !ok || card.Def == nil {
		return nil
	}
	return card.Def.ColorIdentity.Colors()
}

// cardFaceRampsLand reports whether casting or resolving the face puts a land
// onto the battlefield (land ramp), reading the value-oriented effect IR of the
// face's spell ability and any enters/triggered abilities so a land-fetching
// sorcery (Rampant Growth) and a land-fetching enters trigger (Farhaven Elf)
// both count. It looks for an EffectLandRamp atom rather than inspecting search
// internals, keeping ramp detection out of strategy code.
func cardFaceRampsLand(face *game.CardFace) bool {
	if face.SpellAbility.Exists && abilityRampsLand(&face.SpellAbility.Val) {
		return true
	}
	for i := range face.TriggeredAbilities {
		if abilityRampsLand(&face.TriggeredAbilities[i]) {
			return true
		}
	}
	for i := range face.ActivatedAbilities {
		if abilityRampsLand(&face.ActivatedAbilities[i]) {
			return true
		}
	}
	return false
}

func abilityRampsLand(body game.Ability) bool {
	for _, atom := range eval.ScorableEffect(game.BodyContent(body)) {
		if atom.Kind == eval.EffectLandRamp {
			return true
		}
	}
	return false
}

// cardFaceEntersTapped reports whether the face always enters the battlefield
// tapped: it carries a self enters-tapped replacement with no condition and no
// pay-to-untap option. A conditional ("enters tapped unless you control two or
// more other lands") or pay-to-untap land is not flagged, since it can enter
// untapped, so an agent only treats guaranteed taplands as flexible-timing land
// drops.
func cardFaceEntersTapped(face *game.CardFace) bool {
	for i := range face.ReplacementAbilities {
		ability := face.ReplacementAbilities[i]
		if ability.Replacement.EntersTapped &&
			!ability.Replacement.EntersTappedOthers &&
			!ability.Replacement.Condition.Exists &&
			!ability.UnlessPaid.Exists {
			return true
		}
	}
	return false
}

// colorSet collects colors while preserving a stable WUBRG output order.
type colorSet struct {
	present map[color.Color]bool
}

func (s *colorSet) add(c color.Color) {
	if s.present == nil {
		s.present = make(map[color.Color]bool, len(color.AllColors()))
	}
	s.present[c] = true
}

func (s *colorSet) ordered() []color.Color {
	if len(s.present) == 0 {
		return nil
	}
	ordered := make([]color.Color, 0, len(s.present))
	for _, c := range color.AllColors() {
		if s.present[c] {
			ordered = append(ordered, c)
		}
	}
	return ordered
}
