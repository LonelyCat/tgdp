//
// Project: TGDP - Traffic Generator for Diameter Protocol
// Description: Simple tool for testing and debugging the Diameter protocol
//
// Author: Alexander Kefeli <alexander.kefeli@gmail.com>
//
// File: avpstore.go
// Description: Diameter pkg: AVP data store
//

package diameter

import (
	"iter"
	"os"
	"slices"
	"sync"

	"gopkg.in/yaml.v3"

	"tgdp/pkg/diameter/diwe"
)

// Consts
//

// Action constants for LoadFromFile/MakeFromYaml operations.
// These determine how AVPs from YAML are applied to the store.
const (
	// AvpAppend appends AVPs to existing values for that AVP code.
	AvpStoreAppend = -1
	// AvpReplace replaces AVPs at a specific index for that AVP code.
	AvpStoreReplace = -2
	// AvpDelete removes AVPs at a specific index (handled by Delete method).
	AvpStoreDelete = -3
	// AvpPurge removes all AVPs for a specific AVP code (handled by Purge method).
	AvpStorePurge = -4
)

// Types
//

// AvpData holds the encoded value and size of an AVP.
// This is the internal representation used after encoding/decoding.
type AvpData struct {
	// Value is the Go value (type depends on AVP data type).
	Value any
	// Size is the wire format size in bytes.
	Size uint32
}

// AvpStore is a thread-safe container for AVP values indexed by AVP code.
// It supports storing multiple AVPs with the same code (e.g., multiple
// Routing-Host AVPs) and provides concurrent access via RWMutex.
type AvpStore struct {
	mu sync.RWMutex
	// data maps AVP code to a slice of AVP pointers (supports multiple values per code).
	data map[uint32][]*Avp
	// env is a reference to the Diameter environment for dictionary and codec access.
	env *Diameter
}

// Constructors
//

// NewAvpStore creates a new empty AvpStore with the given Diameter environment.
func NewAvpStore(env *Diameter) AvpStore {
	return AvpStore{
		mu:   sync.RWMutex{},
		data: make(map[uint32][]*Avp),
		env:  env,
	}
}

// Methods
//

// Fetch retrieves all AVPs with the given code from the store.
// Returns nil if no AVPs exist for that code.
func (store *AvpStore) Fetch(id uint32) []*Avp {
	store.mu.Lock()
	defer store.mu.Unlock()

	if values, exists := store.data[id]; exists {
		return values
	}

	return nil
}

// Store replaces all AVPs for the given code with the provided slice.
// This is an atomic operation - all existing values are replaced.
func (store *AvpStore) Store(id uint32, values []*Avp) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.data[id] = values
}

// Append adds new AVPs to the existing slice for the given code.
// If no AVPs exist for the code, creates a new slice.
// Returns true on success.
func (store *AvpStore) Append(id uint32, data []*Avp) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.data[id]; !exists {
		store.data[id] = make([]*Avp, 0)
	}
	store.data[id] = append(store.data[id], data...)

	return true
}

// Replace replaces AVPs at a specific index for the given code.
// The index must be within bounds of the existing slice.
// Returns true if replacement was successful, false if index out of bounds.
func (store *AvpStore) Replace(id uint32, index int, data []*Avp) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	if avps, exists := store.data[id]; exists {
		if index >= 0 && index < len(avps) {
			store.data[id] = append(avps[:index], append(data, avps[index+1:]...)...)
			return true
		}
	}

	return false
}

// Delete removes the AVP at the given index for the specified code.
// Returns true if deletion was successful, false if index out of bounds.
func (store *AvpStore) Delete(id uint32, index int) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	if avps, exists := store.data[id]; exists {
		if index >= 0 && index < len(avps) {
			store.data[id] = slices.Delete(avps, index, index+1)
			return true
		}
	}

	return false
}

// Purge removes all AVPs for the given code from the store.
func (store *AvpStore) Purge(id uint32) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.data, id)
}

// Range iterates over all AVP code/value pairs in the store.
// The provided function is called for each AVP code. If the function
// returns false, iteration stops.
func (store *AvpStore) Range(fn func(id uint32, values []*Avp) bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for id, values := range store.data {
		if !fn(id, values) {
			break
		}
	}
}

// LoadFromFile reads AVP data from a YAML file and applies it to the store.
// The yamlFile path should contain YAML-formatted AVP definitions.
// The action specifies how to apply the data (AvpAppend or AvpReplace).
// The index is used for AvpReplace to specify which existing AVP to replace.
func (store *AvpStore) LoadFromFile(yamlFile string, action, index int) error {
	yamlText, err := os.ReadFile(yamlFile)
	if err != nil {
		return err
	}

	return store.MakeFromYaml(string(yamlText), action, index)
}

// MakeFromYaml parses YAML data and applies it to the store.
// Expected YAML format:
//
//	avp_name: value             # Scalar value
//	avp_name:                   # Sequence of values
//	  - value1
//	  - value2
//	avp_name:                   # Grouped AVPs (nested mapping)
//	  nested_avp: value
//
// The action determines how AVPs are applied:
//   - AvpAppend: Add to existing AVPs with same code
//   - AvpReplace: Replace AVPs at specified index
func (store *AvpStore) MakeFromYaml(yamlData string, action, index int) error {
	node := yaml.Node{}
	err := yaml.Unmarshal([]byte(yamlData), &node)
	if err != nil {
		return err
	}

	if node.Content == nil {
		return nil
	}

	// Iterate over each top-level key-value pair in the YAML
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		nodeName := node.Content[0].Content[i]
		nodeData := node.Content[0].Content[i+1]

		// Convert YAML node to AVP(s)
		avps, err := store.yamlNodeToAvps(nodeName, nodeData)
		if err != nil {
			return err
		}

		// Apply the AVPs based on action
		switch action {
		case AvpStoreAppend:
			store.Append(avps[0].Code(), avps)
		case AvpStoreReplace:
			store.Replace(avps[0].Code(), index, avps)
		default:
			return &diwe.ErrUnknownStoreAction{Action: action}
		}
	}

	return nil
}

// yamlNodeToAvp recursively converts a YAML node to one or more AVP instances.
// Handles three YAML node types:
//   - ScalarNode: Simple key-value pair (e.g., "Session-Id: abc123")
//   - SequenceNode: Array of values (e.g., "Auth-Application-Id: [1, 2, 3]")
//   - MappingNode: Nested grouped AVP (e.g., "Destination-Host: { ... }")
//
// Returns a slice of AVPs (can be multiple for sequences) or an error.
func (store *AvpStore) yamlNodeToAvps(nodeName, nodeData *yaml.Node) ([]*Avp, error) {
	avpName := nodeName.Value
	var avpData any
	if err := nodeData.Decode(&avpData); err != nil {
		return nil, err
	}
	if avpData == nil {
		return nil, &diwe.ErrInvalidYamlValue{Line: nodeData.Line, Value: nodeData.Content}
	}

	result := make([]*Avp, 0)
	switch nodeData.Kind {
	case yaml.SequenceNode:
		// Handle arrays: recursively process each element
		for _, node := range nodeData.Content {
			if avps, err := store.yamlNodeToAvps(nodeName, node); err != nil {
				return nil, err
			} else {
				result = append(result, avps...)
			}
		}
		return result, nil
	case yaml.MappingNode:
		avpData = make([]*Avp, 0)
		// Recursively process each nested key-value pair
		for i := 0; i < len(nodeData.Content); i += 2 {
			n := nodeData.Content[i]
			d := nodeData.Content[i+1]
			if data, err := store.yamlNodeToAvps(n, d); err != nil {
				return nil, err
			} else {
				avpData = append(avpData.([]*Avp), data...)
			}
		}
		fallthrough
	case yaml.ScalarNode:
		// Look up AVP definition from dictionary
		if avp, err := store.env.GetAvp(avpName); err != nil {
			return nil, err
		} else {
			// Convert YAML number types to appropriate Go types based on AVP data type
			convNumber(&avpData, avp)
			// Set the value (this performs encoding)
			if err = avp.SetValue(avpData); err != nil {
				return nil, err
			}
			result = append(result, avp)
		}
	default:
		return nil, &diwe.ErrInvalidYamlValue{Line: nodeData.Line, Value: nodeData.Content}
	}

	return result, nil
}

// Iterators
//

// Iter returns a sequence that yields each slice of AVPs in the store.
// The yielded values are the slices of AVPs for each AVP code.
func (store *AvpStore) Iter() iter.Seq[[]*Avp] {
	return func(yield func([]*Avp) bool) {
		store.Range(func(id uint32, values []*Avp) bool {
			return yield(values)
		})
	}
}

// Iter2 returns a sequence that yields (code, AVPs) pairs for each AVP code.
func (store *AvpStore) Iter2() iter.Seq2[uint32, []*Avp] {
	return func(yield func(uint32, []*Avp) bool) {
		store.Range(func(id uint32, values []*Avp) bool {
			return yield(id, values)
		})
	}
}

// Copy returns a deep copy of the AvpStore instance.
// The new store contains copies of all AVPs, preserving the same structure.
// The Diameter environment reference is shared with the original store.
func (store *AvpStore) Copy() (*AvpStore, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	copy := NewAvpStore(store.env)
	for id, avps := range store.data {
		avpCopy := make([]*Avp, len(avps))
		var err error
		for i, avp := range avps {
			avpCopy[i], err = avp.Copy()
			if err != nil {
				return nil, err
			}
		}
		copy.data[id] = avpCopy
	}

	return &copy, nil
}

// Debug
//

// Print outputs the AVP store to stdout in human-readable format.
// This is an alias for Dump with no indentation.
func (store *AvpStore) Print(shift ...int) {
	store.Dump(shift...)
}

// Dump outputs all AVPs in the store to stdout in human-readable format.
// Each AVP is dumped with its name, code, and value.
func (store *AvpStore) Dump(shift ...int) {
	store.Range(func(id uint32, values []*Avp) bool {
		for _, avp := range values {
			avp.Dump(shift...)
		}
		return true
	})
}

// Helpers
//

// convNumber converts YAML number types to the appropriate Go type
// based on the AVP's data type definition from the dictionary.
// YAML decodes all numbers as float64 or int, but Diameter AVP types
// require specific integer sizes (int32, int64, uint32, uint64) or floats.
func convNumber(n *any, avp *Avp) {
	switch (*n).(type) {
	case int:
		switch avp.Type() {
		case avp.Dict().AvpDataType().Integer32:
			*n = int32((*n).(int))
		case avp.Dict().AvpDataType().Integer64:
			*n = int64((*n).(int))
		case avp.Dict().AvpDataType().Unsigned32:
			*n = uint32((*n).(int))
		case avp.Dict().AvpDataType().Unsigned64:
			*n = uint64((*n).(int))
		}
	case float64:
		switch avp.Type() {
		case avp.Dict().AvpDataType().Float32:
			*n = float32((*n).(float64))
		}
	}
}
