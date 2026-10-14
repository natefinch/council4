package preservation

import (
	"errors"
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/tokens"
	"github.com/natefinch/council4/mtg/game"
)

// RegistryInventory constructs and validates actual registered definitions.
// Legacy registrations lack Oracle IDs, so name/face matches remain candidates,
// not invented exact identities. Deck entries and token definitions stay separate.
func RegistryInventory(compiler []Record) ([]Record, error) {
	identities := make(map[string][]string)
	for _, record := range compiler {
		for _, name := range append([]string{record.Name}, record.Faces...) {
			identities[name] = append(identities[name], record.OracleIDs...)
		}
	}
	registry := cards.NewDefaultRegistry()
	names := registry.All()
	slices.Sort(names)
	var result []Record
	for _, name := range names {
		for index, def := range registry.LookupAll(name) {
			record, err := registryRecord(fmt.Sprintf("deck:%s#%d", name, index), def, identities[name])
			if err != nil {
				return nil, err
			}
			result = append(result, record)
		}
	}
	occurrences := make(map[string]int)
	for _, def := range tokens.Cards {
		if def == nil {
			return nil, errors.New("token registry contains a nil definition")
		}
		index := occurrences[def.Name]
		occurrences[def.Name]++
		record, err := registryRecord(fmt.Sprintf("token:%s#%d", def.Name, index), def, identities[def.Name])
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	Sort(result)
	return result, nil
}

func registryRecord(key string, def *game.CardDef, oracleIDs []string) (Record, error) {
	if def == nil {
		return Record{}, fmt.Errorf("registry entry %s returned nil", key)
	}
	record := Record{Key: key, Name: def.Name, Layout: string(def.Layout), OracleIDs: slices.Clone(oracleIDs)}
	slices.Sort(record.OracleIDs)
	record.OracleIDs = slices.Compact(record.OracleIDs)
	for _, face := range []*game.CardFace{&def.CardFace, optionalFace(def.Back.Exists, &def.Back.Val), optionalFace(def.Alternate.Exists, &def.Alternate.Val)} {
		if face == nil {
			continue
		}
		record.Faces = append(record.Faces, face.Name)
		if face.ImplementationID != "" {
			record.ImplementationIDs = append(record.ImplementationIDs, face.ImplementationID)
			record.Refusals = append(record.Refusals, "custom implementation requires an installed handler: "+face.ImplementationID)
		}
	}
	for _, issue := range game.ValidateCardDef(def) {
		record.Refusals = append(record.Refusals, fmt.Sprintf("%s %s: %s", issue.Code, issue.Path, issue.Message))
	}
	if len(record.Refusals) != 0 {
		return record, nil
	}
	fingerprint, err := definitionFingerprint(def)
	if err != nil {
		return Record{}, fmt.Errorf("fingerprinting registered %s: %w", key, err)
	}
	record.DefinitionSHA256 = fingerprint
	return record, nil
}

func optionalFace(exists bool, face *game.CardFace) *game.CardFace {
	if !exists {
		return nil
	}
	return face
}
