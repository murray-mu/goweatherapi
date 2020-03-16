package weather

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"io/ioutil"
	"net/http"
	"testapi/goweatherapi/handler"
	"testapi/goweatherapi/log"
)

func Routes(code string) chi.Router {
	r := chi.NewRouter()
	r.Get("/", makeHandler(code, getHandler))
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", makeHandler(code, listHandler))
	})
	return r
}
func getHandler(code string, w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogEntry(r).WithField("context", "weather")
	id := chi.URLParam(r, "id")
	fmt.Println("TEST", id)
	logger.Infof(id)
}
type handlerFunc func(code string, w http.ResponseWriter, r *http.Request)

func makeHandler(code string, handler handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(code, w, r)
	}
}

func getWeatherInfo(code string, w http.ResponseWriter, r *http.Request) []*weatherInfo {
	logger := log.GetLogEntry(r).WithField("context", "weather")
	url :=  urlWeather + code
	fmt.Println("URL:>", url)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, _ := client.Do(req)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	if resp.StatusCode != 200 {
		return nil
	}
	body, _ := ioutil.ReadAll(resp.Body)

	var resBody map[string]interface{}
	err := json.Unmarshal(body, &resBody)
	if err != nil {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return nil
	}
	time := resBody["time"].(string)
	fmt.Println("time:", time)

	city := resBody["cityInfo"].(map[string]interface{})["city"].(string)
	fmt.Println("city:", city)

	todayInfo := resBody["data"].(map[string]interface{})["forecast"].([] interface{})[0].(map[string]interface{})
	temperature := todayInfo["high"].(string) + "," + todayInfo["low"].(string)
	fmt.Println("temperature:", temperature)

	weather := todayInfo["type"].(string)
	fmt.Println("weather:", weather)

	wind := todayInfo["fx"].(string) +  todayInfo["fl"].(string)
	fmt.Println("wind:", wind)

	jsonCity := weatherInfo{Value1:"City",Value2:city}
	jsonTime := weatherInfo{Value1:"Updated time",Value2:time}
	jsonWeather := weatherInfo{Value1:"Weather",Value2:weather}
	jsonTemperature := weatherInfo{Value1:"Temperature",Value2:temperature}
	jsonWind := weatherInfo{Value1:"Wind",Value2:wind}
	articles := []*weatherInfo{ &jsonCity, &jsonTime ,&jsonWeather, &jsonTemperature, &jsonWind}
	return articles
}

func listHandler(code string, w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogEntry(r).WithField("context", "weather")
	id := chi.URLParam(r, "id")
	weatherInfo := getWeatherInfo(id, w, r)
	err := render.RenderList(w, r, newWeatherInfoListResponse(weatherInfo))
	if err != nil {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
}

func newWeatherInfoListResponse(weatherInfos []*weatherInfo) []render.Renderer {
	list := []render.Renderer{}
	for _, a := range weatherInfos {
		list = append(list, newWeatherInfoResponse(a))
	}
	return list
}

func newWeatherInfoResponse(weather *weatherInfo) *weatherInfoResponse {
	return &weatherInfoResponse{weather}
}

type weatherInfoResponse struct {
	*weatherInfo
}

func (dr *weatherInfoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

