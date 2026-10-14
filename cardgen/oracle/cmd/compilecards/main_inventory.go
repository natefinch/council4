package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/internal/preservation"
)

func writeInventory(path, input string, results []result) error {
	pin, err := preservation.Read[preservation.Pin]("testdata/card-preservation/corpus.json")
	if err != nil {
		return err
	}
	if err := pin.Verify(input); err != nil {
		return err
	}
	manifest := preservation.Manifest{SchemaVersion: 1, Input: pin}
	for _, compiled := range results {
		if compiled.exclusion != "" {
			continue
		}
		card := compiled.card
		record := preservation.Record{
			Key: card.OracleID, Name: card.Name, Layout: card.Layout,
		}
		if card.OracleID != "" {
			record.OracleIDs = []string{card.OracleID}
		} else {
			record.Key = "printing:" + card.ID
		}
		if record.Key == "printing:" {
			return fmt.Errorf("preservation card %q has no stable identity", card.Name)
		}
		var text strings.Builder
		_, _ = text.WriteString(card.OracleText)
		for _, face := range card.CardFaces {
			record.Faces = append(record.Faces, face.Name)
			_ = text.WriteByte(0)
			_, _ = text.WriteString(face.OracleText)
		}
		if len(record.Faces) == 0 {
			record.Faces = []string{card.Name}
		}
		textHash := sha256.Sum256([]byte(text.String()))
		record.OracleTextSHA256 = hex.EncodeToString(textHash[:])
		for _, diagnostic := range compiled.diagnostics {
			record.Refusals = append(record.Refusals, diagnostic.Summary+": "+diagnostic.Detail)
		}
		if compiled.err != nil {
			record.Refusals = append(record.Refusals, compiled.err.Error())
		}
		if compiled.source == "" && len(record.Refusals) == 0 {
			return fmt.Errorf("accepted compiler card %q has no emitted definition", card.Name)
		}
		if len(record.Refusals) == 0 {
			sum := sha256.Sum256([]byte(compiled.source))
			record.DefinitionSHA256 = hex.EncodeToString(sum[:])
		}
		slices.Sort(record.Refusals)
		record.Refusals = slices.Compact(record.Refusals)
		manifest.Compiler = append(manifest.Compiler, record)
	}
	preservation.Sort(manifest.Compiler)
	manifest.Registry, err = preservation.RegistryInventory(manifest.Compiler)
	if err != nil {
		return err
	}
	if err := manifest.Validate(); err != nil {
		return err
	}
	return preservation.Write(path, manifest)
}
