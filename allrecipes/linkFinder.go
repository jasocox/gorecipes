package allrecipes

import "log"

func findRecipeLinksFromUrlAndFollowNext(url string, recipeLinkChannel chan<- string, messageBox chan string) {
  log.Println(url + ": Starting")
  body, err := readBodyFromUrl(url)
  if err != nil {
    log.Println(url + ": Failed to read page of recipe links")
    return
  }

  nextLink := translateHtml("Next", body).(string)
  if nextLink != "" {
    log.Println(url + ": Found next link")
    go findRecipeLinksFromUrlAndFollowNext(nextLink, recipeLinkChannel, messageBox)
  } else {
    log.Println(url + ": Didn't find a next link")
  }

  go findLinksFromBodyAndSend(url, body, recipeLinkChannel)

  log.Println(url + ": Done")
}

func findLinksFromBodyAndSend(url string, body string, recipeLinkChannel chan<- string) {
  recipes := translateHtml("RecipeLink", body).([]string)
  for recipe := range recipes {
    recipeLinkChannel <- recipes[recipe]
  }
}
