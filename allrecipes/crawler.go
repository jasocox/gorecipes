package allrecipes

import (
  "log"
  "strings"
  "regexp"
  "net/http"
  "io/ioutil"
  "gorecipes/recipe"
)

const HOSTNAME = "http://allrecipes.com/recipes/"

const RECIPE_VIEW_ALL = "ViewAll.aspx"

var (
  recipeUrlList = []string{
    "pasta/",
    "drinks/",
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

func NewReader() <-chan *recipe.Recipe {
  reader := make(chan *recipe.Recipe)

  recipeChannel := make(chan *recipe.Recipe, 100)
  go func() {
    for {
      recipe := <-recipeChannel
      reader <- recipe
    }
  }()

  for url := range recipeUrlList {
    addRecipeFinderThatFollowsNext(recipeUrlFromCategory(recipeUrlList[url]), addRecipeReader(recipeChannel))
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

func addRecipeFinder(recipeUrl string, recipeLinkChannel chan<- string) {
  go findLinksFromUrl(recipeUrl, recipeLinkChannel)
}

func addRecipeFinderThatFollowsNext(recipeUrl string, recipeLinkChannel chan<- string) {
  go findLinksFromUrlAndFollowNext(recipeUrl, recipeLinkChannel)
}

func addRecipeReader(recipeChannel chan<- *recipe.Recipe) chan<- string {
  recipeFinderChannel := make(chan string)

  go func() {
    recipeLinkHash := make(map[string]string)
    for {
      recipeLink := <-recipeFinderChannel

      if recipeLinkHash[recipeLink] == "" {
        recipeChannel <- readRecipeLink(recipeLink)
        recipeLinkHash[recipeLink] = recipeLink
      }
    }
  }()

  return recipeFinderChannel
}

func readRecipeLink(recipeUrl string) *recipe.Recipe {
  log.Println(recipeUrl + ": Starting recipe")

  body, err := readBodyFromUrl(recipeUrl)
  if err != nil {
    log.Println(recipeUrl + ": Failed to read recipe link")
    return nil
  }

  r := translateRecipeFromBody(body, recipeUrl)
  log.Println(recipeUrl + ": Done with recipe")

  return &r
}

func findLinksFromUrlAndFollowNext(url string, recipeLinkChannel chan<- string) {
  body, err := readBodyFromUrl(url)
  if err != nil {
    log.Println(url + ": Failed to read page of recipe links")
    return
  }

  nextLink := extractNextLink(filterNextLink(body))
  if nextLink != "" {
    log.Println(url + ": Found next link")
    go findLinksFromBody(url, body, recipeLinkChannel)
    go findLinksFromUrlAndFollowNext(nextLink, recipeLinkChannel)
  } else {
    log.Println(url + ": Didn't find a next link")
    go findLinksFromBody(url, body, recipeLinkChannel)
  }
}

func findLinksFromUrl(url string, recipeLinkChannel chan<- string) {
  body, err := readBodyFromUrl(url)
  if err != nil {
    return
  }

  findLinksFromBody(url, body, recipeLinkChannel)
}

func findLinksFromBody(url string, body string, recipeLinkChannel chan<- string) {
  log.Println(url + ": Starting")
  recipes := filterRecipeLinks(body)
  for recipe := range recipes {
    recipeLinkChannel <- extractRecipeLink(recipes[recipe])
  }

  log.Println(url + ": Done")
}

func translateRecipeFromBody(body string, url string) (r recipe.Recipe) {
  r.Name = translateNameFromBody(body)
  r.Link = url

  return
}

func translateNameFromBody(body string) string {
  return "Recipe"
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
