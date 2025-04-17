package operations

import (
	"encoding/json"
	"fmt"
)

type Operation struct {
	Name           string                  `json:"name"`
	JavascriptCode string                  `json:"javascriptCode"`
	Parameters     map[string]*ValueSchema `json:"parameters"`
	Return         *ValueSchema
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
	Success bool
	Message string
}

type StringSpec struct{}

func (p *StringSpec) Validate(value any) ValidationResult {
	_, isString := value.(string)
	if isString {
		return ValidationResult{Success: true}
	}

	return ValidationResult{Success: false, Message: "value is not a string"}
}

type NumberSpec struct{}

func (p *NumberSpec) Validate(value any) ValidationResult {
	_, isFloat := value.(float64)
	if !isFloat {
		return ValidationResult{Success: false, Message: "value is not a float64"}
	}
	return ValidationResult{Success: true}
}

type BooleanSpec struct{}

func (p *BooleanSpec) Validate(value any) ValidationResult {
	_, isBool := value.(bool)
	if !isBool {
		return ValidationResult{Success: false, Message: "value is not a bool"}
	}
	return ValidationResult{Success: true}
}

type ObjectSpec struct {
	Properties map[string]*ValueSchema
}

func (p *ObjectSpec) Validate(value any) ValidationResult {
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
			return ValidationResult{Success: false, Message: fmt.Sprintf("invalid propery %s: %s", propertyName, properyValidationResult.Message)}
		}
	}

	return ValidationResult{Success: true}
}

type ArraySpec struct {
	Items *ValueSchema
}

func (p *ArraySpec) Validate(value any) ValidationResult {
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
