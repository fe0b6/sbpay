package sbpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	REST_TEST_URL = "https://3dsec.sberbank.ru/payment/rest/"
	REST_WORK_URL = "https://securepayments.sberbank.ru/payment/rest/"
)

// Регистрируем новый заказ в системе оплаты
func Register(o InitObj) (ans AnswerObj, err error) {

	if o.PageView == "" {
		o.PageView = "DESKTOP"
	}

	q := url.Values{}
	q.Add("userName", o.UserName)
	q.Add("password", o.Password)
	q.Add("orderNumber", o.OrderNumber)
	q.Add("amount", strconv.FormatFloat(o.Amount*100, 'f', 0, 64))
	q.Add("returnUrl", o.ReturnUrl)
	q.Add("failUrl", o.FailUrl)
	q.Add("pageView", o.PageView)

	var sbUrl string
	if o.IsTesting {
		sbUrl = REST_TEST_URL
	} else {
		sbUrl = REST_WORK_URL
	}

	// Запрашиваем регистрацию
	resp, err := http.Get(sbUrl + "register.do?" + q.Encode())
	if resp != nil {
		defer resp.Body.Close()
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("[error]", err)
		return
	}
	// Если что то пошло не так
	if resp.StatusCode != 200 {
		log.Println("[error]", resp.StatusCode, string(content))
		err = errors.New(resp.Status)
		return
	}

	// Парсим ответ
	err = json.Unmarshal(content, &ans)
	if err != nil {
		log.Println("[error]", err, string(content))
		return
	}

	// Если что-то пошло не так
	if ans.OrderId == "" {
		err = errors.New("Ошибка получение данных")
		log.Println("[error]", err, string(content))
		return
	}

	return
}

// Проверяем данные пришедшие от сбербанка
func CheckCallbackData(r *http.Request, o InitObj) (err error) {
	// Создаем хэши для значений
	h := make(map[string]string)
	var sum string
	keys := []string{}

	// Парсим форму
	err = r.ParseForm()
	if err != nil {
		log.Println("[error]", err)
		return
	}

	// Собираем значения
	for k, v := range r.Form {
		if k == "checksum" {
			sum = v[0]
		} else {
			h[k] = v[0]
			keys = append(keys, k)
		}
	}

	// Сортируем
	sort.Strings(keys)

	// Собираем данные
	hstr := []string{}
	for _, k := range keys {
		hstr = append(hstr, k, h[k])
	}

	hst := strings.Join(hstr, ";") + ";"

	// Формируем подпись
	sig := hmac.New(sha256.New, []byte(o.CallbackToken))
	sig.Write([]byte(hst))
	fsig := strings.ToUpper(hex.EncodeToString(sig.Sum(nil)))

	// Проверяем
	if fsig != sum {
		err = errors.New("bad sig")
		log.Println("[error]", err, fsig, sum)
		return
	}

	// Если не оплачено
	if r.FormValue("operation") != "deposited" || r.FormValue("status") != "1" {
		err = errors.New("not pay")
		log.Println("[info]", r.FormValue("mdOrder"), r.FormValue("orderNumber"),
			r.FormValue("operation"), r.FormValue("status"))
		return
	}

	return
}
