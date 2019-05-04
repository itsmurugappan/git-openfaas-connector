package main

import (
  "crypto/hmac"
  "crypto/sha1"
  "encoding/hex"
  "errors"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "strings"
  "os"
  "encoding/json"
  "github.com/hashicorp/go-retryablehttp"
  "fmt"
  "bytes"
  "time"
  "strconv"
  "net"
)

func signBody(secret, body []byte) []byte {
  computed := hmac.New(sha1.New, secret)
  computed.Write(body)
  return []byte(computed.Sum(nil))
}

func verifySignature(secret []byte, signature string, body []byte) bool {

  const signaturePrefix = "sha1="
  const signatureLength = 45 // len(SignaturePrefix) + len(hex(sha1))

  if len(signature) != signatureLength || !strings.HasPrefix(signature, signaturePrefix) {
    return false
  }

  actual := make([]byte, 20)
  hex.Decode(actual, []byte(signature[5:]))

  return hmac.Equal(signBody(secret, body), actual)
}

type HookContext struct {
  Signature string
  Event     string
  Id        string
  Payload   []byte
}

func ParseHook(secret []byte, req *http.Request) (*HookContext, error) {
  hc := HookContext{}

  if hc.Signature = req.Header.Get("x-hub-signature"); len(hc.Signature) == 0 {
    return nil, errors.New("No signature!")
  }

  if hc.Event = req.Header.Get("x-github-event"); len(hc.Event) == 0 {
    return nil, errors.New("No event!")
  }

  if hc.Id = req.Header.Get("x-github-delivery"); len(hc.Id) == 0 {
    return nil, errors.New("No event Id!")
  }

  body, err := ioutil.ReadAll(req.Body)

  if err != nil {
    return nil, err
  }

  if !verifySignature(secret, hc.Signature, body) {
    return nil, errors.New("Invalid signature")
  }

  hc.Payload = body

  return &hc, nil
}

func main() {
  log.Printf("start server")
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {

  secret, _ := os.LookupEnv("secret")

  hc, err := ParseHook([]byte(secret), r)

  w.Header().Set("Content-type", "application/json")

  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    log.Printf("Failed processing hook! ('%s')", err)
    io.WriteString(w, "{}")
    return
  }

  //log.Printf("Received %s", hc.Payload)

  var payload map[string]interface{}
  json.Unmarshal([]byte(hc.Payload), &payload)
  repoInterface := payload["repository"]
  repo, _ := repoInterface.(map[string]interface{})
  repoName := fmt.Sprintf("%v",repo["name"])

  fnName, _ := os.LookupEnv(repoName)

  api, api_env_available := os.LookupEnv("api")

  if !api_env_available {
      api = "http://gateway:8080/function/"
  }

  timeout, to_env_available := os.LookupEnv("timeout")

  if !to_env_available {
      timeout = "100s"
  }

  timeout_duration , _ := time.ParseDuration(timeout)

  log.Printf("timeout_int :  %s", timeout_duration)


  retry, retry_env_available := os.LookupEnv("retry")

  if !retry_env_available {
      retry = "3"
  }

  retry_int , _ := strconv.Atoi(retry)

  httpClient := retryablehttp.NewClient()
  httpClient.HTTPClient = &http.Client{
        Transport: &http.Transport{
            Dial: TimeoutDialer(timeout_duration, timeout_duration),
        },
    }
  httpClient.RetryWaitMin = timeout_duration / 7
  httpClient.RetryWaitMax = timeout_duration / 2
  httpClient.RetryMax = retry_int
  httpClient.Backoff = retryablehttp.LinearJitterBackoff

  req, rErr := retryablehttp.NewRequest("POST", api + fnName, bytes.NewReader(hc.Payload))
  if rErr != nil {
    log.Printf("error building request: ('%s')", rErr)
    w.WriteHeader(http.StatusBadRequest)
    io.WriteString(w, "{}")
    return
  }

  req.Header.Set("Content-Type", "application/json")

  _ , hErr := httpClient.Do(req)
  if hErr != nil {
    log.Printf("error calling webhook: ('%s')", hErr)
    w.WriteHeader(http.StatusInternalServerError)
    io.WriteString(w, "{}")
    return
  }

  w.WriteHeader(http.StatusOK)
  io.WriteString(w, "{}")
  return
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
    return func(netw, addr string) (net.Conn, error) {
        conn, err := net.DialTimeout(netw, addr, cTimeout)
        if err != nil {
            return nil, err
        }
        conn.SetDeadline(time.Now().Add(rwTimeout))
        return conn, nil
    }
}

