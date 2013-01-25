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
  recipeUrlList = []string{
    "pasta/",
    "appetizers-and-snacks/",
    "bread/",
    "breakfast-and-brunch/",
    "desserts/",
    "drinks/",
    "main-dish/",
  }

  matchRecipe *regexp.Regexp
  getRecipe *regexp.Regexp
)

func init() {
  recipeUrlMatchString := "\"(.*recipe/.*/detail.aspx)\""
  matchRecipe = regexp.MustCompile("href=" + recipeUrlMatchString)
  getRecipe = regexp.MustCompile(recipeUrlMatchString)
}

func NewReader() <-chan string {
  reader := make(chan string)

  recipeReader := make(chan string)
  go func() {
    for {
      recipe := <-recipeReader
      reader <- recipe
    }
  }()

  for url := range recipeUrlList {
    addRecipeReader(recipeUrlFromCategory(recipeUrlList[url]), recipeReader)
  }

  return reader
}

func recipeUrlFromCategory(url string) string {
  return HOSTNAME + url + RECIPE_VIEW_ALL
}

func extractRecipeLink(href string) string {
  return string(strings.Trim(getRecipe.FindString(href), "\""))
}

func addRecipeReader(recipeUrl string, recipeReader chan<- string) {
  go readLinksFromUrl(recipeUrl, recipeReader)
}

func readLinksFromUrl(url string, r chan<- string) {
  log.Println(url + ": Starting")
  resp, err := http.Get(url)
  defer resp.Body.Close()

  if err != nil {
    log.Println("Failed to process " + url)
    return
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Println("Failed to read the body for " + url)
    return
  }

  recipes := filterRecipeLinks(string(body))
  for recipe := range recipes {
    log.Println(url + ": Read a recipe")
    r <- extractRecipeLink(recipes[recipe])
  }

  log.Println(url + ": Done")
}

func filterRecipeLinks(body string) []string {
  return matchRecipe.FindAllString(body, -1)
}
