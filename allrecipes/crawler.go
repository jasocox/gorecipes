package allrecipes

import (
  "log"
  "strings"
  "regexp"
  "net/http"
  "io/ioutil"
)

const HOSTNAME = "http://allrecipes.com/recipes/"

const RECIPE_VIEW_ALL = "ViewAll.aspx"

var (
  recipeUrlList = []string{"pasta/"}
  matchRecipe *regexp.Regexp
  getRecipe *regexp.Regexp
)

func init() {
  matchRecipe = regexp.MustCompile("href=\"(.*recipe/.*/detail.aspx)\"")
  getRecipe = regexp.MustCompile("\"(.*recipe/.*/detail.aspx)\"")
}

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

  recipes := filterRecipeLinks(string(body))
  for recipe := range recipes {
    r <- string(strings.Trim(getRecipe.FindString(recipes[recipe]), "\""))
  }
}

func filterRecipeLinks(body string) []string {
  return matchRecipe.FindAllString(body, -1)
}
