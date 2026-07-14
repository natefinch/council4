package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// KeywordKind identifies a canonical Oracle keyword. The parser owns keyword
// spelling; downstream stages consume this typed identity.
type KeywordKind string

// Oracle keywords currently consumed by semantic compilation or card generation.
const (
	KeywordUnknown          KeywordKind = ""
	KeywordAffinity         KeywordKind = "KeywordAffinity"
	KeywordAnnihilator      KeywordKind = "KeywordAnnihilator"
	KeywordAscend           KeywordKind = "KeywordAscend"
	KeywordBanding          KeywordKind = "KeywordBanding"
	KeywordBargain          KeywordKind = "KeywordBargain"
	KeywordBloodthirst      KeywordKind = "KeywordBloodthirst"
	KeywordCascade          KeywordKind = "KeywordCascade"
	KeywordChangeling       KeywordKind = "KeywordChangeling"
	KeywordChampion         KeywordKind = "KeywordChampion"
	KeywordCompanion        KeywordKind = "KeywordCompanion"
	KeywordConvoke          KeywordKind = "KeywordConvoke"
	KeywordCumulativeUpkeep KeywordKind = "KeywordCumulativeUpkeep"
	KeywordCycling          KeywordKind = "KeywordCycling"
	KeywordDeathtouch       KeywordKind = "KeywordDeathtouch"
	KeywordDecayed          KeywordKind = "KeywordDecayed"
	KeywordDefender         KeywordKind = "KeywordDefender"
	KeywordDelve            KeywordKind = "KeywordDelve"
	KeywordDevoid           KeywordKind = "KeywordDevoid"
	KeywordDisguise         KeywordKind = "KeywordDisguise"
	KeywordDredge           KeywordKind = "KeywordDredge"
	KeywordDoubleStrike     KeywordKind = "KeywordDoubleStrike"
	// KeywordEcho is the Echo keyword (CR 702.29): "Echo <cost>" is a triggered
	// ability that, at the beginning of the controller's upkeep, sacrifices this
	// permanent unless the controller pays its echo cost, but only if it came
	// under that player's control since the beginning of their last upkeep. The
	// mana echo cost is carried by Parameter as a fixed mana cost, exactly like
	// Cumulative upkeep. The non-mana em-dash forms ("Echo—Discard a card.") are
	// not recognized and fail closed as unsupported.
	KeywordEcho           KeywordKind = "KeywordEcho"
	KeywordEmerge         KeywordKind = "KeywordEmerge"
	KeywordEnchant        KeywordKind = "KeywordEnchant"
	KeywordEquip          KeywordKind = "KeywordEquip"
	KeywordEscape         KeywordKind = "KeywordEscape"
	KeywordEternalize     KeywordKind = "KeywordEternalize"
	KeywordEmbalm         KeywordKind = "KeywordEmbalm"
	KeywordExalted        KeywordKind = "KeywordExalted"
	KeywordEvolve         KeywordKind = "KeywordEvolve"
	KeywordFabricate      KeywordKind = "KeywordFabricate"
	KeywordFear           KeywordKind = "KeywordFear"
	KeywordFirstStrike    KeywordKind = "KeywordFirstStrike"
	KeywordFlash          KeywordKind = "KeywordFlash"
	KeywordFlashback      KeywordKind = "KeywordFlashback"
	KeywordFlying         KeywordKind = "KeywordFlying"
	KeywordForetell       KeywordKind = "KeywordForetell"
	KeywordGift           KeywordKind = "KeywordGift"
	KeywordHaste          KeywordKind = "KeywordHaste"
	KeywordHexproof       KeywordKind = "KeywordHexproof"
	KeywordHorsemanship   KeywordKind = "KeywordHorsemanship"
	KeywordImprovise      KeywordKind = "KeywordImprovise"
	KeywordIndestructible KeywordKind = "KeywordIndestructible"
	KeywordInfect         KeywordKind = "KeywordInfect"
	KeywordIntimidate     KeywordKind = "KeywordIntimidate"
	KeywordJumpStart      KeywordKind = "KeywordJumpStart"
	KeywordKicker         KeywordKind = "KeywordKicker"
	KeywordLifelink       KeywordKind = "KeywordLifelink"
	KeywordLivingWeapon   KeywordKind = "KeywordLivingWeapon"
	KeywordLivingMetal    KeywordKind = "KeywordLivingMetal"
	KeywordMadness        KeywordKind = "KeywordMadness"
	KeywordMenace         KeywordKind = "KeywordMenace"
	KeywordMorph          KeywordKind = "KeywordMorph"
	KeywordMultikicker    KeywordKind = "KeywordMultikicker"
	KeywordMutate         KeywordKind = "KeywordMutate"
	KeywordNinjutsu       KeywordKind = "KeywordNinjutsu"
	KeywordOffspring      KeywordKind = "KeywordOffspring"
	KeywordOutlast        KeywordKind = "KeywordOutlast"
	KeywordPersist        KeywordKind = "KeywordPersist"
	KeywordPlot           KeywordKind = "KeywordPlot"
	KeywordProtection     KeywordKind = "KeywordProtection"
	KeywordProwess        KeywordKind = "KeywordProwess"
	KeywordReadAhead      KeywordKind = "KeywordReadAhead"
	KeywordReach          KeywordKind = "KeywordReach"
	// KeywordRetrace is the Retrace keyword (CR 702.81): "You may cast this card
	// from your graveyard by discarding a land card in addition to paying its
	// other costs." It is a non-parameterized graveyard alternative-casting
	// permission, also conferred on other graveyard cards by "<filter> cards in
	// your graveyard have retrace" statics (Six, Wrenn and Six Emblem).
	KeywordRetrace     KeywordKind = "KeywordRetrace"
	KeywordShadow      KeywordKind = "KeywordShadow"
	KeywordScavenge    KeywordKind = "KeywordScavenge"
	KeywordShroud      KeywordKind = "KeywordShroud"
	KeywordSkulk       KeywordKind = "KeywordSkulk"
	KeywordSplitSecond KeywordKind = "KeywordSplitSecond"
	KeywordStorm       KeywordKind = "KeywordStorm"
	KeywordSuspend     KeywordKind = "KeywordSuspend"
	// KeywordTransmute is the Transmute keyword (CR 702.49): an activated
	// ability that functions only while the card is in its owner's hand,
	// discarding the card and paying its mana cost to search the library, at
	// sorcery speed, for a card with the same mana value as this card.
	KeywordTransmute KeywordKind = "KeywordTransmute"
	// KeywordHideaway is the Hideaway N land keyword (CR 702.75). Its integer
	// parameter is the number of cards looked at from the top of the library
	// when the permanent enters, one of which is exiled face down to be played
	// later by the source's activated ability.
	KeywordHideaway  KeywordKind = "KeywordHideaway"
	KeywordToxic     KeywordKind = "KeywordToxic"
	KeywordTrample   KeywordKind = "KeywordTrample"
	KeywordUnearth   KeywordKind = "KeywordUnearth"
	KeywordUndying   KeywordKind = "KeywordUndying"
	KeywordUnleash   KeywordKind = "KeywordUnleash"
	KeywordVigilance KeywordKind = "KeywordVigilance"
	KeywordWard      KeywordKind = "KeywordWard"
	KeywordWither    KeywordKind = "KeywordWither"
	KeywordRiot      KeywordKind = "KeywordRiot"
	// KeywordLandcycling and the typed variants below are the landcycling
	// keyword family (CR 702.29). Each is a cycling ability whose
	// discard-from-hand activation searches the library for a land matching a
	// fixed land filter rather than drawing a card.
	KeywordLandcycling         KeywordKind = "KeywordLandcycling"
	KeywordBasicLandcycling    KeywordKind = "KeywordBasicLandcycling"
	KeywordArtifactLandcycling KeywordKind = "KeywordArtifactLandcycling"
	KeywordPlainscycling       KeywordKind = "KeywordPlainscycling"
	KeywordIslandcycling       KeywordKind = "KeywordIslandcycling"
	KeywordSwampcycling        KeywordKind = "KeywordSwampcycling"
	KeywordMountaincycling     KeywordKind = "KeywordMountaincycling"
	KeywordForestcycling       KeywordKind = "KeywordForestcycling"
	KeywordDethrone            KeywordKind = "KeywordDethrone"
	KeywordFlanking            KeywordKind = "KeywordFlanking"
	KeywordSoulshift           KeywordKind = "KeywordSoulshift"
	KeywordSplice              KeywordKind = "KeywordSplice"
	KeywordRampage             KeywordKind = "KeywordRampage"
	KeywordTraining            KeywordKind = "KeywordTraining"
	KeywordMyriad              KeywordKind = "KeywordMyriad"
	// KeywordMobilize is the Mobilize N keyword (CR 702.169). Its parameter is
	// the number of tapped-and-attacking 1/1 red Warrior tokens created whenever
	// the creature attacks: a fixed integer ("Mobilize 2") or the rules-derived
	// "Mobilize X, where X is the number of creature cards in your graveyard".
	KeywordMobilize KeywordKind = "KeywordMobilize"
	// KeywordSaddle is the Saddle N keyword (Mounts, CR 702.166). Its integer
	// parameter is the total power of other creatures that must be tapped to
	// make the Mount saddled until end of turn.
	KeywordSaddle KeywordKind = "KeywordSaddle"
	// KeywordCrew is the Crew N keyword (Vehicles, CR 702.122). Its integer
	// parameter is the total power of creatures that must be tapped to make the
	// Vehicle become an artifact creature until end of turn.
	KeywordCrew KeywordKind = "KeywordCrew"
	// KeywordLandwalk and the typed variants below are the landwalk evasion
	// keyword family (CR 702.14). Each typed variant keys off the defending
	// player controlling a land of its named subtype; plain Landwalk keys off
	// any land.
	KeywordLandwalk     KeywordKind = "KeywordLandwalk"
	KeywordPlainswalk   KeywordKind = "KeywordPlainswalk"
	KeywordIslandwalk   KeywordKind = "KeywordIslandwalk"
	KeywordSwampwalk    KeywordKind = "KeywordSwampwalk"
	KeywordMountainwalk KeywordKind = "KeywordMountainwalk"
	KeywordForestwalk   KeywordKind = "KeywordForestwalk"
	KeywordDesertwalk   KeywordKind = "KeywordDesertwalk"
	// KeywordNonbasicLandwalk is the "nonbasic landwalk" qualifier variant: the
	// creature can't be blocked as long as the defending player controls a
	// nonbasic land (a land without the Basic supertype).
	KeywordNonbasicLandwalk KeywordKind = "KeywordNonbasicLandwalk"
	// KeywordEvoke is the Evoke alternative-cost keyword (CR 702.74): "Evoke
	// <cost>" lets the spell be cast for its evoke cost, and the resulting
	// permanent is sacrificed when it enters the battlefield.
	KeywordEvoke KeywordKind = "KeywordEvoke"
	// KeywordRebound is the Rebound keyword (CR 702.88): if you cast this spell
	// from your hand, exile it as it resolves; at the beginning of your next
	// upkeep you may cast it from exile without paying its mana cost.
	KeywordRebound KeywordKind = "KeywordRebound"
	// KeywordSpectacle is the Spectacle alternative-cost keyword (CR 702.107):
	// "Spectacle <cost>" lets the spell be cast for its spectacle cost rather
	// than its mana cost if an opponent lost life this turn.
	KeywordSpectacle KeywordKind = "KeywordSpectacle"
	// KeywordStartEngines is the "Start your engines!" keyword (CR 702.179). It
	// is printed on a permanent with reminder text and seeds the controller's
	// speed to 1 if they have none. The recurring once-per-turn speed increase
	// on opponent life loss and the speed cap of 4 are built-in rules keyed off
	// the player's speed.
	KeywordStartEngines KeywordKind = "KeywordStartEngines"
	// KeywordFuse is the Fuse keyword (CR 702.102), printed on both halves of a
	// fuse split card: "You may cast one or both halves of this card from your
	// hand." It is a static ability of the card in any zone granting the
	// alternative permission to cast both halves as a single fused split spell.
	// It is appended at the end so the existing keyword block stays aligned.
	KeywordFuse KeywordKind = "KeywordFuse"
	// KeywordPartnerWith is the "Partner with <name>" keyword (CR 702.124e). It
	// names a specific partner card and grants an enters trigger that lets the
	// chosen player tutor the named partner into hand. Both the deck-construction
	// "partner commander" permission and the pair-fetch ETB are mechanics the
	// deterministic playtester does not simulate, so the keyword is recognized
	// and represented but not simulated. It is appended at the end so the
	// existing keyword block stays aligned.
	KeywordPartnerWith KeywordKind = "KeywordPartnerWith"
	// KeywordChooseABackground is the "Choose a Background" keyword (CR 702.124f):
	// "You can have a Background as a second commander." It is a deck-construction
	// permission the deterministic playtester does not simulate, so the keyword is
	// recognized and represented but not simulated. It is appended at the end so
	// the existing keyword block stays aligned.
	KeywordChooseABackground KeywordKind = "KeywordChooseABackground"
	// KeywordReconfigure is the Reconfigure keyword (CR 702.151), printed on
	// Equipment creatures: "Reconfigure <cost>" is a mana-cost activated ability
	// that attaches the source to target creature you control, or unattaches it,
	// only as a sorcery; while attached the source isn't a creature. The mana
	// cost is carried by Parameter exactly like Equip. It is appended at the end
	// so the existing keyword block stays aligned.
	KeywordReconfigure KeywordKind = "KeywordReconfigure"
	// KeywordMoreThanMeetsTheEye is the Transformers "More Than Meets the Eye"
	// alternative-cost keyword (CR 712): "More Than Meets the Eye <cost>" lets the
	// card be cast for that cost, and the resulting permanent enters the
	// battlefield converted, as its back face. The mana cost is carried by
	// Parameter exactly like Evoke. It is appended at the end so the existing
	// keyword block stays aligned.
	KeywordMoreThanMeetsTheEye KeywordKind = "KeywordMoreThanMeetsTheEye"
	// KeywordDash is the Dash alternative-cost keyword (CR 702.109): "Dash
	// <cost>" lets the creature spell be cast for its dash cost rather than its
	// mana cost; if it is, the creature gains haste and is returned to its
	// owner's hand at the beginning of the next end step. The mana cost is
	// carried by Parameter exactly like Evoke. It is appended at the end so the
	// existing keyword block stays aligned.
	KeywordDash KeywordKind = "KeywordDash"

	// KeywordBestow is the Bestow alternative-cost keyword (CR 702.103):
	// "Bestow <cost>" lets an enchantment creature card be cast as an Aura
	// spell that enchants a creature for its bestow cost rather than its mana
	// cost; while attached it is not a creature, and it becomes a creature
	// again if it stops being attached. The mana cost is carried by Parameter
	// exactly like Evoke and Dash. It is appended at the end so the existing
	// keyword block stays aligned.
	KeywordBestow KeywordKind = "KeywordBestow"
)

var keywordNames = map[KeywordKind]string{
	KeywordAffinity:            "Affinity",
	KeywordAnnihilator:         "Annihilator",
	KeywordAscend:              "Ascend",
	KeywordBanding:             "Banding",
	KeywordBargain:             "Bargain",
	KeywordBloodthirst:         "Bloodthirst",
	KeywordCascade:             "Cascade",
	KeywordChangeling:          "Changeling",
	KeywordChampion:            "Champion",
	KeywordCompanion:           "Companion",
	KeywordConvoke:             "Convoke",
	KeywordCumulativeUpkeep:    "Cumulative upkeep",
	KeywordCycling:             "Cycling",
	KeywordDeathtouch:          "Deathtouch",
	KeywordDecayed:             "Decayed",
	KeywordDefender:            "Defender",
	KeywordDelve:               "Delve",
	KeywordDevoid:              "Devoid",
	KeywordDisguise:            "Disguise",
	KeywordDredge:              "Dredge",
	KeywordDoubleStrike:        "Double strike",
	KeywordEcho:                "Echo",
	KeywordEmerge:              "Emerge",
	KeywordEnchant:             "Enchant",
	KeywordEquip:               "Equip",
	KeywordEscape:              "Escape",
	KeywordEternalize:          "Eternalize",
	KeywordEmbalm:              "Embalm",
	KeywordExalted:             "Exalted",
	KeywordEvolve:              "Evolve",
	KeywordEvoke:               "Evoke",
	KeywordFabricate:           "Fabricate",
	KeywordFear:                "Fear",
	KeywordFirstStrike:         "First strike",
	KeywordFlash:               "Flash",
	KeywordFlashback:           "Flashback",
	KeywordFlying:              "Flying",
	KeywordForetell:            "Foretell",
	KeywordGift:                "Gift",
	KeywordHaste:               "Haste",
	KeywordHexproof:            "Hexproof",
	KeywordHorsemanship:        "Horsemanship",
	KeywordImprovise:           "Improvise",
	KeywordIndestructible:      "Indestructible",
	KeywordInfect:              "Infect",
	KeywordIntimidate:          "Intimidate",
	KeywordJumpStart:           "Jump-start",
	KeywordKicker:              "Kicker",
	KeywordLifelink:            "Lifelink",
	KeywordLivingWeapon:        "Living weapon",
	KeywordLivingMetal:         "Living metal",
	KeywordMadness:             "Madness",
	KeywordMenace:              "Menace",
	KeywordMorph:               "Morph",
	KeywordMultikicker:         "Multikicker",
	KeywordMutate:              "Mutate",
	KeywordNinjutsu:            "Ninjutsu",
	KeywordOffspring:           "Offspring",
	KeywordOutlast:             "Outlast",
	KeywordPersist:             "Persist",
	KeywordPlot:                "Plot",
	KeywordProtection:          "Protection",
	KeywordProwess:             "Prowess",
	KeywordReadAhead:           "Read ahead",
	KeywordReach:               "Reach",
	KeywordReconfigure:         "Reconfigure",
	KeywordMoreThanMeetsTheEye: "More Than Meets the Eye",
	KeywordRetrace:             "Retrace",
	KeywordShadow:              "Shadow",
	KeywordScavenge:            "Scavenge",
	KeywordShroud:              "Shroud",
	KeywordSkulk:               "Skulk",
	KeywordSplitSecond:         "Split second",
	KeywordStorm:               "Storm",
	KeywordSuspend:             "Suspend",
	KeywordTransmute:           "Transmute",
	KeywordHideaway:            "Hideaway",
	KeywordToxic:               "Toxic",
	KeywordTrample:             "Trample",
	KeywordUnearth:             "Unearth",
	KeywordUndying:             "Undying",
	KeywordUnleash:             "Unleash",
	KeywordVigilance:           "Vigilance",
	KeywordWard:                "Ward",
	KeywordWither:              "Wither",
	KeywordRiot:                "Riot",
	KeywordLandcycling:         "Landcycling",
	KeywordBasicLandcycling:    "Basic landcycling",
	KeywordArtifactLandcycling: "Artifact landcycling",
	KeywordPlainscycling:       "Plainscycling",
	KeywordIslandcycling:       "Islandcycling",
	KeywordSwampcycling:        "Swampcycling",
	KeywordMountaincycling:     "Mountaincycling",
	KeywordForestcycling:       "Forestcycling",
	KeywordDethrone:            "Dethrone",
	KeywordFlanking:            "Flanking",
	KeywordSoulshift:           "Soulshift",
	KeywordSplice:              "Splice onto Arcane",
	KeywordRampage:             "Rampage",
	KeywordTraining:            "Training",
	KeywordMyriad:              "Myriad",
	KeywordMobilize:            "Mobilize",
	KeywordSaddle:              "Saddle",
	KeywordCrew:                "Crew",
	KeywordLandwalk:            "Landwalk",
	KeywordPlainswalk:          "Plainswalk",
	KeywordIslandwalk:          "Islandwalk",
	KeywordSwampwalk:           "Swampwalk",
	KeywordMountainwalk:        "Mountainwalk",
	KeywordForestwalk:          "Forestwalk",
	KeywordDesertwalk:          "Desertwalk",
	KeywordNonbasicLandwalk:    "Nonbasic landwalk",
	KeywordRebound:             "Rebound",
	KeywordSpectacle:           "Spectacle",
	KeywordStartEngines:        "Start your engines!",
	KeywordFuse:                "Fuse",
	KeywordPartnerWith:         "Partner with",
	KeywordChooseABackground:   "Choose a Background",
	KeywordDash:                "Dash",
	KeywordBestow:              "Bestow",
}

// String returns the parser-owned canonical keyword name.
func (k KeywordKind) String() string {
	if name, ok := keywordNames[k]; ok {
		return name
	}
	return "Unknown"
}

// OracleWord returns the lowercase Oracle word(s) for a keyword, the form used in
// "creature token with <keyword>" wording (e.g. KeywordFlying -> "flying",
// KeywordFirstStrike -> "first strike"). It fails closed for the unknown keyword.
func (k KeywordKind) OracleWord() (string, bool) {
	if k == KeywordUnknown {
		return "", false
	}
	name, ok := keywordNames[k]
	if !ok {
		return "", false
	}
	return strings.ToLower(name), true
}

type keywordNameGrammar struct {
	Kind  KeywordKind `json:",omitempty"`
	Words []string    `json:",omitempty"`
}

var keywordNameGrammars = []keywordNameGrammar{
	{Kind: KeywordDoubleStrike, Words: []string{"double", "strike"}},
	{Kind: KeywordFirstStrike, Words: []string{"first", "strike"}},
	{Kind: KeywordCumulativeUpkeep, Words: []string{"cumulative", "upkeep"}},
	{Kind: KeywordLivingWeapon, Words: []string{"living", "weapon"}},
	{Kind: KeywordLivingMetal, Words: []string{"living", "metal"}},
	{Kind: KeywordMoreThanMeetsTheEye, Words: []string{"more", "than", "meets", "the", "eye"}},
	{Kind: KeywordReadAhead, Words: []string{"read", "ahead"}},
	{Kind: KeywordSplitSecond, Words: []string{"split", "second"}},
	{Kind: KeywordPartnerWith, Words: []string{"partner", "with"}},
	{Kind: KeywordChooseABackground, Words: []string{"choose", "a", "background"}},
	{Kind: KeywordBasicLandcycling, Words: []string{"basic", "landcycling"}},
	{Kind: KeywordArtifactLandcycling, Words: []string{"artifact", "landcycling"}},
	{Kind: KeywordAffinity, Words: []string{"affinity"}},
	{Kind: KeywordAnnihilator, Words: []string{"annihilator"}},
	{Kind: KeywordAscend, Words: []string{"ascend"}},
	{Kind: KeywordBanding, Words: []string{"banding"}},
	{Kind: KeywordBargain, Words: []string{"bargain"}},
	{Kind: KeywordBloodthirst, Words: []string{"bloodthirst"}},
	{Kind: KeywordCascade, Words: []string{"cascade"}},
	{Kind: KeywordChangeling, Words: []string{"changeling"}},
	{Kind: KeywordChampion, Words: []string{"champion"}},
	{Kind: KeywordCompanion, Words: []string{"companion"}},
	{Kind: KeywordConvoke, Words: []string{"convoke"}},
	{Kind: KeywordCycling, Words: []string{"cycling"}},
	{Kind: KeywordDeathtouch, Words: []string{"deathtouch"}},
	{Kind: KeywordDecayed, Words: []string{"decayed"}},
	{Kind: KeywordDefender, Words: []string{"defender"}},
	{Kind: KeywordDelve, Words: []string{"delve"}},
	{Kind: KeywordDevoid, Words: []string{"devoid"}},
	{Kind: KeywordDisguise, Words: []string{"disguise"}},
	{Kind: KeywordDredge, Words: []string{"dredge"}},
	{Kind: KeywordEcho, Words: []string{"echo"}},
	{Kind: KeywordEmerge, Words: []string{"emerge"}},
	{Kind: KeywordEnchant, Words: []string{"enchant"}},
	{Kind: KeywordEquip, Words: []string{"equip"}},
	{Kind: KeywordEscape, Words: []string{"escape"}},
	{Kind: KeywordEternalize, Words: []string{"eternalize"}},
	{Kind: KeywordEmbalm, Words: []string{"embalm"}},
	{Kind: KeywordExalted, Words: []string{"exalted"}},
	{Kind: KeywordEvolve, Words: []string{"evolve"}},
	{Kind: KeywordEvoke, Words: []string{"evoke"}},
	{Kind: KeywordFabricate, Words: []string{"fabricate"}},
	{Kind: KeywordFear, Words: []string{"fear"}},
	{Kind: KeywordFlash, Words: []string{"flash"}},
	{Kind: KeywordFlashback, Words: []string{"flashback"}},
	{Kind: KeywordFlying, Words: []string{"flying"}},
	{Kind: KeywordForetell, Words: []string{"foretell"}},
	{Kind: KeywordGift, Words: []string{"gift"}},
	{Kind: KeywordHaste, Words: []string{"haste"}},
	{Kind: KeywordHexproof, Words: []string{"hexproof"}},
	{Kind: KeywordHorsemanship, Words: []string{"horsemanship"}},
	{Kind: KeywordImprovise, Words: []string{"improvise"}},
	{Kind: KeywordIndestructible, Words: []string{"indestructible"}},
	{Kind: KeywordInfect, Words: []string{"infect"}},
	{Kind: KeywordIntimidate, Words: []string{"intimidate"}},
	{Kind: KeywordJumpStart, Words: []string{"jump-start"}},
	{Kind: KeywordKicker, Words: []string{"kicker"}},
	{Kind: KeywordLifelink, Words: []string{"lifelink"}},
	{Kind: KeywordMadness, Words: []string{"madness"}},
	{Kind: KeywordMenace, Words: []string{"menace"}},
	{Kind: KeywordMorph, Words: []string{"morph"}},
	{Kind: KeywordMultikicker, Words: []string{"multikicker"}},
	{Kind: KeywordMutate, Words: []string{"mutate"}},
	{Kind: KeywordNinjutsu, Words: []string{"ninjutsu"}},
	{Kind: KeywordOffspring, Words: []string{"offspring"}},
	{Kind: KeywordOutlast, Words: []string{"outlast"}},
	{Kind: KeywordPersist, Words: []string{"persist"}},
	{Kind: KeywordPlot, Words: []string{"plot"}},
	{Kind: KeywordProtection, Words: []string{"protection"}},
	{Kind: KeywordProwess, Words: []string{"prowess"}},
	{Kind: KeywordReach, Words: []string{"reach"}},
	{Kind: KeywordReconfigure, Words: []string{"reconfigure"}},
	{Kind: KeywordRetrace, Words: []string{"retrace"}},
	{Kind: KeywordShadow, Words: []string{"shadow"}},
	{Kind: KeywordScavenge, Words: []string{"scavenge"}},
	{Kind: KeywordShroud, Words: []string{"shroud"}},
	{Kind: KeywordSkulk, Words: []string{"skulk"}},
	{Kind: KeywordStorm, Words: []string{"storm"}},
	{Kind: KeywordSuspend, Words: []string{"suspend"}},
	{Kind: KeywordTransmute, Words: []string{"transmute"}},
	{Kind: KeywordHideaway, Words: []string{"hideaway"}},
	{Kind: KeywordToxic, Words: []string{"toxic"}},
	{Kind: KeywordTrample, Words: []string{"trample"}},
	{Kind: KeywordUnearth, Words: []string{"unearth"}},
	{Kind: KeywordUndying, Words: []string{"undying"}},
	{Kind: KeywordUnleash, Words: []string{"unleash"}},
	{Kind: KeywordVigilance, Words: []string{"vigilance"}},
	{Kind: KeywordWard, Words: []string{"ward"}},
	{Kind: KeywordWither, Words: []string{"wither"}},
	{Kind: KeywordRiot, Words: []string{"riot"}},
	{Kind: KeywordLandcycling, Words: []string{"landcycling"}},
	{Kind: KeywordPlainscycling, Words: []string{"plainscycling"}},
	{Kind: KeywordIslandcycling, Words: []string{"islandcycling"}},
	{Kind: KeywordSwampcycling, Words: []string{"swampcycling"}},
	{Kind: KeywordMountaincycling, Words: []string{"mountaincycling"}},
	{Kind: KeywordForestcycling, Words: []string{"forestcycling"}},
	{Kind: KeywordDethrone, Words: []string{"dethrone"}},
	{Kind: KeywordFlanking, Words: []string{"flanking"}},
	{Kind: KeywordSoulshift, Words: []string{"soulshift"}},
	{Kind: KeywordSplice, Words: []string{"splice", "onto", "arcane"}},
	{Kind: KeywordRampage, Words: []string{"rampage"}},
	{Kind: KeywordTraining, Words: []string{"training"}},
	{Kind: KeywordMyriad, Words: []string{"myriad"}},
	{Kind: KeywordMobilize, Words: []string{"mobilize"}},
	{Kind: KeywordSaddle, Words: []string{"saddle"}},
	{Kind: KeywordCrew, Words: []string{"crew"}},
	{Kind: KeywordNonbasicLandwalk, Words: []string{"nonbasic", "landwalk"}},
	{Kind: KeywordRebound, Words: []string{"rebound"}},
	{Kind: KeywordLandwalk, Words: []string{"landwalk"}},
	{Kind: KeywordPlainswalk, Words: []string{"plainswalk"}},
	{Kind: KeywordIslandwalk, Words: []string{"islandwalk"}},
	{Kind: KeywordSwampwalk, Words: []string{"swampwalk"}},
	{Kind: KeywordMountainwalk, Words: []string{"mountainwalk"}},
	{Kind: KeywordForestwalk, Words: []string{"forestwalk"}},
	{Kind: KeywordDesertwalk, Words: []string{"desertwalk"}},
	{Kind: KeywordSpectacle, Words: []string{"spectacle"}},
	{Kind: KeywordDash, Words: []string{"dash"}},
	{Kind: KeywordBestow, Words: []string{"bestow"}},
	{Kind: KeywordStartEngines, Words: []string{"start", "your", "engines"}},
	{Kind: KeywordFuse, Words: []string{"fuse"}},
}

// KeywordParameterKind identifies the grammar used by a keyword parameter.
type KeywordParameterKind string

// Typed keyword parameter shapes.
const (
	KeywordParameterNone          KeywordParameterKind = ""
	KeywordParameterManaCost      KeywordParameterKind = "KeywordParameterManaCost"
	KeywordParameterInteger       KeywordParameterKind = "KeywordParameterInteger"
	KeywordParameterEnchantTarget KeywordParameterKind = "KeywordParameterEnchantTarget"
	KeywordParameterChampion      KeywordParameterKind = "KeywordParameterChampion"
	KeywordParameterProtection    KeywordParameterKind = "KeywordParameterProtection"
	KeywordParameterGift          KeywordParameterKind = "KeywordParameterGift"
	// KeywordParameterMobilizeDynamic is a rules-derived Mobilize count that is
	// not a printed integer ("Mobilize X, where X is ..."). The typed
	// MobilizeDynamicKind names which rules-derived count applies; the fixed
	// "Mobilize N" form uses KeywordParameterInteger instead.
	KeywordParameterMobilizeDynamic KeywordParameterKind = "KeywordParameterMobilizeDynamic"
)

// MobilizeDynamicKind identifies a rules-derived Mobilize count (CR 702.169).
// The parser owns the printed wording and maps it to one of these kinds; the
// compiler and lowering are text-blind and consume only the kind.
type MobilizeDynamicKind string

// Typed Mobilize dynamic count kinds.
const (
	MobilizeDynamicNone MobilizeDynamicKind = ""
	// MobilizeDynamicCreatureCardsInGraveyard is "Mobilize X, where X is the
	// number of creature cards in your graveyard" (Avenger of the Fallen).
	MobilizeDynamicCreatureCardsInGraveyard MobilizeDynamicKind = "MobilizeDynamicCreatureCardsInGraveyard"
)

// GiftKind identifies the typed gift promised by a Gift keyword action (CR
// 702.171). The parser maps the printed gift wording ("Gift a card", "Gift a
// Food", "Gift a Treasure", "Gift a tapped Fish") to one of these kinds; the
// compiler and lowering are text-blind and consume only the kind.
type GiftKind string

// Typed gift kinds.
const (
	GiftKindNone       GiftKind = ""
	GiftKindCard       GiftKind = "GiftKindCard"
	GiftKindFood       GiftKind = "GiftKindFood"
	GiftKindTreasure   GiftKind = "GiftKindTreasure"
	GiftKindTappedFish GiftKind = "GiftKindTappedFish"
)

// ProtectionParameter is the composable typed predicate following "Protection
// from". Exactly one predicate family is populated.
type ProtectionParameter struct {
	Everything   bool        `json:",omitempty"`
	EachColor    bool        `json:",omitempty"`
	Multicolored bool        `json:",omitempty"`
	Monocolored  bool        `json:",omitempty"`
	FromColors   []Color     `json:",omitempty"`
	FromTypes    []CardType  `json:",omitempty"`
	FromSubtypes []types.Sub `json:",omitempty"`
	ChosenColor  bool        `json:",omitempty"`
	// CommanderIdentityComplement marks "protection from each color that's not in
	// your commander's color identity" (Commander's Plate). The protected color
	// set is resolved dynamically by the rules from the granting ability
	// controller's commander color identity; the parser only records the family.
	CommanderIdentityComplement bool `json:",omitempty"`
}

// EnchantPredicate is the typed object restriction following an Enchant keyword.
// A permanent matches when it has any listed card type or any listed subtype
// (the union is disjunctive: "artifact or creature", "creature or Vehicle").
// Player and Opponent select a player object; Permanent selects any permanent.
// At most one of Player/Opponent/Permanent is set, and they are never combined
// with CardTypes or Subtypes. The zero value is the fail-closed unknown
// predicate.
type EnchantPredicate struct {
	Player    bool        `json:",omitempty"`
	Opponent  bool        `json:",omitempty"`
	Permanent bool        `json:",omitempty"`
	CardTypes []CardType  `json:",omitempty"`
	Subtypes  []types.Sub `json:",omitempty"`
	// YouControl restricts a permanent target to one the enchanting player
	// controls ("Enchant creature or planeswalker you control"). It applies only
	// to the card-type/subtype predicate; it is never set with Player or Opponent.
	YouControl bool `json:",omitempty"`
	// InGraveyard restricts the target to a matching card in a graveyard
	// ("Enchant creature card in a graveyard"). It marks the graveyard-card Aura
	// enchant restriction of the reanimation Aura family (Animate Dead, Dance of
	// the Dead) and is set only alongside a card-type predicate.
	InGraveyard bool `json:",omitempty"`
}

// Empty reports whether the predicate carries no recognized restriction. A bare
// "you control" controller restriction is not a recognized object class on its
// own, so it does not make a predicate non-empty.
func (p EnchantPredicate) Empty() bool {
	return !p.Player && !p.Opponent && !p.Permanent &&
		len(p.CardTypes) == 0 && len(p.Subtypes) == 0
}

func cloneEnchantPredicate(predicate EnchantPredicate) EnchantPredicate {
	predicate.CardTypes = slices.Clone(predicate.CardTypes)
	predicate.Subtypes = slices.Clone(predicate.Subtypes)
	return predicate
}

type keywordParameterDetails struct {
	ManaCost        cost.Mana           `json:",omitempty"`
	Integer         int                 `json:",omitempty"`
	EnchantTarget   EnchantPredicate    `json:",omitzero"`
	Protection      ProtectionParameter `json:",omitzero"`
	Gift            GiftKind            `json:",omitempty"`
	MobilizeDynamic MobilizeDynamicKind `json:",omitempty"`
}

// KeywordParameter is source-spanned typed syntax for one keyword parameter.
// Text is parser-owned canonical text retained for diagnostics and source-stable
// compiler metadata; semantic consumers use Kind and the typed accessors.
type KeywordParameter struct {
	Kind    KeywordParameterKind `json:",omitempty"`
	Span    shared.Span          `json:"-"`
	Text    string               `json:",omitempty"`
	details *keywordParameterDetails
}

// NewManaKeywordParameter constructs a typed mana-cost parameter.
func NewManaKeywordParameter(span shared.Span, manaCost cost.Mana) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterManaCost,
		Span:    span,
		Text:    manaCost.String(),
		details: &keywordParameterDetails{ManaCost: slices.Clone(manaCost)},
	}
}

// NewIntegerKeywordParameter constructs a typed integer parameter.
func NewIntegerKeywordParameter(span shared.Span, value int) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterInteger,
		Span:    span,
		Text:    strconv.Itoa(value),
		details: &keywordParameterDetails{Integer: value},
	}
}

// NewEnchantTargetKeywordParameter constructs a typed Enchant target parameter.
func NewEnchantTargetKeywordParameter(span shared.Span, target EnchantPredicate) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterEnchantTarget,
		Span:    span,
		Text:    enchantTargetName(target),
		details: &keywordParameterDetails{EnchantTarget: cloneEnchantPredicate(target)},
	}
}

// NewChampionKeywordParameter constructs a typed Champion type parameter. The
// predicate names the creature kind the keyword's enters-the-battlefield exile
// chooses ("Champion a creature", "Champion a Goblin", "Champion a Goblin or
// Shaman"); it reuses the disjunctive EnchantPredicate shape.
func NewChampionKeywordParameter(span shared.Span, target EnchantPredicate) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterChampion,
		Span:    span,
		Text:    enchantTargetName(target),
		details: &keywordParameterDetails{EnchantTarget: cloneEnchantPredicate(target)},
	}
}

// NewProtectionKeywordParameter constructs a typed Protection predicate.
func NewProtectionKeywordParameter(span shared.Span, text string, protection ProtectionParameter) KeywordParameter {
	return KeywordParameter{
		Kind: KeywordParameterProtection,
		Span: span,
		Text: text,
		details: &keywordParameterDetails{
			Protection: cloneProtectionParameter(protection),
		},
	}
}

// NewGiftKeywordParameter constructs a typed gift parameter naming the promised
// gift of a Gift keyword action (CR 702.171).
func NewGiftKeywordParameter(span shared.Span, kind GiftKind, text string) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterGift,
		Span:    span,
		Text:    text,
		details: &keywordParameterDetails{Gift: kind},
	}
}

// NewMobilizeDynamicKeywordParameter constructs a typed rules-derived Mobilize
// count parameter ("Mobilize X, where X is ..."), naming which dynamic count the
// keyword uses (CR 702.169).
func NewMobilizeDynamicKeywordParameter(span shared.Span, kind MobilizeDynamicKind, text string) KeywordParameter {
	return KeywordParameter{
		Kind:    KeywordParameterMobilizeDynamic,
		Span:    span,
		Text:    text,
		details: &keywordParameterDetails{MobilizeDynamic: kind},
	}
}

// ManaCost returns a copy of the typed mana-cost parameter.
func (p KeywordParameter) ManaCost() cost.Mana {
	if p.details == nil {
		return nil
	}
	return slices.Clone(p.details.ManaCost)
}

// Integer returns the typed integer parameter.
func (p KeywordParameter) Integer() int {
	if p.details == nil {
		return 0
	}
	return p.details.Integer
}

// EnchantTarget returns the typed Enchant target parameter.
func (p KeywordParameter) EnchantTarget() EnchantPredicate {
	if p.details == nil {
		return EnchantPredicate{}
	}
	return cloneEnchantPredicate(p.details.EnchantTarget)
}

// Protection returns a copy of the typed Protection predicate.
func (p KeywordParameter) Protection() ProtectionParameter {
	if p.details == nil {
		return ProtectionParameter{}
	}
	return cloneProtectionParameter(p.details.Protection)
}

// Gift returns the typed gift kind of a Gift keyword parameter.
func (p KeywordParameter) Gift() GiftKind {
	if p.details == nil {
		return GiftKindNone
	}
	return p.details.Gift
}

// MobilizeDynamic returns the typed rules-derived Mobilize count kind, or
// MobilizeDynamicNone when the parameter is not a dynamic Mobilize count.
func (p KeywordParameter) MobilizeDynamic() MobilizeDynamicKind {
	if p.details == nil {
		return MobilizeDynamicNone
	}
	return p.details.MobilizeDynamic
}

// Keyword is one source-spanned recognized keyword and its typed parameter.
type Keyword struct {
	Kind      KeywordKind      `json:",omitempty"`
	NameSpan  shared.Span      `json:"-"`
	Span      shared.Span      `json:"-"`
	Text      string           `json:",omitempty"`
	Parameter KeywordParameter `json:",omitzero"`
	// WardCost is the typed non-mana or composite payment of a "Ward—<cost>"
	// ability (CR 702.21), or nil for a mana-only Ward whose cost is carried by
	// Parameter. It models the em-dash forms "Ward—Pay N life.", "Ward—Sacrifice
	// a creature.", "Ward—Discard a card.", and the composite "Ward—{2}, Pay 2
	// life." Its components are the same comma-separated cost operations the
	// activated-ability cost parser recognizes.
	WardCost *Cost `json:",omitempty"`
	// EquipRestriction is the typed quality restriction on a restricted Equip
	// ability ("Equip legendary creature {3}", "Equip Knight {2}"), or nil for an
	// unrestricted Equip. The mana cost is still carried by Parameter.
	EquipRestriction *KeywordEquipRestriction `json:",omitempty"`
}

// KeywordEquipRestriction is the typed quality restriction on a restricted Equip
// ability: the Equipment may attach only to a creature that has every listed
// supertype and at least one of the listed subtypes (CR 301.5c). It models
// "Equip legendary creature {3}" (supertype Legendary) and "Equip <subtype>
// {N}" forms such as "Equip Knight {2}".
type KeywordEquipRestriction struct {
	Span       shared.Span `json:"-"`
	Supertypes []Supertype `json:",omitempty"`
	Subtypes   []types.Sub `json:",omitempty"`
	// Commander marks "Equip commander {cost}" (Commander's Plate): the Equipment
	// may attach only to a commander its controller controls. It is orthogonal to
	// the supertype/subtype quality restriction and is never combined with them.
	Commander bool `json:",omitempty"`
}

// KeywordSelectorForm identifies how a selector introduces its keyword.
type KeywordSelectorForm string

// Keyword-selector forms.
const (
	KeywordSelectorFormUnknown KeywordSelectorForm = ""
	KeywordSelectorFormDirect  KeywordSelectorForm = "KeywordSelectorFormDirect"
	KeywordSelectorFormAbility KeywordSelectorForm = "KeywordSelectorFormAbility"
)

// KeywordSelector is composable "with/without <keyword>" selector syntax.
type KeywordSelector struct {
	Keyword  KeywordKind         `json:",omitempty"`
	Form     KeywordSelectorForm `json:",omitempty"`
	Span     shared.Span         `json:"-"`
	Excluded bool                `json:",omitempty"`
}

// expandBushidoKeyword rewrites each printed "Bushido N" keyword line into the
// triggered ability it abbreviates: "Whenever this creature blocks or becomes
// blocked, it gets +N/+N until end of turn." (CR 702.46a). Bushido is pure
// shorthand for that combat trigger, so expanding it to canonical wording lets
// the standard trigger pipeline lower it. The rewrite is parser-owned because it
// is a wording substitution; downstream stages see only the expanded ability.
func expandBushidoKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := bushidoLineRank(line)
		if !ok {
			continue
		}
		lines[i] = "Whenever this creature blocks or becomes blocked, it gets +" +
			strconv.Itoa(rank) + "/+" + strconv.Itoa(rank) + " until end of turn."
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// bushidoLineRank reports the rank N of a line that is exactly the printed
// "Bushido N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, or pair it with other
// rules text, are left untouched.
func bushidoLineRank(line string) (int, bool) {
	const prefix = "Bushido "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// expandAnnihilatorKeyword rewrites each printed "Annihilator N" keyword line
// into the triggered ability it abbreviates: "Whenever this creature attacks,
// defending player sacrifices N permanents of their choice." (CR 702.85a, the
// Eldrazi keyword). Annihilator is pure shorthand for that combat trigger, so
// expanding it to canonical wording lets the standard trigger pipeline lower it.
// The rewrite is parser-owned because it is a wording substitution; downstream
// stages see only the expanded ability.
func expandAnnihilatorKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := annihilatorLineRank(line)
		if !ok {
			continue
		}
		lines[i] = annihilatorCanonicalText(rank)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// annihilatorCanonicalText is the triggered ability that the printed
// "Annihilator N" keyword abbreviates, with N spelled as its Oracle wording.
func annihilatorCanonicalText(rank int) string {
	if rank == 1 {
		return "Whenever this creature attacks, defending player sacrifices a permanent of their choice."
	}
	word, ok := cardinalWord(rank)
	if !ok {
		word = strconv.Itoa(rank)
	}
	return "Whenever this creature attacks, defending player sacrifices " + word + " permanents of their choice."
}

// annihilatorLineRank reports the rank N of a line that is exactly the printed
// "Annihilator N" keyword, optionally followed only by its parenthesized
// reminder text. Lines that merely contain the word elsewhere, or pair it with
// other rules text, are left untouched.
func annihilatorLineRank(line string) (int, bool) {
	const prefix = "Annihilator "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// expandAfflictKeyword rewrites each printed "Afflict N" keyword line into the
// triggered ability it abbreviates: "Whenever this creature becomes blocked,
// defending player loses N life." (CR 702.131). Afflict is pure shorthand for
// that combat trigger, so expanding it to canonical wording lets the standard
// trigger pipeline lower it. The rewrite is parser-owned because it is a wording
// substitution; downstream stages see only the expanded ability.
func expandAfflictKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := afflictLineRank(line)
		if !ok {
			continue
		}
		lines[i] = afflictCanonicalText(rank)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// expandFrenzyKeyword rewrites each printed "Frenzy N" keyword line into the
// triggered ability it abbreviates: "Whenever this creature attacks and isn't
// blocked, it gets +N/+0 until end of turn." (CR 702.35). Frenzy is pure
// shorthand for that unblocked-attacker combat trigger, so expanding it to
// canonical wording lets the standard trigger pipeline lower it. The rewrite is
// parser-owned because it is a wording substitution; downstream stages see only
// the expanded ability.
func expandFrenzyKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := frenzyLineRank(line)
		if !ok {
			continue
		}
		lines[i] = frenzyCanonicalText(rank)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// afflictCanonicalText is the triggered ability that the printed "Afflict N"
// keyword abbreviates. The life-loss amount is always written as a numeral, as
// in the printed reminder text.
func afflictCanonicalText(rank int) string {
	return "Whenever this creature becomes blocked, defending player loses " +
		strconv.Itoa(rank) + " life."
}

// afflictLineRank reports the rank N of a line that is exactly the printed
// "Afflict N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere (e.g. "creatures you
// control have afflict 2"), or pair it with other rules text, are left
// untouched.
func afflictLineRank(line string) (int, bool) {
	const prefix = "Afflict "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// frenzyCanonicalText is the triggered ability that the printed "Frenzy N"
// keyword abbreviates, with N spelled as its signed power bonus.
func frenzyCanonicalText(rank int) string {
	bonus := strconv.Itoa(rank)
	return "Whenever this creature attacks and isn't blocked, it gets +" + bonus + "/+0 until end of turn."
}

// frenzyLineRank reports the rank N of a line that is exactly the printed
// "Frenzy N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, or pair it with other
// rules text, are left untouched.
func frenzyLineRank(line string) (int, bool) {
	const prefix = "Frenzy "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// expandAfterlifeKeyword rewrites each printed "Afterlife N" keyword line into
// the dies-triggered token creation it abbreviates: "When this creature dies,
// create N 1/1 white and black Spirit creature tokens with flying." (CR
// 702.135). Afterlife is pure shorthand for that death trigger, so expanding it
// to canonical wording lets the standard trigger pipeline lower it. The rewrite
// is parser-owned because it is a wording substitution; downstream stages see
// only the expanded ability.
func expandAfterlifeKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := afterlifeLineRank(line)
		if !ok {
			continue
		}
		lines[i] = afterlifeCanonicalText(rank)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// afterlifeCanonicalText is the dies-triggered ability that the printed
// "Afterlife N" keyword abbreviates, with N spelled as its Oracle wording.
func afterlifeCanonicalText(rank int) string {
	if rank == 1 {
		return "When this creature dies, create a 1/1 white and black Spirit creature token with flying."
	}
	word, ok := cardinalWord(rank)
	if !ok {
		word = strconv.Itoa(rank)
	}
	return "When this creature dies, create " + word + " 1/1 white and black Spirit creature tokens with flying."
}

// afterlifeLineRank reports the rank N of a line that is exactly the printed
// "Afterlife N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, or pair it with other
// rules text, are left untouched.
func afterlifeLineRank(line string) (int, bool) {
	const prefix = "Afterlife "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// expandRenownKeyword rewrites each printed "Renown N" keyword line into the
// triggered ability it abbreviates. Renown is pure shorthand for a fixed
// triggered ability (CR 702.111), so expanding it to canonical wording lets the
// standard combat-damage trigger pipeline lower it. The "renown N" body is a
// keyword action whose runtime guard (it applies only once, when the permanent
// is not already renowned) subsumes the printed "if it isn't renowned"
// intervening-if, mirroring how Amass collapses its counter-placement wording.
// The rewrite is parser-owned because it is a wording substitution; downstream
// stages see only the expanded ability.
func expandRenownKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := renownLineRank(line)
		if !ok {
			continue
		}
		lines[i] = renownCanonicalText(rank)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// renownCanonicalText is the triggered ability that the printed "Renown N"
// keyword abbreviates, with the renown keyword action carrying the rank N.
func renownCanonicalText(rank int) string {
	return "When this creature deals combat damage to a player, renown " + strconv.Itoa(rank) + "."
}

// renownLineRank reports the rank N of a line that is exactly the printed
// "Renown N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, or pair it with other
// rules text, are left untouched.
func renownLineRank(line string) (int, bool) {
	const prefix = "Renown "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// extortCanonicalText is the triggered ability that the printed "Extort" keyword
// abbreviates (CR 702.99a).
const extortCanonicalText = "Whenever you cast a spell, you may pay {W/B}. " +
	"If you do, each opponent loses 1 life and you gain that much life."

// expandExtortKeyword rewrites each printed "Extort" keyword line into the
// triggered ability it abbreviates. Like Bushido, Extort is pure shorthand for a
// fixed triggered ability, so expanding it to canonical wording lets the standard
// trigger pipeline lower it. Multiple printed instances each expand to their own
// trigger, matching the rule that each Extort instance triggers separately. The
// rewrite is parser-owned because it is a wording substitution; downstream stages
// see only the expanded ability.
func expandExtortKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		if !isExtortKeywordLine(line) {
			continue
		}
		lines[i] = extortCanonicalText
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// isExtortKeywordLine reports whether a line is exactly the printed "Extort"
// keyword, optionally followed only by its parenthesized reminder text. Lines
// that merely contain the word elsewhere, or pair it with other rules text, are
// left untouched.
func isExtortKeywordLine(line string) bool {
	const keyword = "Extort"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	tail := strings.TrimSpace(trimmed[len(keyword):])
	if tail == "" {
		return true
	}
	return strings.HasPrefix(tail, "(") && strings.HasSuffix(tail, ")")
}

// modularLineRank reports the rank N of a line that is exactly the printed
// "Modular N" keyword, optionally followed only by its parenthesized reminder
// text. Lines that merely contain the word elsewhere, that pair it with other
// rules text, or that use a variable form ("Modular—Sunburst") are left
// untouched.
func modularLineRank(line string) (int, bool) {
	const prefix = "Modular "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// modularCounterPhrase spells the enters-with-counters quantity for Modular rank
// N as Oracle text ("a +1/+1 counter", "two +1/+1 counters"). It fails closed
// for ranks outside the small-cardinal vocabulary the enters-with-counters
// static can spell.
func modularCounterPhrase(rank int) (string, bool) {
	if rank == 1 {
		return "a +1/+1 counter", true
	}
	word, ok := cardinalNumberWord(rank)
	if !ok {
		return "", false
	}
	return word + " +1/+1 counters", true
}

// cardinalNumberWord spells a small positive integer as its Oracle cardinal word
// ("two" … "ten"), the inverse of CardinalWordValue for the values a keyword
// expansion needs. It fails closed outside that range.
func cardinalNumberWord(n int) (string, bool) {
	switch n {
	case 2:
		return "two", true
	case 3:
		return "three", true
	case 4:
		return "four", true
	case 5:
		return "five", true
	case 6:
		return "six", true
	case 7:
		return "seven", true
	case 8:
		return "eight", true
	case 9:
		return "nine", true
	case 10:
		return "ten", true
	default:
		return "", false
	}
}

// expandModularKeyword rewrites each printed "Modular N" keyword line into the
// two abilities it abbreviates (CR 702.43c): a static placing N +1/+1 counters as
// the creature enters, and a dies-trigger that moves those counters onto a target
// artifact creature. Like Bushido and Extort, Modular is pure shorthand for fixed
// abilities, so expanding it to canonical wording lets the standard
// enters-with-counters and trigger pipelines lower it. The rewrite is
// parser-owned because it is a wording substitution; downstream stages see only
// the expanded abilities.
func expandModularKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := modularLineRank(line)
		if !ok {
			continue
		}
		counters, ok := modularCounterPhrase(rank)
		if !ok {
			continue
		}
		lines[i] = "This creature enters with " + counters + " on it.\n" +
			"When this creature dies, you may move all +1/+1 counters from this creature " +
			"onto target artifact creature."
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// groupModularSubject reports the group subject and rank N of a printed
// "<subject> have modular N" / "<subject> has modular N" line, optionally
// followed only by its parenthesized reminder text. Only the group-granted form
// qualifies; the self-form "Modular N" keyword and unrelated text return false so
// they are left for expandModularKeyword and the normal pipeline.
func groupModularSubject(line string) (subject string, rank int, ok bool) {
	trimmed := strings.TrimSpace(line)
	for _, verb := range []string{" have modular ", " has modular "} {
		before, rest, found := strings.Cut(trimmed, verb)
		if !found {
			continue
		}
		subject = strings.TrimSpace(before)
		digits := 0
		for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
			digits++
		}
		if digits == 0 {
			return "", 0, false
		}
		value, err := strconv.Atoi(rest[:digits])
		if err != nil || value <= 0 {
			return "", 0, false
		}
		tail := strings.TrimSpace(rest[digits:])
		if !strings.HasPrefix(tail, ".") {
			return "", 0, false
		}
		reminder := strings.TrimSpace(tail[1:])
		if reminder != "" && (!strings.HasPrefix(reminder, "(") || !strings.HasSuffix(reminder, ")")) {
			return "", 0, false
		}
		return subject, value, true
	}
	return "", 0, false
}

// groupModularUnionSubject rewrites a coordinated group subject joined by "and"
// into the equivalent disjunctive ("or") subject that the selection pipeline
// lowers as a union, distributing a shared leading "nontoken" qualifier onto the
// second noun so both members carry it. "X and Y have Z" and "X or Y have Z" name
// the same group when granting a shared ability, so the rewrite is meaning
// preserving. It fails closed for any subject that is not a single-"and"
// controlled group.
func groupModularUnionSubject(subject string) (string, bool) {
	const suffix = " you control"
	if !strings.HasSuffix(subject, suffix) {
		return "", false
	}
	core := strings.TrimSuffix(subject, suffix)
	parts := strings.Split(core, " and ")
	if len(parts) != 2 {
		return "", false
	}
	left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if left == "" || right == "" {
		return "", false
	}
	if subjectContainsWord(left, "nontoken") && !subjectContainsWord(right, "nontoken") {
		right = "nontoken " + right
	}
	return left + " or " + right + suffix, true
}

// subjectContainsWord reports whether text contains word as a whole,
// punctuation-trimmed lexeme, case-insensitively.
func subjectContainsWord(text, word string) bool {
	for field := range strings.FieldsSeq(text) {
		if strings.EqualFold(strings.Trim(field, ",."), word) {
			return true
		}
	}
	return false
}

// lowerFirstASCIILetter lowercases only a leading ASCII capital so a sentence
// subject can be spliced mid-sentence ("Other …" → "other …").
func lowerFirstASCIILetter(text string) string {
	if text == "" {
		return text
	}
	if b := text[0]; b >= 'A' && b <= 'Z' {
		return string(b-'A'+'a') + text[1:]
	}
	return text
}

// expandGroupModularKeyword rewrites a printed "<group> have modular N" line —
// the group-granted form of Modular carried by cards like Blaster, Combat DJ —
// into the two abilities it grants each member: entering with an additional
// +1/+1 counter, and the Modular dies-trigger that moves its counters onto a
// target artifact creature. Like the self-form expandModularKeyword this is a
// parser-owned wording substitution, letting the standard enters-with-counters
// and quoted-ability-grant pipelines lower it. Only rank 1 is expanded because
// the group enters-with-counters static spells a single additional counter;
// other ranks and non-group subjects are left untouched.
func expandGroupModularKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		subject, rank, ok := groupModularSubject(line)
		if !ok || rank != 1 {
			continue
		}
		unionSubject, ok := groupModularUnionSubject(subject)
		if !ok {
			continue
		}
		lines[i] = "Each " + lowerFirstASCIILetter(unionSubject) +
			" enters with an additional +1/+1 counter on it.\n" +
			unionSubject + ` have "When this creature dies, you may move all ` +
			`+1/+1 counters from this creature onto target artifact creature."`
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}
func graftLineRank(line string) (int, bool) {
	const prefix = "Graft "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}

// expandGraftKeyword rewrites each printed "Graft N" keyword line into the two
// abilities it abbreviates (CR 702.57): a static placing N +1/+1 counters as the
// creature enters, and a trigger that may move a +1/+1 counter from this creature
// onto another creature as that creature enters. Like Modular, Graft is pure
// shorthand for fixed abilities, so expanding it to canonical wording lets the
// standard enters-with-counters and trigger pipelines lower it; "that creature"
// names the triggering permanent so the moved counter lands on the entering
// creature rather than the source. The rewrite is parser-owned because it is a
// wording substitution; downstream stages see only the expanded abilities.
func expandGraftKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		rank, ok := graftLineRank(line)
		if !ok {
			continue
		}
		counters, ok := modularCounterPhrase(rank)
		if !ok {
			continue
		}
		lines[i] = "This creature enters with " + counters + " on it.\n" +
			"Whenever another creature enters, you may move a +1/+1 counter " +
			"from this creature onto that creature."
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// expandAffinityKeyword rewrites each printed "Affinity for <permanents>"
// keyword line into the static cast cost reduction it abbreviates (CR 702.41a):
// "This spell costs {1} less to cast for each <permanent> you control." Affinity
// is pure shorthand for that self cost reduction, so expanding it to canonical
// wording lets the standard source-spell cost-reduction pipeline lower it
// without a dedicated keyword path. The rewrite is parser-owned because it is a
// wording substitution; downstream stages see only the expanded ability. A line
// whose noun cannot be singularized into a countable subject the downstream
// pipeline recognizes is still rewritten and simply fails closed there, so it
// stays unsupported rather than silently dropping the reduction.
func expandAffinityKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		noun, ok := affinityLineNoun(line)
		if !ok {
			continue
		}
		lines[i] = "This spell costs {1} less to cast for each " +
			affinitySingularNoun(noun) + " you control."
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// affinityLineNoun reports the plural permanent noun of a line that is exactly
// the printed "Affinity for <noun>" keyword, optionally followed only by its
// parenthesized reminder text. Lines that merely contain the word elsewhere, or
// pair it with other rules text, are left untouched. The noun must be a simple
// word phrase so a malformed line fails closed.
func affinityLineNoun(line string) (string, bool) {
	const prefix = "Affinity for "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	if open := strings.Index(rest, "("); open >= 0 {
		reminder := strings.TrimSpace(rest[open:])
		if !strings.HasPrefix(reminder, "(") || !strings.HasSuffix(reminder, ")") {
			return "", false
		}
		rest = strings.TrimSpace(rest[:open])
	}
	if rest == "" {
		return "", false
	}
	for _, r := range rest {
		if r != ' ' && (r < 'A' || r > 'z' || (r > 'Z' && r < 'a')) {
			return "", false
		}
	}
	return rest, true
}

// affinitySingularNoun singularizes the head word of an Affinity noun phrase so
// the canonical "for each <noun> you control" wording counts a single permanent
// kind ("artifacts" → "artifact", "Allies" → "Ally", "Forests" → "Forest"). Only
// the trailing head noun is singularized; any leading qualifiers ("artifact" in
// "artifact creatures", "snow" in "snow lands") are preserved. "Plains" is both
// singular and plural and is left unchanged.
func affinitySingularNoun(plural string) string {
	fields := strings.Fields(plural)
	if len(fields) == 0 {
		return plural
	}
	head := fields[len(fields)-1]
	switch {
	case head == "Plains":
	case strings.HasSuffix(head, "ss"):
	case len(head) > 4 && strings.HasSuffix(head, "ies"):
		head = strings.TrimSuffix(head, "ies") + "y"
	case len(head) > 1 && strings.HasSuffix(head, "s"):
		head = strings.TrimSuffix(head, "s")
	default:
	}
	fields[len(fields)-1] = head
	return strings.Join(fields, " ")
}

// battleCryCanonicalText is the triggered ability that the printed "Battle cry"
// keyword abbreviates (CR 702.91a).
const battleCryCanonicalText = "Whenever this creature attacks, " +
	"each other attacking creature gets +1/+0 until end of turn."

// expandBattleCryKeyword rewrites each printed "Battle cry" keyword line into the
// triggered ability it abbreviates. Like Extort, Battle cry is pure shorthand for
// a fixed triggered ability, so expanding it to canonical wording lets the
// standard trigger pipeline lower it. The rewrite is parser-owned because it is a
// wording substitution; downstream stages see only the expanded ability.
func expandBattleCryKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		if !isBattleCryKeywordLine(line) {
			continue
		}
		lines[i] = battleCryCanonicalText
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// isBattleCryKeywordLine reports whether a line is exactly the printed "Battle
// cry" keyword, optionally followed only by its parenthesized reminder text.
// Lines that merely contain the words elsewhere, or pair the keyword with other
// rules text (such as a sticker-cost prefix), are left untouched.
func isBattleCryKeywordLine(line string) bool {
	const keyword = "Battle cry"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	tail := strings.TrimSpace(trimmed[len(keyword):])
	if tail == "" {
		return true
	}
	return strings.HasPrefix(tail, "(") && strings.HasSuffix(tail, ")")
}

// mentorCanonicalText is the full Oracle wording Mentor abbreviates (CR 702.123).
const mentorCanonicalText = "Whenever this creature attacks, " +
	"put a +1/+1 counter on target attacking creature with lesser power."

// expandMentorKeyword rewrites each bare "Mentor" keyword line into its full
// triggered-ability Oracle text so the existing attacks-trigger and counter
// placement pipeline lowers it. Parser owns the wording.
func expandMentorKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		if !isMentorKeywordLine(line) {
			continue
		}
		lines[i] = mentorCanonicalText
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// isMentorKeywordLine reports whether a line is exactly the printed "Mentor"
// keyword, optionally followed only by its parenthesized reminder text. Lines
// that merely contain the word elsewhere, or pair it with other rules text, are
// left untouched.
func isMentorKeywordLine(line string) bool {
	const keyword = "Mentor"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	tail := strings.TrimSpace(trimmed[len(keyword):])
	if tail == "" {
		return true
	}
	return strings.HasPrefix(tail, "(") && strings.HasSuffix(tail, ")")
}

// meleeCanonicalText is the full Oracle wording Melee abbreviates (CR 702.72).
const meleeCanonicalText = "Whenever this creature attacks, it gets +1/+1 until " +
	"end of turn for each opponent you attacked this combat."

// expandMeleeKeyword rewrites each bare "Melee" keyword line into its full
// triggered-ability Oracle text so the existing attacks-trigger and dynamic
// power/toughness pipeline lowers it. The "for each opponent you attacked this
// combat" count resolves from the current combat's attack declarations. Parser
// owns the wording.
func expandMeleeKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		if !isMeleeKeywordLine(line) {
			continue
		}
		lines[i] = meleeCanonicalText
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// isMeleeKeywordLine reports whether a line is exactly the printed "Melee"
// keyword, optionally followed only by its parenthesized reminder text. Lines
// that merely contain the word elsewhere, or pair it with other rules text, are
// left untouched.
func isMeleeKeywordLine(line string) bool {
	const keyword = "Melee"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, keyword) {
		return false
	}
	tail := strings.TrimSpace(trimmed[len(keyword):])
	if tail == "" {
		return true
	}
	return strings.HasPrefix(tail, "(") && strings.HasSuffix(tail, ")")
}

func scanKeywords(tokens []shared.Token, atoms Atoms) []Keyword {
	var keywords []Keyword
	for i := 0; i < len(tokens); i++ {
		kind, width, ok := recognizeKeywordNameAt(tokens, i)
		if !ok {
			continue
		}
		nameSpan := shared.SpanOf(tokens[i : i+width])
		// A keyword word that falls inside an occurrence of the card's own name
		// (e.g. "Storm" in "Command the Storm") is part of the name, not a
		// granted ability keyword, so it must not be scanned as one.
		if atoms.SelfNameAt(nameSpan) {
			i += width - 1
			continue
		}
		// A keyword word immediately followed by "counter(s)" names a keyword
		// counter (CR 122.1c), e.g. "a vigilance counter" / "a flying counter",
		// not a granted keyword ability, so it must not be scanned as one.
		if i+width < len(tokens) &&
			(equalWord(tokens[i+width], "counter") || equalWord(tokens[i+width], "counters")) {
			i += width - 1
			continue
		}
		// "flash" in the cast-permission idiom "as though they had flash" (or
		// "... it had flash") names the timing reference, not a granted Flash
		// keyword, so it is parsed by the cast-as-though-flash static instead.
		if kind == KeywordFlash && i > 0 && equalWord(tokens[i-1], "had") {
			continue
		}
		// "haste" in the activation-permission idiom "as though those creatures
		// had haste" (or "... it had haste") names the timing reference, not a
		// granted Haste keyword, so it is parsed by the activate-abilities-as-
		// though-haste static instead.
		if kind == KeywordHaste && i > 0 && equalWord(tokens[i-1], "had") {
			continue
		}
		// The Gift keyword action names its promised gift ("Gift a card", "Gift a
		// Food", "Gift a Treasure", "Gift a tapped Fish"; CR 702.171). The bare
		// word "gift" also appears in the per-effect condition "the gift was
		// promised", so the keyword is recognized only when a typed gift
		// parameter follows; otherwise the word is left for the condition parser.
		if kind == KeywordGift {
			parameter, giftEnd, ok := parseGiftKeywordParameter(tokens, i+width)
			if !ok {
				continue
			}
			keywords = append(keywords, Keyword{
				Kind:      KeywordGift,
				NameSpan:  nameSpan,
				Span:      shared.SpanOf(tokens[i:giftEnd]),
				Text:      joinTokens(tokens[i:giftEnd]),
				Parameter: parameter,
			})
			i = giftEnd - 1
			continue
		}
		// "Splice onto Arcane" is supported only in its printed mana-cost form
		// ("Splice onto Arcane {1}{R}"; CR 702.47). The keyword is recognized
		// only when a mana cost follows the name; the em-dash nonmana form
		// ("Splice onto Arcane—Exile ...") and any other variant produce no
		// keyword and stay unsupported (fail closed).
		if kind == KeywordSplice {
			manaCost, spliceEnd, ok := parseKeywordManaCost(tokens, i+width)
			if !ok {
				continue
			}
			keywords = append(keywords, Keyword{
				Kind:      KeywordSplice,
				NameSpan:  nameSpan,
				Span:      shared.SpanOf(tokens[i:spliceEnd]),
				Text:      joinTokens(tokens[i:spliceEnd]),
				Parameter: NewManaKeywordParameter(shared.SpanOf(tokens[i+width:spliceEnd]), manaCost),
			})
			i = spliceEnd - 1
			continue
		}
		end := i + width
		// The Echo keyword (CR 702.29) is recognized only when a fixed mana echo
		// cost follows ("Echo {3}{W}{W}"). The non-mana em-dash forms
		// ("Echo—Discard a card.", "Echo—Sacrifice two lands.") carry no mana
		// parameter, so they are left unrecognized here and fail closed as an
		// unsupported ability rather than being misread. The word also appears in
		// flavored ability names ("Echo of the First Murder —"), which likewise
		// have no mana cost following the word and so are not scanned as a keyword.
		if kind == KeywordEcho {
			if _, _, ok := parseKeywordManaCost(tokens, end); !ok {
				continue
			}
		}
		// The Bestow keyword (CR 702.103) is recognized only when a fixed mana
		// bestow cost follows ("Bestow {1}{G}"). The non-mana em-dash form
		// ("Bestow—<cost>", e.g. "Bestow—{R}, Collect evidence 6") carries no
		// mana parameter, so it is left unrecognized here and fails closed as an
		// unsupported ability. Variable ({X}) bestow costs do parse as a mana
		// cost here, but are rejected later during lowering, which supports only
		// fixed bestow costs.
		if kind == KeywordBestow {
			if _, _, ok := parseKeywordManaCost(tokens, end); !ok {
				continue
			}
		}
		var equipRestriction *KeywordEquipRestriction
		if kind == KeywordEquip {
			if restriction, manaStart, ok := parseEquipRestriction(tokens, end, atoms); ok {
				equipRestriction = restriction
				end = manaStart
			}
		}
		if kind == KeywordMobilize {
			if parameter, dynamicEnd, ok := parseMobilizeDynamicParameter(tokens, end); ok {
				keywords = append(keywords, Keyword{
					Kind:      KeywordMobilize,
					NameSpan:  nameSpan,
					Span:      shared.SpanOf(tokens[i:dynamicEnd]),
					Text:      joinTokens(tokens[i:dynamicEnd]),
					Parameter: parameter,
				})
				i = dynamicEnd - 1
				continue
			}
		}
		parameter, parameterEnd := parseKeywordParameter(kind, tokens, end, atoms)
		end = parameterEnd
		if kind == KeywordWard {
			end = wardEmDashCostEnd(tokens, end)
		}
		keywords = append(keywords, Keyword{
			Kind:             kind,
			NameSpan:         nameSpan,
			Span:             shared.SpanOf(tokens[i:end]),
			Text:             joinTokens(tokens[i:end]),
			Parameter:        parameter,
			EquipRestriction: equipRestriction,
		})
		i = end - 1
	}
	return keywords
}

// wardEmDashCostEnd extends a Ward keyword atom past the em dash and the
// non-mana or composite cost clause that follows it ("Ward—Pay N life.",
// "Ward—{2}, Pay 2 life."), returning the token index after the cost so the
// keyword span covers its whole printed cost (CR 702.21). It stops before the
// trailing top-level period and is a no-op when no em dash follows the keyword
// name, leaving the mana-only "Ward {N}" form untouched.
func wardEmDashCostEnd(tokens []shared.Token, start int) int {
	if start >= len(tokens) || tokens[start].Kind != shared.EmDash {
		return start
	}
	end := start + 1
	for end < len(tokens) && tokens[end].Kind != shared.Period {
		end++
	}
	if end == start+1 {
		return start
	}
	return end
}

func recognizeKeywordNameAt(tokens []shared.Token, start int) (KeywordKind, int, bool) {
	for _, grammar := range keywordNameGrammars {
		if atomWordsAt(tokens, start, grammar.Words...) {
			return grammar.Kind, len(grammar.Words), true
		}
	}
	return KeywordUnknown, 0, false
}

func parseKeywordParameter(
	kind KeywordKind,
	tokens []shared.Token,
	start int,
	atoms Atoms,
) (parameter KeywordParameter, end int) {
	switch kind {
	case KeywordProtection:
		return parseProtectionKeywordParameter(tokens, start, atoms)
	case KeywordHexproof:
		return parseHexproofKeywordParameter(tokens, start, atoms)
	case KeywordEnchant:
		if predicate, end, ok := parseEnchantTargetPredicate(tokens, start, atoms); ok {
			return NewEnchantTargetKeywordParameter(shared.SpanOf(tokens[start:end]), predicate), end
		}
		return KeywordParameter{}, start
	case KeywordChampion:
		typeStart := start
		if typeStart < len(tokens) && (equalWord(tokens[typeStart], "a") || equalWord(tokens[typeStart], "an")) {
			typeStart++
		}
		if predicate, end, ok := parseEnchantTargetPredicate(tokens, typeStart, atoms); ok {
			return NewChampionKeywordParameter(shared.SpanOf(tokens[start:end]), predicate), end
		}
		return KeywordParameter{}, start
	default:
	}
	if manaCost, end, ok := parseKeywordManaCost(tokens, start); ok {
		return NewManaKeywordParameter(shared.SpanOf(tokens[start:end]), manaCost), end
	}
	if start < len(tokens) && tokens[start].Kind == shared.Integer {
		value, err := strconv.Atoi(tokens[start].Text)
		if err == nil {
			return NewIntegerKeywordParameter(tokens[start].Span, value), start + 1
		}
	}
	return KeywordParameter{}, start
}

// parseGiftKeywordParameter recognizes the typed gift a Gift keyword action
// promises (CR 702.171): "a card" (the opponent draws a card), "a Food", "a
// Treasure", or "a tapped Fish" (a tapped 1/1 blue Fish creature token). It
// returns ok=false when no recognized gift form follows, so the bare word
// "gift" in the per-effect condition "the gift was promised" is never mistaken
// for the keyword and an unsupported gift form stays unsupported.
func parseGiftKeywordParameter(tokens []shared.Token, start int) (KeywordParameter, int, bool) {
	i := start
	if i >= len(tokens) || (!equalWord(tokens[i], "a") && !equalWord(tokens[i], "an")) {
		return KeywordParameter{}, start, false
	}
	i++
	if i >= len(tokens) {
		return KeywordParameter{}, start, false
	}
	giftParam := func(kind GiftKind, end int) (KeywordParameter, int, bool) {
		span := shared.SpanOf(tokens[start:end])
		return NewGiftKeywordParameter(span, kind, joinTokens(tokens[start:end])), end, true
	}
	switch {
	case equalWord(tokens[i], "card"):
		return giftParam(GiftKindCard, i+1)
	case equalWord(tokens[i], "Food"):
		return giftParam(GiftKindFood, i+1)
	case equalWord(tokens[i], "Treasure"):
		return giftParam(GiftKindTreasure, i+1)
	case equalWord(tokens[i], "tapped") && i+1 < len(tokens) && equalWord(tokens[i+1], "Fish"):
		return giftParam(GiftKindTappedFish, i+2)
	}
	return KeywordParameter{}, start, false
}

// parseMobilizeDynamicParameter recognizes the rules-derived Mobilize count
// "X, where X is the number of creature cards in your graveyard" (Avenger of the
// Fallen, CR 702.169), returning a typed dynamic parameter that spans the whole
// clause so keywordOnlyCovered still reports the keyword as covering its text.
// It returns ok=false for the fixed "Mobilize N" form (handled by the integer
// parameter path) and for any other dynamic wording, which then fails closed in
// lowering rather than being silently misread.
func parseMobilizeDynamicParameter(tokens []shared.Token, start int) (KeywordParameter, int, bool) {
	if start >= len(tokens) || !equalWord(tokens[start], "X") {
		return KeywordParameter{}, start, false
	}
	i := start + 1
	if i >= len(tokens) || tokens[i].Kind != shared.Comma {
		return KeywordParameter{}, start, false
	}
	i++
	words := []string{"where", "X", "is", "the", "number", "of", "creature", "cards", "in", "your", "graveyard"}
	if !atomWordsAt(tokens, i, words...) {
		return KeywordParameter{}, start, false
	}
	end := i + len(words)
	span := shared.SpanOf(tokens[start:end])
	parameter := NewMobilizeDynamicKeywordParameter(span, MobilizeDynamicCreatureCardsInGraveyard, joinTokens(tokens[start:end]))
	return parameter, end, true
}

// ability ("Equip legendary creature {3}", "Equip Knight {2}", "Equip Shaman,
// Warlock, or Wizard {2}") between the Equip keyword and its mana cost. It
// consumes supertype, subtype, and the implied "creature" card-type words (plus
// list separators), returning the typed restriction and the index of the
// following mana symbol. It fails closed (ok=false) when there is no restriction
// quality, when an unrecognized word appears, or when no mana cost follows, so
// an unsupported restricted Equip stays unsupported rather than silently
// dropping the restriction.
func parseEquipRestriction(tokens []shared.Token, start int, atoms Atoms) (*KeywordEquipRestriction, int, bool) {
	restriction := &KeywordEquipRestriction{}
	j := start
	for j < len(tokens) {
		token := tokens[j]
		if token.Kind == shared.Symbol {
			break
		}
		if token.Kind == shared.Comma || equalWord(token, "or") {
			j++
			continue
		}
		if supertype, ok := atoms.SupertypeAt(token.Span); ok {
			restriction.Supertypes = append(restriction.Supertypes, supertype)
			j++
			continue
		}
		if subtype, ok := atoms.SubtypeAt(token.Span); ok {
			restriction.Subtypes = append(restriction.Subtypes, subtype)
			j++
			continue
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok && cardType == CardTypeCreature {
			j++
			continue
		}
		if equalWord(token, "commander") {
			restriction.Commander = true
			j++
			continue
		}
		return nil, start, false
	}
	if len(restriction.Supertypes) == 0 && len(restriction.Subtypes) == 0 &&
		!restriction.Commander {
		return nil, start, false
	}
	if j >= len(tokens) || tokens[j].Kind != shared.Symbol {
		return nil, start, false
	}
	restriction.Span = shared.SpanOf(tokens[start:j])
	return restriction, j, true
}

// parseEnchantTargetPredicate recognizes the object restriction following an
// Enchant keyword: a single player word ("player", "opponent"), the
// any-permanent word ("permanent"), or a disjunctive list of permanent card
// types and subtypes ("creature", "artifact or creature", "creature, artifact,
// or land", "Forest", "creature or Vehicle"). It consumes only the recognized
// predicate tokens and returns the index after the last one, so any trailing
// qualifier the executable backend does not support (a controller, color,
// power, or zone restriction) is left uncovered and the Enchant ability fails
// closed downstream. It returns ok=false when the first token is not a
// recognized predicate word, so an unrecognized restriction stays unsupported.
func parseEnchantTargetPredicate(tokens []shared.Token, start int, atoms Atoms) (EnchantPredicate, int, bool) {
	if start >= len(tokens) {
		return EnchantPredicate{}, start, false
	}
	switch {
	case equalWord(tokens[start], "player"):
		return EnchantPredicate{Player: true}, start + 1, true
	case equalWord(tokens[start], "opponent"):
		return EnchantPredicate{Opponent: true}, start + 1, true
	case equalWord(tokens[start], "permanent"):
		return EnchantPredicate{Permanent: true}, start + 1, true
	}
	predicate := EnchantPredicate{}
	end := start
	// items requires a separator (comma or "or") between consecutive type and
	// subtype words. Adjacent words without a separator are a single conjunctive
	// type line ("artifact creature" = an artifact creature), which a disjunctive
	// predicate cannot represent, so the second word is left uncovered to fail
	// closed rather than silently widened to a disjunction.
	expectItem := true
	for i := start; i < len(tokens); {
		token := tokens[i]
		// A comma or "or" separates list items; it is meaningful only between
		// recognized words, so end does not advance past a trailing separator.
		if token.Kind == shared.Comma || equalWord(token, "or") {
			expectItem = true
			i++
			continue
		}
		if !expectItem {
			break
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok {
			// The Enchant grammar uses singular nouns ("Enchant creature"); the
			// atom scanner also normalizes plurals, so reject a non-singular form
			// ("Enchant creatures") by leaving it uncovered to fail closed.
			if word, ok := cardTypeWord(cardType); ok && strings.EqualFold(token.Text, word) {
				predicate.CardTypes = append(predicate.CardTypes, cardType)
				expectItem = false
				i++
				end = i
				continue
			}
			break
		}
		if subtype, ok := atoms.SubtypeAt(token.Span); ok {
			if strings.EqualFold(token.Text, string(subtype)) {
				predicate.Subtypes = append(predicate.Subtypes, subtype)
				expectItem = false
				i++
				end = i
				continue
			}
			break
		}
		break
	}
	if predicate.Empty() {
		return EnchantPredicate{}, start, false
	}
	// A trailing "card in a graveyard" narrows a card-type predicate to a
	// matching card in a graveyard ("Enchant creature card in a graveyard").
	// This is the graveyard-card Aura enchant restriction of the reanimation
	// Aura family; it is consumed only after a recognized card-type predicate so
	// the keyword span covers the whole restriction.
	if !expectItem && end+3 < len(tokens) &&
		equalWord(tokens[end], "card") &&
		equalWord(tokens[end+1], "in") &&
		equalWord(tokens[end+2], "a") &&
		equalWord(tokens[end+3], "graveyard") {
		predicate.InGraveyard = true
		end += 4
		return predicate, end, true
	}
	// A trailing "you control" controller restriction narrows the permanent
	// predicate to the enchanting player's own permanents ("Enchant creature or
	// planeswalker you control"). It is consumed only after a recognized
	// card-type/subtype predicate so the keyword span covers the whole
	// restriction; an unrecognized trailing qualifier is left uncovered to fail
	// closed downstream.
	if end+1 < len(tokens) && equalWord(tokens[end], "you") && equalWord(tokens[end+1], "control") {
		predicate.YouControl = true
		end += 2
	}
	return predicate, end, true
}

// enchantTargetName renders the parser-canonical display text for an Enchant
// target predicate, retained on the keyword parameter for diagnostics.
func enchantTargetName(predicate EnchantPredicate) string {
	switch {
	case predicate.Player:
		return "player"
	case predicate.Opponent:
		return "opponent"
	case predicate.Permanent:
		return "permanent"
	}
	words := make([]string, 0, len(predicate.CardTypes)+len(predicate.Subtypes))
	for _, cardType := range predicate.CardTypes {
		if word, ok := cardTypeWord(cardType); ok {
			words = append(words, word)
		}
	}
	for _, subtype := range predicate.Subtypes {
		words = append(words, strings.ToLower(string(subtype)))
	}
	name := strings.Join(words, " or ")
	if predicate.InGraveyard {
		name += " card in a graveyard"
	}
	if predicate.YouControl {
		name += " you control"
	}
	return name
}

func parseKeywordManaCost(tokens []shared.Token, start int) (cost.Mana, int, bool) {
	end := start
	var result cost.Mana
	for end < len(tokens) && tokens[end].Kind == shared.Symbol {
		symbol, ok := parseKeywordManaSymbol(tokens[end].Text)
		if !ok {
			return nil, start, false
		}
		result = append(result, symbol)
		end++
	}
	return result, end, len(result) > 0
}

func parseKeywordManaSymbol(text string) (cost.Symbol, bool) {
	symbol, ok := strings.CutPrefix(text, "{")
	if !ok {
		return cost.Symbol{}, false
	}
	symbol, ok = strings.CutSuffix(symbol, "}")
	if !ok {
		return cost.Symbol{}, false
	}
	switch symbol {
	case "X":
		return cost.X, true
	case "C":
		return cost.C, true
	case "S":
		return cost.S, true
	case "W":
		return cost.W, true
	case "U":
		return cost.U, true
	case "B":
		return cost.B, true
	case "R":
		return cost.R, true
	case "G":
		return cost.G, true
	default:
	}
	if value, err := strconv.Atoi(symbol); err == nil {
		return cost.O(value), true
	}
	if colorName, phyrexian := strings.CutSuffix(symbol, "/P"); phyrexian {
		color, colorOK := keywordManaColor(colorName)
		if colorOK {
			return cost.PhyrexianMana(color), true
		}
		return cost.Symbol{}, false
	}
	first, second, hybrid := strings.Cut(symbol, "/")
	if !hybrid {
		return cost.Symbol{}, false
	}
	if first == "2" {
		color, colorOK := keywordManaColor(second)
		if colorOK {
			return cost.Twobrid(color), true
		}
		return cost.Symbol{}, false
	}
	firstColor, firstOK := keywordManaColor(first)
	secondColor, secondOK := keywordManaColor(second)
	if !firstOK || !secondOK {
		return cost.Symbol{}, false
	}
	return cost.HybridMana(firstColor, secondColor), true
}

func keywordManaColor(name string) (mana.Color, bool) {
	switch name {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	default:
		return "", false
	}
}

// parseHexproofKeywordParameter recognizes the source-color qualifier on
// "hexproof from <colors>" ("hexproof from black", "hexproof from blue and from
// black"), reusing the protection color-list grammar. Only the color form is
// recognized; a bare "hexproof" with no "from" qualifier falls through to a
// non-parameterized simple keyword. The colors are carried in a protection
// parameter so the compiler reuses compileProtectionKeyword; lowering reads
// them back for a HexproofFromKeyword grant.
func parseHexproofKeywordParameter(
	tokens []shared.Token,
	start int,
	atoms Atoms,
) (parameter KeywordParameter, end int) {
	if start+1 >= len(tokens) || !equalWord(tokens[start], "from") {
		return KeywordParameter{}, start
	}
	if colors, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (Color, bool) {
		return atoms.ColorAt(token.Span)
	}); ok {
		names := make([]string, len(colors))
		for i, c := range colors {
			names[i] = colorName(c)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			strings.Join(names, ","),
			ProtectionParameter{FromColors: colors},
		), end
	}
	return KeywordParameter{}, start
}

func parseProtectionKeywordParameter(
	tokens []shared.Token,
	start int,
	atoms Atoms,
) (parameter KeywordParameter, end int) {
	if start+1 >= len(tokens) || !equalWord(tokens[start], "from") {
		return KeywordParameter{}, start
	}
	if equalWord(tokens[start+1], "everything") {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+2]),
			"everything",
			ProtectionParameter{Everything: true},
		), start + 2
	}
	if start+5 < len(tokens) &&
		(equalWord(tokens[start+1], "the") || equalWord(tokens[start+1], "a")) &&
		equalWord(tokens[start+2], "color") && equalWord(tokens[start+3], "of") &&
		equalWord(tokens[start+4], "your") && equalWord(tokens[start+5], "choice") {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+6]),
			"color of your choice",
			ProtectionParameter{ChosenColor: true},
		), start + 6
	}
	if start+2 < len(tokens) && equalWord(tokens[start+1], "the") &&
		equalWord(tokens[start+2], "chosen") && start+3 < len(tokens) &&
		equalWord(tokens[start+3], "color") {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+4]),
			"the chosen color",
			ProtectionParameter{ChosenColor: true},
		), start + 4
	}
	if qualifier, ok := atoms.ColorQualifierAt(tokens[start+1].Span); ok {
		switch qualifier {
		case ColorQualifierMulticolored:
			return NewProtectionKeywordParameter(
				shared.SpanOf(tokens[start:start+2]),
				"multicolored",
				ProtectionParameter{Multicolored: true},
			), start + 2
		case ColorQualifierMonocolored:
			return NewProtectionKeywordParameter(
				shared.SpanOf(tokens[start:start+2]),
				"monocolored",
				ProtectionParameter{Monocolored: true},
			), start + 2
		default:
		}
	}
	if start+9 < len(tokens) &&
		equalWord(tokens[start+1], "each") && equalWord(tokens[start+2], "color") &&
		equalWord(tokens[start+3], "that's") && equalWord(tokens[start+4], "not") &&
		equalWord(tokens[start+5], "in") && equalWord(tokens[start+6], "your") &&
		equalWord(tokens[start+7], "commander's") && equalWord(tokens[start+8], "color") &&
		equalWord(tokens[start+9], "identity") {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+10]),
			"commander identity complement",
			ProtectionParameter{CommanderIdentityComplement: true},
		), start + 10
	}
	if start+2 < len(tokens) &&
		(equalWord(tokens[start+1], "each") && equalWord(tokens[start+2], "color") ||
			equalWord(tokens[start+1], "all") &&
				(equalWord(tokens[start+2], "color") || equalWord(tokens[start+2], "colors"))) {
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:start+3]),
			"eachcolor",
			ProtectionParameter{EachColor: true},
		), start + 3
	}
	if colors, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (Color, bool) {
		return atoms.ColorAt(token.Span)
	}); ok {
		names := make([]string, len(colors))
		for i, color := range colors {
			names[i] = colorName(color)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			strings.Join(names, ","),
			ProtectionParameter{FromColors: colors},
		), end
	}
	if cardTypes, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (CardType, bool) {
		cardType, found := atoms.CardTypeAt(token.Span)
		return cardType, found && protectionCardType(cardType)
	}); ok {
		names := make([]string, len(cardTypes))
		for i, cardType := range cardTypes {
			names[i] = cardTypeName(cardType)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			"types:"+strings.Join(names, ","),
			ProtectionParameter{FromTypes: cardTypes},
		), end
	}
	if subtypes, end, ok := parseProtectionList(tokens, start, func(token shared.Token) (types.Sub, bool) {
		subtype, found := atoms.SubtypeAt(token.Span)
		return subtype, found && SubtypeMatchesAnyRuntimeCardType(subtype, []types.Card{types.Creature, types.Land})
	}); ok {
		names := make([]string, len(subtypes))
		for i, subtype := range subtypes {
			names[i] = string(subtype)
		}
		return NewProtectionKeywordParameter(
			shared.SpanOf(tokens[start:end]),
			"subtypes:"+strings.Join(names, ","),
			ProtectionParameter{FromSubtypes: subtypes},
		), end
	}
	return KeywordParameter{}, start
}

func parseProtectionList[T any](
	tokens []shared.Token,
	start int,
	parse func(shared.Token) (T, bool),
) (values []T, end int, ok bool) {
	first, ok := parse(tokens[start+1])
	if !ok {
		return nil, start, false
	}
	values = []T{first}
	end = start + 2
	for end < len(tokens) {
		next := end
		if tokens[next].Kind == shared.Comma {
			next++
		} else if !equalWord(tokens[next], "and") {
			break
		}
		if next < len(tokens) && equalWord(tokens[next], "and") {
			next++
		}
		if next >= len(tokens) || !equalWord(tokens[next], "from") {
			break
		}
		if next+1 >= len(tokens) {
			return nil, start, false
		}
		value, found := parse(tokens[next+1])
		if !found {
			return nil, start, false
		}
		values = append(values, value)
		end = next + 2
	}
	return values, end, true
}

func protectionCardType(cardType CardType) bool {
	switch cardType {
	case CardTypeArtifact, CardTypeCreature, CardTypeEnchantment, CardTypeInstant,
		CardTypeLand, CardTypePlaneswalker, CardTypeSorcery:
		return true
	default:
		return false
	}
}

func colorName(color Color) string {
	switch color {
	case ColorWhite:
		return "white"
	case ColorBlue:
		return "blue"
	case ColorBlack:
		return "black"
	case ColorRed:
		return "red"
	case ColorGreen:
		return "green"
	default:
		return ""
	}
}

func cardTypeName(cardType CardType) string {
	switch cardType {
	case CardTypeArtifact:
		return "artifact"
	case CardTypeCreature:
		return "creature"
	case CardTypeEnchantment:
		return "enchantment"
	case CardTypeInstant:
		return "instant"
	case CardTypeLand:
		return "land"
	case CardTypePlaneswalker:
		return "planeswalker"
	case CardTypeSorcery:
		return "sorcery"
	default:
		return ""
	}
}

func cloneProtectionParameter(protection ProtectionParameter) ProtectionParameter {
	protection.FromColors = slices.Clone(protection.FromColors)
	protection.FromTypes = slices.Clone(protection.FromTypes)
	protection.FromSubtypes = slices.Clone(protection.FromSubtypes)
	return protection
}

func scanKeywordSelectors(tokens []shared.Token) []KeywordSelector {
	var selectors []KeywordSelector
	for i := range tokens {
		excluded := false
		nameStart := 0
		form := KeywordSelectorFormDirect
		switch {
		case equalWord(tokens[i], "with"):
			nameStart = i + 1
			if nameStart < len(tokens) && equalWord(tokens[nameStart], "a") {
				nameStart++
				form = KeywordSelectorFormAbility
			}
		case equalWord(tokens[i], "without"):
			excluded = true
			nameStart = i + 1
		default:
			continue
		}
		kind, width, ok := recognizeKeywordNameAt(tokens, nameStart)
		if !ok {
			continue
		}
		end := nameStart + width
		if nameStart == i+2 {
			if end >= len(tokens) || !equalWord(tokens[end], "ability") {
				continue
			}
			end++
		}
		selectors = append(selectors, KeywordSelector{
			Keyword:  kind,
			Form:     form,
			Span:     shared.SpanOf(tokens[i:end]),
			Excluded: excluded,
		})
	}
	return selectors
}

// devourSubject describes one printed Devour variant: the optional permanent-type
// word that follows "Devour" (empty for the plain creature form), the plural and
// singular nouns its canonical as-enters wording uses, and the structured filter
// downstream stages apply when choosing what may be sacrificed. cardType is set
// for base card types (artifact, land); subtype is set for typed permanents named
// by subtype (Food). Both are zero for the creature form, which keeps the
// existing creature-only Devour lowering and rendering byte-for-byte unchanged.
type devourSubject struct {
	keyword  string
	plural   string
	singular string
	cardType types.Card
	subtype  types.Sub
}

// devourSubjects lists the Devour variants the parser expands (CR 702.81). The
// creature form is first and carries no structured filter; the typed forms name
// the permanents their controller may sacrifice as the creature enters.
var devourSubjects = []devourSubject{
	{plural: "creatures", singular: "creature"},
	{keyword: "artifact", plural: "artifacts", singular: "artifact", cardType: types.Artifact},
	{keyword: "land", plural: "lands", singular: "land", cardType: types.Land},
	{keyword: "Food", plural: "Foods", singular: "Food", subtype: types.Food},
}

// devourCanonicalText is the canonical as-enters replacement that a printed
// Devour keyword abbreviates (CR 702.81), naming the sacrificed permanents from
// subject and writing the per-sacrificed-permanent +1/+1 counter multiplier N as
// a plain integer. parseDevourEffect recognizes this exact wording and recovers
// both N and the subject.
func devourCanonicalText(subject devourSubject, n int) string {
	return "As this creature enters, you may sacrifice any number of " + subject.plural + ", " +
		"then it enters with " + strconv.Itoa(n) + " +1/+1 counters on it for each " + subject.singular + " sacrificed."
}

// expandDevourKeyword rewrites each printed Devour keyword line into the
// canonical as-enters replacement it abbreviates (CR 702.81). Like Bushido and
// Extort, Devour is shorthand for a fixed ability, so expanding it to canonical
// wording lets the standard replacement pipeline lower it. The creature form
// ("Devour N") and the typed permanent forms ("Devour artifact N", "Devour land
// N", "Devour Food N") are expanded; the variable form ("Devour X ...") is left
// untouched. The rewrite is parser-owned because it is a wording substitution;
// downstream stages see only the expanded ability.
func expandDevourKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		subject, n, ok := devourLineRank(line)
		if !ok {
			continue
		}
		lines[i] = devourCanonicalText(subject, n)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// devourLineRank reports the subject and rank N of a line that is exactly a
// printed Devour keyword, optionally followed only by its parenthesized reminder
// text. The creature form begins with the rank digits; the typed forms begin
// with the permanent-type word ("artifact"/"land"/"Food") and then the rank. The
// variable "Devour X ..." form, lines that merely contain the word elsewhere,
// and lines that pair it with other rules text are left untouched.
func devourLineRank(line string) (devourSubject, int, bool) {
	const prefix = "Devour "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return devourSubject{}, 0, false
	}
	subject, rest, ok := devourSubjectPrefix(strings.TrimSpace(trimmed[len(prefix):]))
	if !ok {
		return devourSubject{}, 0, false
	}
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return devourSubject{}, 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return devourSubject{}, 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return devourSubject{}, 0, false
	}
	return subject, rank, true
}

// devourSubjectPrefix splits the text after "Devour " into its subject and the
// remaining text that should begin with the rank digits. The creature form has
// no type word, so text that starts with a digit yields the creature subject
// unchanged; otherwise the leading word must name a typed Devour permanent.
func devourSubjectPrefix(rest string) (devourSubject, string, bool) {
	if rest != "" && rest[0] >= '0' && rest[0] <= '9' {
		return devourSubjects[0], rest, true
	}
	for _, subject := range devourSubjects[1:] {
		word := subject.keyword + " "
		if strings.HasPrefix(rest, word) {
			return subject, strings.TrimSpace(rest[len(word):]), true
		}
	}
	return devourSubject{}, "", false
}

// devourSubjectByNouns finds the Devour subject whose canonical plural and
// singular nouns match the given lower-cased words, recovering the structured
// sacrifice filter from the expanded wording.
func devourSubjectByNouns(plural, singular string) (devourSubject, bool) {
	for _, subject := range devourSubjects {
		if strings.ToLower(subject.plural) == plural && strings.ToLower(subject.singular) == singular {
			return subject, true
		}
	}
	return devourSubject{}, false
}

// tributeCanonicalText is the canonical as-enters replacement that the printed
// "Tribute N" keyword abbreviates (CR 702.110), with the +1/+1 counter count N
// written as a plain integer. parseTributeEffect recognizes this exact wording
// and recovers N.
func tributeCanonicalText(n int) string {
	return "As this creature enters, an opponent of your choice may put " +
		strconv.Itoa(n) + " +1/+1 counters on it."
}

// expandTributeKeyword rewrites each printed "Tribute N" keyword line into the
// canonical as-enters replacement it abbreviates (CR 702.110). Like Devour,
// Tribute is shorthand for a fixed ability, so expanding it to canonical wording
// lets the standard replacement pipeline lower it; the paired printed "When this
// creature enters, if tribute wasn't paid, ..." ability is left untouched. The
// rewrite is parser-owned because it is a wording substitution; downstream stages
// see only the expanded ability.
func expandTributeKeyword(source string) string {
	lines := strings.Split(source, "\n")
	changed := false
	for i, line := range lines {
		n, ok := tributeLineRank(line)
		if !ok {
			continue
		}
		lines[i] = tributeCanonicalText(n)
		changed = true
	}
	if !changed {
		return source
	}
	return strings.Join(lines, "\n")
}

// tributeLineRank reports the rank N of a line that is exactly the printed
// "Tribute N" keyword, optionally followed only by its parenthesized reminder
// text. The word immediately after "Tribute " must be the rank digits. Lines
// that merely contain the word elsewhere, or pair it with other rules text, are
// left untouched.
func tributeLineRank(line string) (int, bool) {
	const prefix = "Tribute "
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(trimmed[len(prefix):])
	digits := 0
	for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, false
	}
	rank, err := strconv.Atoi(rest[:digits])
	if err != nil || rank <= 0 {
		return 0, false
	}
	tail := strings.TrimSpace(rest[digits:])
	if tail != "" && (!strings.HasPrefix(tail, "(") || !strings.HasSuffix(tail, ")")) {
		return 0, false
	}
	return rank, true
}
