package main

import (
  "database/sql"
  "log"
  "net/http"
  "os"
  "sync"
  "fmt"
  _ "github.com/lib/pq"
  _ "net/http/pprof"

  "github.com/reviewed/brush/lex"
)

const poolSize int = 70

type lexingTest struct {
  body string
  slug string
  ok bool
  pos int
}

func main() {
  db := dbConnect()
  rows, err := db.Query("select body, name from article_sections order by created_at desc")
  if err != nil {
    log.Fatal(err)
  }

  go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
  }()

  var bodies = make(chan lexingTest)
  var results = make(chan bool)
  var completion = make(chan bool)

  var wg sync.WaitGroup
  for i := 0; i < poolSize; i++ {
    wg.Add(1)
    go lexBody(completion, bodies, results)
  }

  go collectResults(results, completion)

  go func() {
    var body, slug string
    for rows.Next() {
      rows.Scan(&body, &slug)
      bodies <- lexingTest{body, slug, false, 0}
    }
    close(bodies)
  }()

  for _ = range results{
  }
}

func collectResults(results chan bool, kill chan bool) {
  succeeded := 0
  failed := 0
  killcount := 0
  for {
    select {
    case result := <-results:
      if result == true {
        succeeded++
      } else {
        failed++
      }
      if total := succeeded + failed; total % 100 == 0 {
        fmt.Printf("Completed: %d, %d/%d\n", total, failed, succeeded)
      }
    case <-kill:
      killcount++
      if killcount == poolSize {
        fmt.Printf("Succeeded: %d :: Failed %d", succeeded, failed)
        close(results)
        return
      }
    }
  }
}

func lexBody(completion chan bool, bodyChan chan lexingTest, resultChan chan bool) {
  for test := range bodyChan {
    defer func() {
      if r := recover(); r != nil {
        fmt.Printf("Body: %s", test.body)
      }
    }()
    lexer := lex.NewLexer(test.body)
    var result bool
    for {
      tok := lexer.NextToken()
      if int(tok.Type) == 10 {
        result = true
        break
      }
      if int(tok.Type) == 11 {
        //fmt.Println(test.body)
        //fmt.Println("===")
        result = false
        break
      }
    }
    resultChan <- result
  }
  completion <- true
}

func dbConnect() *sql.DB {
  os.Setenv("PGDATABASE", "reviewed_the_guide_development")
  os.Setenv("PGSSLMODE", "disable")
  conn, err := sql.Open("postgres", "")
  if err != nil {
    log.Fatal(err)
  }
  return conn
}
