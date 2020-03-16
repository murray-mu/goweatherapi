package weather

import (
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")
var urlWeather = "http://t.weather.sojson.com/api/weather/city/"
type weatherInfo struct {
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
}








