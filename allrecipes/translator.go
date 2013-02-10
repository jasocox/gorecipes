package allrecipes

import "regexp"

var (
  translatorMap map[string]Translator
)

type Translator struct {
  Name string
  Translator func(string) interface{}
}

func generateTranslators(translatorConfig [][]interface{}) {
  translatorMap = make(map[string]Translator)

  for _, translator := range translatorConfig {
    addTranslator(translator[0].(string),
      generateTranslator(translator[1].(string), translator[2].(func(string, *regexp.Regexp) interface{})))
  }
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

func translateHtml(name string, body string) interface{} {
  translator := translatorMap[name]

  return translator.Translator(body)
}
