package main

import (
  "import github.com/zemirco/couchdb"
  "fmt"
)

// create your own document
type DummyDocument struct {
  couchdb.Document
  Foo string `json:"foo"`
  Beep string `json:"beep"`
}

// make sure your own document confirms to type CouchDoc
func (doc *DummyDocument) GetDocument() *couchdb.Document {
  return &doc.Document
}

// just some helper function
func check(err error) {
  if err != nil {
    panic(err)
  }
}

// start
func main() {

  // create a new client
  client, err := couchdb.NewClient("http://127.0.0.1:5984/")
  check(err)

  // get some information about your CouchDB
  info, err := client.Info()
  check(err)
  fmt.Println(info)

  // create a database
  _, err = client.Create("dummy")
  check(err)

  // use your new "dummy" database and create a document
  db := client.Use("dummy")
  doc := &DummyDocument{
    Foo: "bar",
    Beep: "bopp",
  }
  result, err := db.Post(doc)
  check(err)

  // get id and current revision.
  err = db.Get(doc, result.Id)
  check(err)

  // delete document
  _, err = db.Delete(doc)
  check(err)

  // and finally delete the database
  _, err = client.Delete("dummy")
  check(err)

}
