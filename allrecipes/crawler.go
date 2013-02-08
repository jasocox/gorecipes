package allrecipes

import (
  "log"
  "regexp"
  "net/http"
  "io/ioutil"
  "gorecipes/recipe"
)

const RECIPE_VIEW_ALL = "http://allrecipes.com/recipes/ViewAll.aspx"

var (
  translatorMap map[string]Translator
)

func init() {
  translatorMap = make(map[string]Translator)

  addTranslator("RecipeLink", generateTranslator("<a[^>]*id=\"[^\"]*_lnkRecipeTitle\"[^>]*href=\"(.*recipe/.*/detail.aspx)\"", listFilter))
  addTranslator("Next", generateTranslator("<a href=\"([^<]*)\">NEXT »</a>", simpleFilter))
  addTranslator("Name", generateTranslator("<h1 id=\"itemTitle\"[^>]*>([^<>]*)</h1>", simpleFilter))
  addTranslator("ImageLink", generateTranslator("<img id=\"imgPhoto\"[^>]*src=\"([^\"]*)\"[^>]*>", simpleFilter))
  addTranslator("Rating", generateTranslator("<meta itemprop=\"ratingValue\" content=\"([^\"]*)\"[^>]*>", simpleFilter))
  addTranslator("ReadyTimeMins", generateTranslator("<span id=\"readyMinsSpan\"><em>([^<>]*)", simpleFilter))
  addTranslator("ReadyTimeHours", generateTranslator("<span id=\"readyMinsSpan\"><em>([^<>]*)<", simpleFilter))
  addTranslator("CookTimeMins", generateTranslator("<span id=\"cookMinsSpan\"><em>([^<>]*)<", simpleFilter))
  addTranslator("CookTimeHours", generateTranslator("<span id=\"cookHoursSpan\"><em>([^<>]*)<", simpleFilter))
  addTranslator("Directions", generateTranslator("<span class=\"plaincharacterwrap break\">([^<>]*)</span>", listFilter))
  addTranslator("AmountsAndIngredients",
    generateTranslator("(<span [^>]*class=\"ingredient-amount\">([^<>]*)</span>)?[^<>]*" +
                       "<span [^>]*class=\"ingredient-name\">([^<>]*)</span>", listTupleFilter))
}

func NewRecipeReader() <-chan *recipe.Recipe {
  reader := make(chan *recipe.Recipe)

  recipeChannel := make(chan *recipe.Recipe, 100)
  go func() {
    for {
      recipe := <-recipeChannel
      reader <- recipe
    }
  }()

  go findRecipeLinksFromUrlAndFollowNext(RECIPE_VIEW_ALL, addRecipeLinkReader(recipeChannel))

  return reader
}

type Translator struct {
  Name string
  Translator func(string) interface{}
}

func addTranslator(name string, translator func(string) interface{}) {
  translatorMap[name] = Translator{Name: name, Translator: translator}
}

func generateTranslator(matchRegexp string, filter func (string, *regexp.Regexp) interface{}) func(string) interface{} {
  matcher := regexp.MustCompile(matchRegexp)

  return func(body string) interface{} {
    return filter(body, matcher)
  }
}

func simpleFilter(body string, matcher *regexp.Regexp) interface{} {
  var retVal string
  match := matcher.FindStringSubmatch(body)

  if (match == nil) || (len(match) < 2) {
    retVal = ""
  } else {
    retVal = match[1]
  }

  return retVal
}

func listFilter(body string, matcher *regexp.Regexp) interface{} {
  var retVal []string
  match := matcher.FindAllStringSubmatch(body, -1)

  if match != nil {
    for i := range match {
      if match[i] != nil && (len(match[i]) > 1) {
        retVal = append(retVal, match[i][1])
      }
    }
  }

  return retVal
}

func listTupleFilter(body string, matcher *regexp.Regexp) interface{} {
  var retVal [][2]string
  match := matcher.FindAllStringSubmatch(body, -1)

  if match != nil {
    for i := range match {
      if match[i] != nil && (len(match[i]) > 2) && match[i][3] != "&nbsp;" {
        retVal = append(retVal, [2]string{match[i][2], match[i][3]})
      }
    }
  }

  return retVal
}

func translate(name string, body string) interface{} {
  translator := translatorMap[name]

  return translator.Translator(body)
}

func addRecipeLinkReader(recipeChannel chan<- *recipe.Recipe) chan<- string {
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

func findRecipeLinksFromUrlAndFollowNext(url string, recipeLinkChannel chan<- string) {
  body, err := readBodyFromUrl(url)
  if err != nil {
    log.Println(url + ": Failed to read page of recipe links")
    return
  }

  nextLink := translate("Next", body).(string)
  if nextLink != "" {
    log.Println(url + ": Found next link")
    go findLinksFromBody(url, body, recipeLinkChannel)
    go findRecipeLinksFromUrlAndFollowNext(nextLink, recipeLinkChannel)
  } else {
    log.Println(url + ": Didn't find a next link")
    go findLinksFromBody(url, body, recipeLinkChannel)
  }
}

func findRecipeLinksFromUrl(url string, recipeLinkChannel chan<- string) {
  body, err := readBodyFromUrl(url)
  if err != nil {
    return
  }

  findLinksFromBody(url, body, recipeLinkChannel)
}

func findLinksFromBody(url string, body string, recipeLinkChannel chan<- string) {
  log.Println(url + ": Starting")
  recipes := translate("RecipeLink", body).([]string)
  for recipe := range recipes {
    recipeLinkChannel <- recipes[recipe]
  }

  log.Println(url + ": Done")
}

func translateRecipeFromBody(body string, url string) (r recipe.Recipe) {
  r.Name = translate("Name", body).(string)
  r.Link = url
  r.ImageLink = translate("ImageLink", body).(string)
  r.Rating = translate("Rating", body).(string)
  r.ReadyTimeMins = translate("ReadyTimeMins", body).(string)
  r.ReadyTimeHours = translate("ReadyTimeHours", body).(string)
  r.CookTimeMins = translate("CookTimeMins", body).(string)
  r.CookTimeHours = translate("CookTimeHours", body).(string)
  r.AmountsAndIngredients = translate("AmountsAndIngredients", body).([][2]string)
  r.Directions = translate("Directions", body).([]string)

  return
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
