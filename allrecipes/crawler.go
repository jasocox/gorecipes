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
  matchNext *regexp.Regexp
  getNext *regexp.Regexp
)

func init() {
  recipeUrlMatchString := "\"(.*recipe/.*/detail.aspx)\""
  matchRecipe = regexp.MustCompile("href=" + recipeUrlMatchString)
  getRecipe = regexp.MustCompile(recipeUrlMatchString)

  nextUrlMatchString := "\"[^<]*\""
  matchNext = regexp.MustCompile("<a href=" + nextUrlMatchString + ">NEXT »</a>")
  getNext = regexp.MustCompile(nextUrlMatchString)
}

func NewReader() <-chan string {
  reader := make(chan string)

  recipeReader := make(chan string, 1000)
  go func() {
    recipeHash := make(map[string]string)
    for {
      recipe := <-recipeReader

      // Ignore any that have already been found
      if recipeHash[recipe] == "" {
        reader <- recipe
        recipeHash[recipe] = recipe
      }
    }
  }()

  for url := range recipeUrlList {
    addRecipeReaderThatFollowsNext(recipeUrlFromCategory(recipeUrlList[url]), recipeReader)
  }

  return reader
}

func recipeUrlFromCategory(url string) string {
  return HOSTNAME + url + RECIPE_VIEW_ALL
}

func extractRecipeLink(href string) string {
  return string(strings.Trim(getRecipe.FindString(href), "\""))
}

func extractNextLink(href string) string {
  return string(strings.Trim(getNext.FindString(href), "\""))
}

func filterRecipeLinks(body string) []string {
  return matchRecipe.FindAllString(body, -1)
}

func filterNextLink(body string) string {
  return matchNext.FindString(body)
}

func addRecipeReader(recipeUrl string, recipeReader chan<- string) {
  go readLinksFromUrl(recipeUrl, recipeReader)
}

func addRecipeReaderThatFollowsNext(recipeUrl string, recipeReader chan<- string) {
  go readLinksFromUrlAndFollowNext(recipeUrl, recipeReader)
}

func readLinksFromUrlAndFollowNext(url string, r chan<- string) {
  body, err := readBodyFromUrl(url)
  if err != nil {
    return
  }

  nextLink := extractNextLink(filterNextLink(body))
  if nextLink != "" {
    log.Println(url + ": Found next link")
    go readLinksFromBody(url, body, r)
    go readLinksFromUrlAndFollowNext(nextLink, r)
  } else {
    log.Println(url + ": Didn't find a next link")
    go readLinksFromBody(url, body, r)
  }
}

func readLinksFromUrl(url string, r chan<- string) {
  body, err := readBodyFromUrl(url)
  if err != nil {
    return
  }

  readLinksFromBody(url, body, r)
}

func readLinksFromBody(url string, body string, r chan<- string) {
  log.Println(url + ": Starting")
  recipes := filterRecipeLinks(body)
  for recipe := range recipes {
    r <- extractRecipeLink(recipes[recipe])
  }

  log.Println(url + ": Done")
}

func readBodyFromUrl(url string) (string, error) {
  resp, err := http.Get(url)
  defer resp.Body.Close()

  if err != nil {
    log.Println("Failed to process " + url)
    return "", err
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Println("Failed to read the body for " + url)
    return "", err
  }

  return string(body), nil
}
