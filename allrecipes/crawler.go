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
  matchName *regexp.Regexp
  getName *regexp.Regexp
  matchImageLink *regexp.Regexp
  getImageLink *regexp.Regexp
  matchRating *regexp.Regexp
  getRating *regexp.Regexp

  translatorMap map[string]Translator
)

func init() {
  translatorMap = make(map[string]Translator)

  recipeUrlMatchString := "\"(.*recipe/.*/detail.aspx)\""
  matchRecipe = regexp.MustCompile("href=" + recipeUrlMatchString)
  getRecipe = regexp.MustCompile(recipeUrlMatchString)

  nextUrlMatchString := "\"[^<]*\""
  addTranslator("Next", generateTranslatorsFilter("<a href=" + nextUrlMatchString + ">NEXT »</a>",
                                                  nextUrlMatchString, 0,
                                                  func(body string) string {
                                                    return strings.Trim(body, "\"")
                                                  }))

  nameMatchString := ">[^<>]*<"
  matchName = regexp.MustCompile("<h1 id=\"itemTitle\"[^>]*" + nameMatchString + "/h1>")
  getName = regexp.MustCompile(nameMatchString)

  imageLinkMatchString := "src=\"[^\"]*\""
  matchImageLink = regexp.MustCompile("<img id=\"imgPhoto\"[^>]*" + imageLinkMatchString + "[^>]*>")
  getImageLink = regexp.MustCompile(imageLinkMatchString)

  ratingMatchString := "content=\"[^\"]*\""
  matchRating = regexp.MustCompile("<meta itemprop=\"ratingValue\" " + ratingMatchString + "[^>]*>")
  getRating = regexp.MustCompile(ratingMatchString)
}

type Translator struct {
  Name string
  Translator func(string) string
}

func addTranslator(name string, translator func(string) string) {
  translatorMap[name] = Translator{Name: name, Translator: translator}
}

func generateTranslatorsFilter(matchRegexp string, getRegexp string, cutIndex int,
                       trimmer func(string) string) func(string) string {
  matcher := regexp.MustCompile(matchRegexp)
  getter := regexp.MustCompile(getRegexp)

  return func(body string) string {
    return trimmer(getter.FindString(matcher.FindString(body))[cutIndex:])
  }
}

func translate(name string, body string) string {
  translator := translatorMap[name]

  return translator.Translator(body)
}

func translateNameFromBody(body string) string {
  return strings.Trim(getName.FindString(matchName.FindString(body)), "<>")
}

func translateImageLinkFromBody(body string) string {
  return strings.Trim(getImageLink.FindString(matchImageLink.FindString(body))[4:], "\"")
}

func translateRatingFromBody(body string) string {
  return strings.Trim(getRating.FindString(matchRating.FindString(body))[8:], "\"")
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

func filterRecipeLinks(body string) []string {
  return matchRecipe.FindAllString(body, -1)
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

  nextLink := translate("Next", body)
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
  r.ImageLink = translateImageLinkFromBody(body)
  r.Rating = translateRatingFromBody(body)
  r.ReviewsLink = translateReviewsLinkFromBody(body)
  r.ReadyTime = translateReadyTimeFromBody(body)
  r.CookTime = translateCookTimeFromBody(body)
  r.Ingredients = translateIngredientsFromBody(body)
  r.Directions = translateDirectionsFromBody(body)

  return
}

func translateReviewsLinkFromBody(body string) string {
  return "ReviewsLink"
}

func translateReadyTimeFromBody(body string) string {
  return "ReadyTime"
}

func translateCookTimeFromBody(body string) string {
  return "CookTime"
}

func translateIngredientsFromBody(body string) []string {
  return []string{"Ingredients"}
}

func translateDirectionsFromBody(body string) []string {
  return []string{"Directions"}
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
