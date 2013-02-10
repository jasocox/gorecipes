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

  linkChannel := make(chan string, 1000)
  addRecipeLinkReader(linkChannel, recipeChannel)
  go startRecipeLinkFinder(RECIPE_VIEW_ALL, linkChannel)

  return reader, messageBox
}

func addRecipeLinkReader(linkFindingChannel <-chan string, recipeChannel chan<- *recipe.Recipe) {
  addRecipeLinkReaders(linkFindingChannel, recipeChannel, 1)
}

func addRecipeLinkReaders(linkFindingChannel <-chan string, recipeChannel chan<- *recipe.Recipe, processes int) {
  for i:=0; i<processes; i++ {
    go func() {
      recipeLinkHash := make(map[string]string)
      for {
        recipeLink := <-linkFindingChannel

        if recipeLinkHash[recipeLink] == "" {
          recipeChannel <- readRecipeLink(recipeLink)
          recipeLinkHash[recipeLink] = recipeLink
        }
      }
    }()
  }
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

func startRecipeLinkFinder(url string, recipeLinkChannel chan<- string) {
  go findRecipeLinksFromUrlAndFollowNext(url, recipeLinkChannel)
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
