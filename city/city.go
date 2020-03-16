package city

import (
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")
type cityInfo struct {
	Value string `json:"value"`
	Label string `json:"label"`
}








