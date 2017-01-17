package couchdb

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/segmentio/pointer"
)

var (
	client *Client
	c      *Client
	cView  *Client
	db     Database
	dbView Database
)

func TestMain(m *testing.M) {
	u, err := url.Parse("http://127.0.0.1:5984/")
	if err != nil {
		panic(err)
	}
	client, err = NewClient(u)
	if err != nil {
		panic(err)
	}
	c, err = NewClient(u)
	if err != nil {
		panic(err)
	}
	db = c.Use("dummy")
	cView, err = NewClient(u)
	if err != nil {
		panic(err)
	}
	dbView = cView.Use("gotest")
	code := m.Run()
	// clean up
	os.Exit(code)
}

func TestInfo(t *testing.T) {
	info, err := client.Info()
	if err != nil {
		t.Fatal(err)
	}
	if info.Couchdb != "Welcome" {
		t.Error("Couchdb error")
	}
}

func TestActiveTasks(t *testing.T) {
	res, err := client.ActiveTasks()
	if err != nil {
		t.Fatal(err)
	}
	out := make([]Task, 0)
	if !reflect.DeepEqual(out, res) {
		t.Error("active tasks should be an empty array")
	}
}

func TestAll(t *testing.T) {
	res, err := client.All()
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "_replicator" || res[1] != "_users" {
		t.Error("slice error")
	}
}

func TestGet(t *testing.T) {
	info, err := client.Get("_users")
	if err != nil {
		t.Fatal(err)
	}
	if info.DbName != "_users" {
		t.Error("DbName error")
	}
	if info.CompactRunning {
		t.Error("CompactRunning error")
	}
}

func TestCreate(t *testing.T) {
	status, err := client.Create("dummy")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Error("status error")
	}
}

func TestCreateFail(t *testing.T) {
	_, err := client.Create("dummy")
	if err == nil {
		t.Fatal("should not create duplicate database")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusPreconditionFailed {
			t.Fatal("should not create duplicate database")
		}
	}
}

func TestCreateUser(t *testing.T) {
	user := NewUser("john", "password", []string{})
	res, err := client.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok || res.ID != "org.couchdb.user:john" {
		t.Error("create user error")
	}
}

func TestCreateSession(t *testing.T) {
	res, err := client.CreateSession("john", "password")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok || res.Name != "john" {
		t.Error("create session error")
	}
}

func TestGetSession(t *testing.T) {
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if !session.Ok || session.UserContext.Name != "john" {
		t.Error("get session error")
	}
}

func TestDeleteSession(t *testing.T) {
	res, err := client.DeleteSession()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Error("delete session error")
	}
}

func TestGetUser(t *testing.T) {
	user, err := client.GetUser("john")
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "john" || user.Type != "user" || user.Iterations != 10 {
		t.Error("get user error")
	}
}

func TestDeleteUser(t *testing.T) {
	user, err := client.GetUser("john")
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.DeleteUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok || res.ID != "org.couchdb.user:john" {
		t.Error("delete user error")
	}
}

func TestGetSessionAdmin(t *testing.T) {
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if !session.Ok {
		t.Error("session response is false")
	}
	roles := []string{"_admin"}
	if !reflect.DeepEqual(roles, session.UserContext.Roles) {
		t.Error("session roles are wrong")
	}
}

func TestDelete(t *testing.T) {
	status, err := client.Delete("dummy")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Error("status error")
	}
}

func TestDeleteFail(t *testing.T) {
	_, err := client.Delete("dummy")
	if err == nil {
		t.Fatal("should not delete non existing database")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusNotFound {
			t.Fatal("should not delete non existing database")
		}
	}
}

func TestUse(t *testing.T) {
	db := client.Use("_users")
	if db.Name != "_users/" {
		t.Errorf("expected _users/ got %s", db.Name)
	}
}

type animal struct {
	Document
	Type   string `json:"type"`
	Animal string `json:"animal"`
	Owner  string `json:"owner"`
}

func TestReplication(t *testing.T) {
	name := "replication"
	name2 := "replication2"
	// create database
	res, err := client.Create(name)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", res)
	// add some documents to database
	db := client.Use(name)
	for _, a := range []string{"dog", "mouse", "cat"} {
		doc := &animal{
			Type:   "animal",
			Animal: a,
		}
		if _, err := db.Post(doc); err != nil {
			t.Error(err)
		}
	}
	// replicate
	req := ReplicationRequest{
		CreateTarget: true,
		Source:       "http://localhost:5984/" + name,
		Target:       "http://localhost:5984/" + name2,
	}
	r, err := c.Replicate(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", r)
	if !r.Ok {
		t.Error("expected ok to be true but got false instead")
	}
	// remove both databases
	for _, d := range []string{name, name2} {
		if _, err := client.Delete(d); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReplicationFilter(t *testing.T) {
	dbName := "replication_filter"
	dbName2 := "replication_filter2"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	// add some documents to database
	db := client.Use(dbName)
	docs := []animal{
		{
			Type:   "animal",
			Animal: "dog",
			Owner:  "john",
		},
		{
			Type:   "animal",
			Animal: "cat",
			Owner:  "john",
		},
		{
			Type:   "animal",
			Animal: "horse",
			Owner:  "steve",
		},
	}
	for _, doc := range docs {
		if _, err := db.Post(&doc); err != nil {
			t.Error(err)
		}
	}
	// create view document with filter function in first database
	designDocument := &DesignDocument{
		Document: Document{
			ID: "_design/animals",
		},
		Language: "javascript",
		Filters: map[string]string{
			"byOwner": `
				function(doc, req) {
					if (doc.owner === req.query.owner) {
						return true
					}
					return false
				}
			`,
		},
	}
	if _, err := db.Post(designDocument); err != nil {
		t.Error(err)
	}
	// create replication with filter function
	req := ReplicationRequest{
		CreateTarget: true,
		Source:       "http://localhost:5984/" + dbName,
		Target:       "http://localhost:5984/" + dbName2,
		Filter:       "animals/byOwner",
		QueryParams: map[string]string{
			"owner": "john",
		},
	}
	if _, err := c.Replicate(req); err != nil {
		t.Error(err)
	}
	// check replicated database
	db = client.Use(dbName2)
	allDocs, err := db.AllDocs(nil)
	if err != nil {
		t.Error(err)
	}
	if len(allDocs.Rows) != 2 {
		t.Errorf("expected exactly two documents but got %d instead", len(allDocs.Rows))
	}
	// remove both databases
	for _, d := range []string{dbName, dbName2} {
		if _, err := client.Delete(d); err != nil {
			t.Fatal(err)
		}
	}
}

// test continuous replication to test getting replication document
// with custom time format.
func TestReplicationContinuous(t *testing.T) {
	dbName := "continuous"
	dbName2 := "continuous2"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	// create replication document inside _replicate database
	req := ReplicationRequest{
		Document: Document{
			ID: "awesome",
		},
		Continuous:   true,
		CreateTarget: true,
		Source:       "http://localhost:5984/" + dbName,
		Target:       "http://localhost:5984/" + dbName2,
	}
	res, err := c.Replicate(req)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", res)
	tasks, err := c.ActiveTasks()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", tasks)
	if tasks[0].Type != "replication" {
		t.Errorf("expected type replication but got %s instead", tasks[0].Type)
	}
	// remove both databases
	for _, d := range []string{dbName, dbName2} {
		if _, err := client.Delete(d); err != nil {
			t.Fatal(err)
		}
	}
}

// database tests
type DummyDocument struct {
	Document
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
}

func TestBefore(t *testing.T) {
	_, err := client.Create("dummy")
	if err != nil {
		panic(err)
	}
}

func TestDocumentPost(t *testing.T) {
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	if doc.Rev != "" {
		t.Error("new document should not have a revision")
	}
	res, err := db.Post(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Error("post document error")
	}
}

func TestDocumentHead(t *testing.T) {
	head, err := db.Head("testid")
	if err != nil {
		t.Fatal(err)
	}
	if head.StatusCode != 200 {
		t.Error("document head error")
	}
}

func TestDocumentGet(t *testing.T) {
	doc := new(DummyDocument)
	err := db.Get(doc, "testid")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Foo != "bar" || doc.Beep != "bopp" {
		t.Error("document fields error")
	}
}

func TestDocumentPut(t *testing.T) {
	// get document
	doc := new(DummyDocument)
	err := db.Get(doc, "testid")
	if err != nil {
		t.Fatal(err)
	}
	// change document
	doc.Foo = "baz"
	res, err := db.Put(doc)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "testid" || !res.Ok {
		t.Error("put document response error")
	}
}

func TestDocumentDelete(t *testing.T) {
	// get document
	doc := new(DummyDocument)
	err := db.Get(doc, "testid")
	if err != nil {
		t.Fatal(err)
	}
	// delete document
	res, err := db.Delete(doc)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "testid" || !res.Ok {
		t.Error("delete document response error")
	}
}

func TestDocumentPutAttachment(t *testing.T) {
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	res, err := db.PutAttachment(doc, "./test/dog.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "testid" || !res.Ok {
		t.Error("put attachment error")
	}
}

// Test added because updating an existing document that had an attachment caused an error.
// After adding more fields to Attachment struct it now works.
func TestUpdateDocumentWithAttachment(t *testing.T) {
	// get existing document
	doc := &DummyDocument{}
	err := db.Get(doc, "testid")
	if err != nil {
		t.Fatal(err)
	}
	// update document with attachment
	doc.Foo = "awesome"
	res, err := db.Put(doc)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "testid" || !res.Ok {
		t.Error("put document response error")
	}
}

func TestDocumentBulkDocs(t *testing.T) {
	// first dummy document
	doc1 := &DummyDocument{
		Foo:  "foo1",
		Beep: "beep1",
	}
	// second dummy document
	doc2 := &DummyDocument{
		Foo:  "foo2",
		Beep: "beep2",
	}
	// slice of dummy document
	docs := []CouchDoc{doc1, doc2}

	res, err := db.Bulk(docs)
	if err != nil {
		t.Fatal(err)
	}
	if !res[0].Ok || !res[1].Ok {
		t.Error("bulk docs error")
	}
}

func TestAllDocs(t *testing.T) {
	res, err := db.AllDocs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 3 {
		t.Errorf("expected total rows equals 3 but got %v", res.TotalRows)
	}
	if len(res.Rows) != 3 {
		t.Errorf("expected length rows equals 3 but got %v", len(res.Rows))
	}
}

func TestPurge(t *testing.T) {
	dbName := "purge"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	db := client.Use(dbName)
	// create documents
	doc := &DummyDocument{
		Foo:  "bar",
		Beep: "bopp",
	}
	postResponse, err := db.Post(doc)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", postResponse)
	// purge
	req := map[string][]string{
		postResponse.ID: {
			postResponse.Rev,
		},
	}
	purgeResponse, err := db.Purge(req)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", purgeResponse)
	if purgeResponse.PurgeSeq != 1 {
		t.Errorf("expected purge seq to be 1 but got %v instead", purgeResponse.PurgeSeq)
	}
	revisions, ok := purgeResponse.Purged[postResponse.ID]
	if !ok {
		t.Error("expected to find entry at post response ID but could not find any")
	}
	if revisions[0] != postResponse.Rev {
		t.Error("expected purged revision to be the same as posted document revision")
	}
	// remove database
	if _, err := client.Delete(dbName); err != nil {
		t.Error(err)
	}
}

func TestSecurity(t *testing.T) {
	dbName := "sec"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	db := client.Use(dbName)
	// test putting security document first
	secDoc := SecurityDocument{
		Admins: Element{
			Names: []string{
				"admin1",
			},
			Roles: []string{
				"",
			},
		},
		Members: Element{
			Names: []string{
				"member1",
			},
			Roles: []string{
				"",
			},
		},
	}
	res, err := db.PutSecurity(secDoc)
	if err != nil {
		t.Error(err)
	}
	if !res.Ok {
		t.Error("expected true but got false")
	}
	// test getting security document
	doc, err := db.GetSecurity()
	if err != nil {
		t.Error(err)
	}
	if doc.Admins.Names[0] != "admin1" {
		t.Errorf("expected name admin1 but got %s instead", doc.Admins.Names[0])
	}
	if doc.Members.Names[0] != "member1" {
		t.Errorf("expected name member1 but got %s instead", doc.Members.Names[0])
	}
	// remove database
	if _, err := client.Delete(dbName); err != nil {
		t.Error(err)
	}
}

func TestAfter(t *testing.T) {
	t.Log("deleting dummy database")
	_, err := client.Delete("dummy")
	if err != nil {
		t.Fatal(err)
	}
}

// end database tests

// view tests
type DataDocument struct {
	Document
	Type string `json:"type"`
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
	Age  int    `json:"age"`
}

type Person struct {
	Document
	Type   string  `json:"type"`
	Name   string  `json:"name"`
	Age    float64 `json:"age"`
	Gender string  `json:"gender"`
}

func TestViewBefore(t *testing.T) {
	// create database
	if _, err := cView.Create("gotest"); err != nil {
		t.Fatal(err)
	}
	design := &DesignDocument{
		Document: Document{
			ID: "_design/test",
		},
		Language: "javascript",
		Views: map[string]DesignDocumentView{
			"foo": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit(doc.foo);
						}
					}
				`,
			},
			"int": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit([doc.foo, doc.age]);
						}
					}
				`,
			},
			"complex": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit([doc.foo, doc.beep]);
						}
					}
				`,
			},
		},
	}
	if _, err := dbView.Post(design); err != nil {
		t.Fatal(err)
	}
	// create design document for person
	designPerson := DesignDocument{
		Document: Document{
			ID: "_design/person",
		},
		Language: "javascript",
		Views: map[string]DesignDocumentView{
			"ageByGender": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'person') {
							emit(doc.gender, doc.age);
						}
					}
				`,
				Reduce: `
					function(keys, values, rereduce) {
						return sum(values);
					}
				`,
			},
		},
	}
	if _, err := dbView.Post(&designPerson); err != nil {
		t.Fatal(err)
	}
	// create dummy data
	doc1 := &DataDocument{
		Type: "data",
		Foo:  "foo1",
		Beep: "beep1",
		Age:  10,
	}
	if _, err := dbView.Post(doc1); err != nil {
		t.Fatal(err)
	}
	doc2 := &DataDocument{
		Type: "data",
		Foo:  "foo2",
		Beep: "beep2",
		Age:  20,
	}
	if _, err := dbView.Post(doc2); err != nil {
		t.Fatal(err)
	}
	// create multiple persons
	data := []struct {
		Name   string
		Age    float64
		Gender string
	}{
		{"John", 45, "male"},
		{"Frank", 40, "male"},
		{"Steve", 60, "male"},
		{"Max", 26, "male"},
		{"Marc", 36, "male"},
		{"Nick", 18, "male"},
		{"Jessica", 49, "female"},
		{"Lily", 20, "female"},
		{"Sophia", 66, "female"},
		{"Chloe", 12, "female"},
	}
	people := make([]CouchDoc, len(data))
	for index, d := range data {
		people[index] = &Person{
			Type:   "person",
			Name:   d.Name,
			Age:    d.Age,
			Gender: d.Gender,
		}
	}
	// bulk save people to database
	if _, err := dbView.Bulk(people); err != nil {
		t.Fatal(err)
	}
}

func TestViewGet(t *testing.T) {
	view := dbView.View("test")
	params := QueryParameters{}
	res, err := view.Get("foo", params)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 2 || res.Offset != 0 {
		t.Error("view get error")
	}
}

func TestDesignDocumentName(t *testing.T) {
	doc := new(DesignDocument)
	err := dbView.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name() != "test" {
		t.Error("design document Name() error")
	}
}

func TestDesignDocumentView(t *testing.T) {
	doc := new(DesignDocument)
	err := dbView.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := doc.Views["foo"]
	if !ok {
		t.Error("design document view error")
	}
}

func TestViewGetWithQueryParameters(t *testing.T) {
	view := dbView.View("test")
	params := QueryParameters{
		Key: pointer.String(fmt.Sprintf("%q", "foo1")),
	}
	res, err := view.Get("foo", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewGetWithStartKeyEndKey(t *testing.T) {
	view := dbView.View("test")
	params := QueryParameters{
		StartKey: pointer.String(fmt.Sprintf("[%q,%q]", "foo2", "beep2")),
		EndKey:   pointer.String(fmt.Sprintf("[%q,%q]", "foo2", "beep2")),
	}
	res, err := view.Get("complex", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewGetWithInteger(t *testing.T) {
	view := dbView.View("test")
	params := QueryParameters{
		StartKey: pointer.String(fmt.Sprintf("[%q,%d]", "foo2", 20)),
		EndKey:   pointer.String(fmt.Sprintf("[%q,%d]", "foo2", 20)),
	}
	res, err := view.Get("int", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewGetWithReduce(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{}
	res, err := view.Get("ageByGender", params)
	if err != nil {
		t.Fatal(err)
	}
	ageTotalSum := res.Rows[0].Value.(float64)
	if ageTotalSum != 372 {
		t.Fatalf("expected age 372 but got %v", ageTotalSum)
	}
}

func TestViewGetWithReduceAndGroup(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{
		Key:        pointer.String(fmt.Sprintf("%q", "female")),
		GroupLevel: pointer.Int(1),
	}
	res, err := view.Get("ageByGender", params)
	if err != nil {
		t.Fatal(err)
	}
	ageTotalFemale := res.Rows[0].Value.(float64)
	if ageTotalFemale != 147 {
		t.Fatalf("expected age 147 but got %v", ageTotalFemale)
	}
}

func TestViewGetWithoutReduce(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{
		Key:    pointer.String(fmt.Sprintf("%q", "male")),
		Reduce: pointer.Bool(false),
	}
	res, err := view.Get("ageByGender", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 6 {
		t.Fatalf("expected 6 rows but got %d instead", len(res.Rows))
	}
}

func TestViewPost(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{
		Reduce: pointer.Bool(false),
	}
	res, err := view.Post("ageByGender", []string{"male"}, params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 6 {
		t.Fatalf("expected 6 rows but got %d instead", len(res.Rows))
	}
}

func TestViewAfter(t *testing.T) {
	if _, err := cView.Delete("gotest"); err != nil {
		t.Fatal(err)
	}
}

// test utils

// mimeType()
var mimeTypeTests = []struct {
	in  string
	out string
}{
	{"image.jpg", "image/jpeg"},
	{"presentation.pdf", "application/pdf"},
	{"file.text", "text/plain; charset=utf-8"},
	{"archive.zip", "application/zip"},
	{"movie.avi", "video/x-msvideo"},
}

func TestMimeType(t *testing.T) {
	for _, tt := range mimeTypeTests {
		actual := mimeType(tt.in)
		if actual != tt.out {
			t.Errorf("mimeType(%s): expected %s, actual %s", tt.in, tt.out, actual)
		}
	}
}
