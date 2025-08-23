package govte

import (
	"fmt"
	"strings"
)

// MaxParams is the maximum number of parameters and subparameters
const MaxParams = 32

// Params holds the parameters and subparameters for escape sequences
type Params struct {
	// subparams stores the number of subparameters for each parameter
	subparams [MaxParams]uint8

	// params stores all parameters and subparameters
	params [MaxParams]uint16

	// currentSubparams tracks the number of subparameters in the current parameter
	currentSubparams uint8

	// len is the total number of parameters and subparameters
	len int
}

// NewParams creates a new Params instance
func NewParams() *Params {
	return &Params{}
}

// Len returns the total number of parameters and subparameters
func (p *Params) Len() int {
	return p.len
}

// IsEmpty returns true if there are no parameters
func (p *Params) IsEmpty() bool {
	return p.len == 0
}

// IsFull returns true if the params buffer is full
func (p *Params) IsFull() bool {
	return p.len >= MaxParams
}

// Clear removes all parameters
func (p *Params) Clear() {
	p.currentSubparams = 0
	p.len = 0
	// Clear arrays
	for i := range p.subparams {
		p.subparams[i] = 0
	}
	for i := range p.params {
		p.params[i] = 0
	}
}

// Push adds a new parameter (starts a new parameter group)
func (p *Params) Push(value uint16) {
	if p.IsFull() {
		return
	}

	// Store the parameter
	p.params[p.len] = value
	p.subparams[p.len] = 1 // This parameter group starts with 1 element (the main param)
	p.currentSubparams = 0
	p.len++
}

// Extend adds a subparameter to the current parameter group
func (p *Params) Extend(value uint16) {
	if p.IsFull() {
		return
	}

	if p.len == 0 {
		// No parameter to extend, treat as Push
		p.Push(value)
		return
	}

	// Find the start of the current parameter group
	// We need to find the last non-zero entry in subparams
	groupStart := p.len - 1
	for groupStart >= 0 && p.subparams[groupStart] == 0 {
		groupStart--
	}

	if groupStart < 0 {
		// No valid group found, treat as Push
		p.Push(value)
		return
	}

	// Store the subparameter
	p.params[p.len] = value
	p.subparams[p.len] = 0 // Mark this as a subparameter (not a group start)

	// Increment the count for the parameter group
	p.subparams[groupStart]++
	p.currentSubparams++
	p.len++
}

// Iter returns an iterator over parameters and their subparameters
func (p *Params) Iter() [][]uint16 {
	if p.len == 0 {
		return nil
	}

	var result [][]uint16
	i := 0

	for i < p.len {
		count := int(p.subparams[i])
		if count == 0 {
			// This is a subparameter position, skip (shouldn't happen if Push/Extend are used correctly)
			i++
			continue
		}

		// Collect this parameter group
		group := make([]uint16, 0, count)
		for j := 0; j < count && i+j < p.len; j++ {
			group = append(group, p.params[i+j])
		}

		result = append(result, group)
		i += count
	}

	return result
}

// String returns a string representation of the parameters
func (p *Params) String() string {
	iter := p.Iter()
	if len(iter) == 0 {
		return "Params{}"
	}

	var parts []string
	for _, group := range iter {
		if len(group) == 1 {
			parts = append(parts, fmt.Sprintf("%d", group[0]))
		} else {
			var subparts []string
			for _, v := range group {
				subparts = append(subparts, fmt.Sprintf("%d", v))
			}
			parts = append(parts, strings.Join(subparts, ":"))
		}
	}

	return fmt.Sprintf("Params{%s}", strings.Join(parts, ";"))
}
