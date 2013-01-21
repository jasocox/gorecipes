package allrecipes

import (
  "log"
  "net/http"
  "io/ioutil"
)

const HOSTNAME = "http://allrecipes.com"

func NewReader() <-chan string {
  r := make(chan string)

  go readAllRecipes(r)

  return r
}

func readAllRecipes(r chan<- string) {
  resp, err := http.Get(HOSTNAME)
  defer resp.Body.Close()

  if err != nil {
    log.Println("Failed to read from " + HOSTNAME)
    return
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return
  }

  r <- string(body)
}
