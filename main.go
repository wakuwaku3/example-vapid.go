package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/rs/cors"
)

var subscription = &webpush.Subscription{}

func main() {
	opt := &webpush.Options{
		Subscriber:      "example@example.com",
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		TTL:             30,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/notifications/subscribe", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if strings.ToUpper(r.Method) == "POST" {
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, r.Body); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := json.Unmarshal(buf.Bytes(), subscription); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.ToUpper(r.Method) == "GET" {
			m := map[string]interface{}{
				"body":  "test-body",
				"title": "test-title",
			}
			b, err := json.Marshal(m)
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp, err := webpush.SendNotification(b, subscription, opt)
			if err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, resp.Body); err != nil {
				log.Print(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	handler := cors.Default().Handler(mux)
	log.Print(http.ListenAndServe(":8080", handler))
}
