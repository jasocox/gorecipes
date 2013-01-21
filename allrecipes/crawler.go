package allrecipes

const HOSTNAME = "allrecipes.com"

func NewReader() <-chan string {
  r := make(chan string)
  go func() {
    for {
      r <- "Hello World"
    }
  }()

  return r
}
