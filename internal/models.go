package internal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Item struct {
	gorm.Model
	NameKey    string `json:"nameKey" gorm:"unique_index"`
	Name       string `json:"name"`
	Attributes JSON   `json:"attributes"`
}

type JSON json.RawMessage

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j JSON) GetMap() (map[string]interface{}, error) {
	b, err := json.RawMessage(j).MarshalJSON()
	if err != nil {
		return nil, err
	}

	ret := map[string]interface{}{}
	if err := json.Unmarshal(b, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func MapToJson(m map[string]interface{}) (JSON, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	ret := JSON{}
	if err := ret.Scan(b); err != nil {
		return nil, err
	}

	return ret, nil
}
