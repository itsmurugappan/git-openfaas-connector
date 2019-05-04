package main

import (
  "net/http"
  "net/http/httptest"
  "os"
  "fmt"
  "testing"
  "crypto/hmac"
  "crypto/sha1"
  "encoding/hex"
  "strings"
  "io/ioutil"
  "time"
  "io"
)

const testSecret = "1234"

func checkError(err error, t *testing.T) {
  if err != nil {
    t.Errorf("An error occurred. %v", err)
  }
}

func TestSuccessRequest(t *testing.T) {

  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, client")
  }))
  defer ts.Close()

  os.Setenv("secret", "1234")

  os.Setenv("api", ts.URL)
  os.Setenv("fnName", "test")

  jsonFile, _ := os.Open("test_payload.json")

  byteValue, _ := ioutil.ReadAll(jsonFile)

  body := string(byteValue)

  // Fails on invalid request
  req, err := http.NewRequest("POST", "/", strings.NewReader(body))

  req.Header.Add("x-hub-signature", signature(body))
  req.Header.Add("x-github-event", "push event")
  req.Header.Add("x-github-delivery", "push id")

  checkError(err, t)

  rr := httptest.NewRecorder()

  http.HandlerFunc(handler).
    ServeHTTP(rr, req)

  if status := rr.Code; status != http.StatusOK {
      t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
    }
}

func TestTimeoutRequest(t *testing.T) {

  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, client")
    time.Sleep(12 * time.Second)
  }))
  defer ts.Close()

  os.Setenv("secret", "1234")

  os.Setenv("api", ts.URL)
  os.Setenv("fnName", "test")
  os.Setenv("retry", "1")
  os.Setenv("timeout", "5s")

  jsonFile, _ := os.Open("test_payload.json")

  byteValue, _ := ioutil.ReadAll(jsonFile)

  body := string(byteValue)

  // Fails on invalid request
  req, err := http.NewRequest("POST", "/", strings.NewReader(body))

  req.Header.Add("x-hub-signature", signature(body))
  req.Header.Add("x-github-event", "push event")
  req.Header.Add("x-github-delivery", "push id")

  checkError(err, t)

  rr := httptest.NewRecorder()

  http.HandlerFunc(handler).
    ServeHTTP(rr, req)

  if status := rr.Code; status != http.StatusInternalServerError {
      t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
    }
}

func TestErrorRequest(t *testing.T) {

  ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     w.WriteHeader(http.StatusInternalServerError)
     io.WriteString(w, "{}")
  }))
  defer ts.Close()

  os.Setenv("secret", "1234")

  os.Setenv("api", ts.URL)
  os.Setenv("fnName", "test")
  os.Setenv("retry", "1")
  os.Setenv("timeout", "5s")

  jsonFile, _ := os.Open("test_payload.json")

  byteValue, _ := ioutil.ReadAll(jsonFile)

  body := string(byteValue)

  // Fails on invalid request
  req, err := http.NewRequest("POST", "/", strings.NewReader(body))

  req.Header.Add("x-hub-signature", signature(body))
  req.Header.Add("x-github-event", "push event")
  req.Header.Add("x-github-delivery", "push id")

  checkError(err, t)

  rr := httptest.NewRecorder()

  http.HandlerFunc(handler).
    ServeHTTP(rr, req)

  if status := rr.Code; status != http.StatusInternalServerError {
      t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
    }
}

func signature(body string) string {
  dst := make([]byte, 40)
  computed := hmac.New(sha1.New, []byte(testSecret))
  computed.Write([]byte(body))
  hex.Encode(dst, computed.Sum(nil))
  return "sha1=" + string(dst)
}
