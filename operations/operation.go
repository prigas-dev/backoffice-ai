package operations

import (
	"encoding/json"
	"fmt"
)

type Operation struct {
	Name           string                  `json:"name"`
	JavascriptCode string                  `json:"javascriptCode"`
	Parameters     map[string]*ValueSchema `json:"parameters"`
	Return         *ValueSchema            `json:"return"`
}

type ValueSchema struct {
	Type    Type            `json:"type"`
	Spec    Spec            `json:"-"`
	SpecRaw json.RawMessage `json:"spec"`
}

type Type string

const (
	String  Type = "string"
	Number  Type = "number"
	Boolean Type = "boolean"
	Object  Type = "object"
	Array   Type = "array"
)

type Spec interface {
	Validate(value any) ValidationResult
}

type ValidationResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type StringSpec struct {
	Nullable bool `json:"nullable"`
}

func (p *StringSpec) Validate(value any) ValidationResult {
	if p.Nullable && value == nil {
		return ValidationResult{Success: true}
	}
	_, isString := value.(string)
	if isString {
		return ValidationResult{Success: true}
	}

	return ValidationResult{Success: false, Message: "value is not a string"}
}

type NumberSpec struct {
	Nullable bool `json:"nullable"`
}

func (p *NumberSpec) Validate(value any) ValidationResult {
	if p.Nullable && value == nil {
		return ValidationResult{Success: true}
	}
	_, isFloat := value.(float64)
	if !isFloat {
		_, isInt := value.(int64)
		if !isInt {
			return ValidationResult{Success: false, Message: "value is not a float64 or int64"}
		}
	}
	return ValidationResult{Success: true}
}

type BooleanSpec struct {
	Nullable bool `json:"nullable"`
}

func (p *BooleanSpec) Validate(value any) ValidationResult {
	if p.Nullable && value == nil {
		return ValidationResult{Success: true}
	}
	_, isBool := value.(bool)
	if !isBool {
		return ValidationResult{Success: false, Message: "value is not a bool"}
	}
	return ValidationResult{Success: true}
}

type ObjectSpec struct {
	Nullable   bool                    `json:"nullable"`
	Properties map[string]*ValueSchema `json:"properties"`
}

func (p *ObjectSpec) Validate(value any) ValidationResult {
	if p.Nullable && value == nil {
		return ValidationResult{Success: true}
	}
	object, isMap := value.(map[string]any)
	if !isMap {
		return ValidationResult{Success: false, Message: "value is not a map"}
	}
	for propertyName, property := range p.Properties {
		propertyValue, hasProperty := object[propertyName]
		if !hasProperty {
			return ValidationResult{Success: false, Message: fmt.Sprintf("missing property %s", propertyName)}
		}

		properyValidationResult := property.Spec.Validate(propertyValue)
		if !properyValidationResult.Success {
			return ValidationResult{Success: false, Message: fmt.Sprintf("invalid property %s: %s", propertyName, properyValidationResult.Message)}
		}
	}

	return ValidationResult{Success: true}
}

type ArraySpec struct {
	Nullable bool         `json:"nullable"`
	Items    *ValueSchema `json:"items"`
}

func (p *ArraySpec) Validate(value any) ValidationResult {
	if p.Nullable && value == nil {
		return ValidationResult{Success: true}
	}
	array, isSlice := value.([]any)
	if !isSlice {
		return ValidationResult{Success: false, Message: "value is not a slice"}
	}
	for index, item := range array {
		itemValidationResult := p.Items.Spec.Validate(item)
		if !itemValidationResult.Success {
			return ValidationResult{Success: false, Message: fmt.Sprintf("invalid item %d: %s", index, itemValidationResult.Message)}
		}
	}

	return ValidationResult{Success: true}
}

type _valueSchema ValueSchema

func (s *ValueSchema) MarshalJSON() ([]byte, error) {
	typePropertiesRaw, err := json.Marshal(s.Spec)
	if err != nil {
		return nil, err
	}
	s.SpecRaw = typePropertiesRaw

	return json.Marshal((*_valueSchema)(s))
}

func (s *ValueSchema) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, (*_valueSchema)(s))
	if err != nil {
		return err
	}

	switch s.Type {
	case String:
		spec := &StringSpec{}
		err := json.Unmarshal(s.SpecRaw, spec)
		if err != nil {
			return err
		}
		s.Spec = spec
	case Number:
		spec := &NumberSpec{}
		err := json.Unmarshal(s.SpecRaw, spec)
		if err != nil {
			return err
		}
		s.Spec = spec
	case Boolean:
		spec := &BooleanSpec{}
		err := json.Unmarshal(s.SpecRaw, spec)
		if err != nil {
			return err
		}
		s.Spec = spec
	case Object:
		spec := &ObjectSpec{}
		err := json.Unmarshal(s.SpecRaw, spec)
		if err != nil {
			return err
		}
		s.Spec = spec
	case Array:
		spec := &ArraySpec{}
		err := json.Unmarshal(s.SpecRaw, spec)
		if err != nil {
			return err
		}
		s.Spec = spec
	}

	return nil
}
