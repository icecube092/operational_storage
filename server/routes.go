package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// статусы ответа
var (
	defaultLife = time.Second * 5
	notFound    = "Not found"
	updated     = "Updated"
	created     = "Created"
	deleted     = "Deleted"
	good        = "Ok"
	noMethod    = "No method"
	notValid    = "Not valid"
)

type stringLife struct{
	value        string
	updated      time.Time
	lifeDuration time.Duration
}

type httpHandler struct{
	storage map[string]stringLife
	mux     sync.RWMutex
}

type Content struct{
	Key string
	Value string
	Duration int64
}

type Response struct{
	Status string
	Value string
	Error error
}

func (h *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request){
	contentChan := make(chan Content, 1)
	respChan := make(chan Response, 1)
	resp := Response{}
	defer request.Body.Close()
	if ok, err := isValid(request, contentChan); !ok{
		resp.Error = err
		resp.Status = notValid
		resp.Value = ""
		return
	}else{
		switch request.Method {
		case "PUT":
			putHandle(<-contentChan, respChan, h)
		case "GET":
			getHandle(<-contentChan, respChan, h)
		case "DELETE":
			deleteHandle(<-contentChan, respChan, h)
		default:
			<-contentChan
			resp = Response{
				Status: noMethod,
				Value:  "",
				Error:  nil,
			}
		}
		bytearray, _ := json.Marshal(<-respChan)
		response.WriteHeader(200)
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		response.Write(bytearray)
	}
}

// проверка валидности запроса
func isValid(request *http.Request, channel chan Content)(bool, error){
	content := Content{}
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&content)
	if err != nil{
		return false, err
	}
	contentType := []string{"application/json"}
	for index := range request.Header["Content-Type"]{
		if request.Header["Content-Type"][index] != contentType[index]{
			return false, nil
		}
	}
	channel<-content
	return true, nil
}

// добавление данных в хранилище
func putHandle(cont Content, respChan chan Response, h *httpHandler){
	lifeValue := stringLife{
		value:   cont.Value,
		updated: time.Now(),
	}
	life := time.Duration(cont.Duration)
	if life != defaultLife && life != 0{
		lifeValue.lifeDuration = time.Second * time.Duration(cont.Duration)
	}else{
		lifeValue.lifeDuration = defaultLife
	}
	response := Response{}
	if _, ok := h.storage[cont.Key]; ok{
		response.Status = updated
	}else{
		response.Status = created
	}
	response.Error = nil
	response.Value = ""

	h.mux.Lock()
	h.storage[cont.Key] = lifeValue
	h.mux.Unlock()

	respChan<-response
}

// возврат данных из хранилища пользователю
func getHandle(cont Content, respChan chan Response, h *httpHandler){
	response := Response{}
	h.mux.RLock()
	if Value, ok := h.storage[cont.Key]; ok{
		response = Response{
			Value:  Value.value,
			Status: good,
			Error:  nil,
		}
	}else{
		response = Response{
			Value:  "",
			Status: notFound,
			Error:  nil,
		}
	}
	h.mux.RUnlock()
	respChan<-response
}

// удаление данных из хранилища
func deleteHandle(cont Content, respChan chan Response, h *httpHandler){
	response := Response{}
	if _, ok := h.storage[cont.Key]; ok{
		response.Status = deleted
		h.mux.Lock()
		delete(h.storage, cont.Key)
		h.mux.Unlock()
	}else{
		response.Status = notFound
	}
	response.Value = ""
	response.Error = nil

	respChan<-response
}

// проверка истечения времени
func checking(handler *httpHandler){
	// bottleneck
	for {
		handler.mux.RLock()
		for key, lifeValue := range handler.storage {
			if time.Since(lifeValue.updated) >= lifeValue.lifeDuration{
				delete(handler.storage, key)
			}
		}
		handler.mux.RUnlock()
	}
}