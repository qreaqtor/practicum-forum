package utils

import (
	"encoding/json"
	"net/http"
)

/*
Вызывается в случае поялвения ошибки, пишет msg в логи.
Статус ответа и сообщение достает из msg.
Возвращает 500 в случае неудачной записи в w.
*/
func WriteError(w http.ResponseWriter, msg *logMsg) {
	msg.Error()
	// w.Header().Set("Content-Type", "application/json")
	// response, err := json.Marshal(map[string]interface{}{
	// 	"msg": msg.Message,
	// })
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// _, err = w.Write(response)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	http.Error(w, msg.Message, msg.Status)
}

/*
Выполняет сериализацию data и пишет в w.
В случаае появления ошибки вызывает writeError().
*/
func WriteData(w http.ResponseWriter, msg *logMsg, data any) {
	response, err := json.Marshal(data)
	if err != nil {
		msg.Message = err.Error()
		msg.Status = http.StatusInternalServerError
		WriteError(w, msg)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		msg.Message = err.Error()
		msg.Status = http.StatusInternalServerError
		WriteError(w, msg)
		return
	}
	msg.Info()
}
