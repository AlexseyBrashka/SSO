package models

import "encoding/json"

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (t Tokens) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t Tokens) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &t)
}
