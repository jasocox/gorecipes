package allrecipes

import (
  "log"
  "net/http"
  "io/ioutil"
)

const HOSTNAME = "http://allrecipes.com/recipes/"

const RECIPE_VIEW_ALL = "ViewAll.aspx"

var recipeUrlList = []string{"pasta/"}

func NewReader() <-chan string {
  r := make(chan string)

  // Fan-in pattern
  for url := range recipeUrlList {
    go readPage(translateUrl(recipeUrlList[url]), r)
  }

  return r
}

func translateUrl(url string) string {
  return HOSTNAME + url + RECIPE_VIEW_ALL
}

func readPage(url string, r chan<- string) {
  resp, err := http.Get(url)
  defer resp.Body.Close()

  if err != nil {
    log.Println("Failed to read from " + url)
    return
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return
  }

  r <- string(body)
}

func filterRecipeLinks(body string) []string {
  return []string{body}
}
