package allrecipes

import (
  "log"
  "net/http"
  "io/ioutil"
  "gorecipes/recipe"
)

const RECIPE_VIEW_ALL = "http://allrecipes.com/recipes/ViewAll.aspx"

func init() {
  translatorConfig := [][]interface{}{
    []interface{}{"RecipeLink", "<a[^>]*id=\"[^\"]*_lnkRecipeTitle\"[^>]*href=\"(.*recipe/.*/detail.aspx)\"", listFilter},
    []interface{}{"Next", "<a href=\"([^<]*)\">NEXT »</a>", simpleFilter},
    []interface{}{"Name", "<h1 id=\"itemTitle\"[^>]*>([^<>]*)</h1>", simpleFilter},
    []interface{}{"ImageLink", "<img id=\"imgPhoto\"[^>]*src=\"([^\"]*)\"[^>]*>", simpleFilter},
    []interface{}{"Rating", "<meta itemprop=\"ratingValue\" content=\"([^\"]*)\"[^>]*>", simpleFilter},
    []interface{}{"ReadyTimeMins", "<span id=\"readyMinsSpan\"><em>([^<>]*)", simpleFilter},
    []interface{}{"ReadyTimeHours", "<span id=\"readyMinsSpan\"><em>([^<>]*)<", simpleFilter},
    []interface{}{"CookTimeMins", "<span id=\"cookMinsSpan\"><em>([^<>]*)<", simpleFilter},
    []interface{}{"CookTimeHours", "<span id=\"cookHoursSpan\"><em>([^<>]*)<", simpleFilter},
    []interface{}{"Directions", "<span class=\"plaincharacterwrap break\">([^<>]*)</span>", listFilter},
    []interface{}{"AmountsAndIngredients", "(<span [^>]*class=\"ingredient-amount\">([^<>]*)</span>)?[^<>]*" +
      "<span [^>]*class=\"ingredient-name\">([^<>]*)</span>", listTupleFilter},
  }

  generateTranslators(translatorConfig)
}

func NewRecipeReader() (<-chan *recipe.Recipe, <-chan string) {
  reader := make(chan *recipe.Recipe)
  messageBox := make(chan string)

  recipeChannel := make(chan *recipe.Recipe, 100)
  go func() {
    count := 0
    for {
      count++
      recipe := <-recipeChannel
      reader <- recipe
      if count >= 1 {
        messageBox <- "done"
        break
      }
    }
  }()

  go findRecipeLinksFromUrlAndFollowNext(RECIPE_VIEW_ALL, addRecipeLinkReader(recipeChannel))

  return reader, messageBox
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
