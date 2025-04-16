package operations

import (
	"encoding/json"
	"fmt"
)

type Operation struct {
	Name           string         `json:"name"`
	JavascriptCode string         `json:"javascriptCode"`
	Parameters     []*ValueSchema `json:"parameters"`
	Return         *ValueSchema
}

type ValueSchema struct {
	Name              string          `json:"name"`
	Type              Type            `json:"type"`
	TypeProperties    TypeProperties  `json:"-"`
	TypePropertiesRaw json.RawMessage `json:"typeProperties"`
}

type Type string

const (
	String  Type = "string"
	Number  Type = "number"
	Boolean Type = "boolean"
	Object  Type = "object"
	Array   Type = "array"
)

type TypeProperties interface {
	Validate(value any) ValidationResult
}

type ValidationResult struct {
	Success bool
	Message string
}

type StringProperties struct{}

func (p *StringProperties) Validate(value any) ValidationResult {
	_, isString := value.(string)
	if isString {
		return ValidationResult{Success: true}
	}

	return ValidationResult{Success: false, Message: "value is not a string"}
}

type NumberProperties struct{}

func (p *NumberProperties) Validate(value any) ValidationResult {
	_, isFloat := value.(float64)
	if !isFloat {
		return ValidationResult{Success: false, Message: "value is not a float64"}
	}
	return ValidationResult{Success: true}
}

type BooleanProperties struct{}

func (p *BooleanProperties) Validate(value any) ValidationResult {
	_, isBool := value.(bool)
	if !isBool {
		return ValidationResult{Success: false, Message: "value is not a bool"}
	}
	return ValidationResult{Success: true}
}

type ObjectProperties struct {
	Properties map[string]*ValueSchema
}

func (p *ObjectProperties) Validate(value any) ValidationResult {
	object, isMap := value.(map[string]any)
	if !isMap {
		return ValidationResult{Success: false, Message: "value is not a map"}
	}
	for _, objectProperty := range p.Properties {
		propertyValue, hasProperty := object[objectProperty.Name]
		if !hasProperty {
			return ValidationResult{Success: false, Message: fmt.Sprintf("missing property %s", objectProperty.Name)}
		}

		properyValidationResult := objectProperty.TypeProperties.Validate(propertyValue)
		if !properyValidationResult.Success {
			return ValidationResult{Success: false, Message: fmt.Sprintf("invalid propery %s: %s", objectProperty.Name, properyValidationResult.Message)}
		}
	}

	return ValidationResult{Success: true}
}

type ArrayProperties struct {
	Items *ValueSchema
}

func (p *ArrayProperties) Validate(value any) ValidationResult {
	array, isSlice := value.([]any)
	if !isSlice {
		return ValidationResult{Success: false, Message: "value is not a slice"}
	}
	for index, item := range array {
		itemValidationResult := p.Items.TypeProperties.Validate(item)
		if !itemValidationResult.Success {
			return ValidationResult{Success: false, Message: fmt.Sprintf("invalid item %d: %s", index, itemValidationResult.Message)}
		}
	}

	return ValidationResult{Success: true}
}

type _valueSchema ValueSchema

func (s *ValueSchema) MarshalJSON() ([]byte, error) {
	typePropertiesRaw, err := json.Marshal(s.TypeProperties)
	if err != nil {
		return nil, err
	}
	s.TypePropertiesRaw = typePropertiesRaw

	return json.Marshal((*_valueSchema)(s))
}

func (s *ValueSchema) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, (*_valueSchema)(s))
	if err != nil {
		return err
	}

	switch s.Type {
	case String:
		stringProperties := &StringProperties{}
		err := json.Unmarshal(s.TypePropertiesRaw, stringProperties)
		if err != nil {
			return err
		}
		s.TypeProperties = stringProperties
	case Number:
		numberProperties := &NumberProperties{}
		err := json.Unmarshal(s.TypePropertiesRaw, numberProperties)
		if err != nil {
			return err
		}
		s.TypeProperties = numberProperties
	case Boolean:
		booleanProperties := &BooleanProperties{}
		err := json.Unmarshal(s.TypePropertiesRaw, booleanProperties)
		if err != nil {
			return err
		}
		s.TypeProperties = booleanProperties
	case Object:
		objectProperties := &ObjectProperties{}
		err := json.Unmarshal(s.TypePropertiesRaw, objectProperties)
		if err != nil {
			return err
		}
		s.TypeProperties = objectProperties
	case Array:
		arrayProperties := &ArrayProperties{}
		err := json.Unmarshal(s.TypePropertiesRaw, arrayProperties)
		if err != nil {
			return err
		}
		s.TypeProperties = arrayProperties
	}

	return nil
}
