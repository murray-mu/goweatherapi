package city

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"os"
	"testapi/goweatherapi/handler"
	"testapi/goweatherapi/log"
)

func Routes(code string) chi.Router {
	r := chi.NewRouter()
	r.Get("/", makeHandler(code, getCityinfoHandler))
	return r
}

type handlerFunc func(code string, w http.ResponseWriter, r *http.Request)

func makeHandler(code string, handler handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(code, w, r)
	}
}

func getCityInfo(code string, w http.ResponseWriter, r *http.Request) []*cityInfo {
	inputFile, inputError := os.Open("./bin/test.json")
	if inputError != nil {
		return nil
	}
	defer inputFile.Close()
	var s string
	inputReader := bufio.NewReader(inputFile)
	for {
		inputString, readerError := inputReader.ReadString('\n')
		if readerError == io.EOF {
			break
		}
		s = s + inputString
	}
	fmt.Printf("all: \n%s", s)

	var cityInfoList []cityInfo

	if err := json.Unmarshal([]byte(s), &cityInfoList); err == nil {
		fmt.Println(cityInfoList)
		cityInfos := make ([]*cityInfo, len(cityInfoList))
		for i := 0; i < len(cityInfoList); i++ {
			jsonCity := cityInfo{Value:cityInfoList[i].Value, Label:cityInfoList[i].Label}
			cityInfos[i] = &jsonCity
			fmt.Println(cityInfoList[i].Label)
		}
		return cityInfos;
	}else{
		fmt.Println("转换失败")
		return nil
	}
	return nil
}

func getCityinfoHandler(code string, w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogEntry(r).WithField("context", "cityInfo")
	id := chi.URLParam(r, "id")
	cityInfo := getCityInfo(id, w, r)
	err := render.RenderList(w, r, newCityInfoListResponse(cityInfo))
	if err != nil {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
}

func newCityInfoListResponse(cityInfos []*cityInfo) []render.Renderer {
	list := []render.Renderer{}
	for _, a := range cityInfos {
		list = append(list, newCityInfoResponse(a))
	}
	return list
}

func newCityInfoResponse(city *cityInfo) *cityInfoResponse {
	return &cityInfoResponse{city}
}

type cityInfoResponse struct {
	*cityInfo
}

func (dr *cityInfoResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

