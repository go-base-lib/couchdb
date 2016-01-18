package couchdb

import (
	"fmt"
	"strings"
)

// http://docs.couchdb.org/en/latest/intro/api.html#server
type Server struct {
	Couchdb string
	Uuid    string
	Vendor  struct {
		Version string
		Name    string
	}
	Version string
}

// http://docs.couchdb.org/en/latest/api/database/common.html#get--db
type DatabaseInfo struct {
	DbName             string `json:"db_name"`
	DocCount           int    `json:"doc_count"`
	DocDelCount        int    `json:"doc_del_count"`
	UpdateSeq          int    `json:"update_seq"`
	PurgeSeq           int    `json:"purge_seq"`
	CompactRunning     bool   `json:"compact_running"`
	DiskSize           int    `json:"disk_size"`
	DataSize           int    `json:"data_size"`
	InstanceStartTime  string `json:"instance_start_time"`
	DiskFormatVersion  int    `json:"disk_format_version"`
	CommittedUpdateSeq int    `json:"committed_update_seq"`
}

type DatabaseResponse struct {
	Ok bool
}

type Error struct {
	Method     string
	Url        string
	StatusCode int
	Type       string `json:"error"`
	Reason     string
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"CouchDB - %s %s, Status Code: %d, Error: %s, Reason: %s",
		e.Method,
		e.Url,
		e.StatusCode,
		e.Type,
		e.Reason,
	)
}

type CouchDoc interface {
	GetDocument() *Document
}

type Document struct {
	Id          string                `json:"_id,omitempty"`
	Rev         string                `json:"_rev,omitempty"`
	Attachments map[string]Attachment `json:"_attachments,omitempty"`
}

func (d *Document) GetDocument() *Document {
	return d
}

type DesignDocument struct {
	Document
	Language string                        `json:"language,omitempty"`
	Views    map[string]DesignDocumentView `json:"views,omitempty"`
}

func (dd DesignDocument) Name() string {
	return strings.TrimPrefix(dd.Id, "_design/")
}

type DesignDocumentView struct {
	Map    string `json:"map,omitempty"`
	Reduce string `json:"reduce,omitempty"`
}

// http://docs.couchdb.org/en/latest/api/document/common.html#creating-multiple-attachments
type Attachment struct {
	Follows     bool   `json:"follows"`
	ContentType string `json:"content_type"`
	Length      int64  `json:"length"`
}

type DocumentResponse struct {
	Ok  bool
	Id  string
	Rev string
}

// http://docs.couchdb.org/en/latest/api/server/common.html#active-tasks
type Task struct {
	ChangesDone  int `json:"changes_done"`
	Database     string
	Pid          string
	Progress     int
	StartedOn    int `json:"started_on"`
	Status       string
	Task         string
	TotalChanges int `json:"total_changes"`
	Type         string
	UpdatedOn    string `json:"updated_on"`
}

type QueryParameters struct {
	Conflicts       bool   `url:"conflicts"`
	Descending      bool   `url:"descending"`
	EndKey          string `url:"endkey,comma,omitempty"`
	EndKeyDocId     string `url:"end_key_doc_id,omitempty"`
	Group           bool   `url:"group"`
	GroupLevel      int    `url:"group_level,omitempty"`
	IncludeDocs     bool   `url:"include_docs"`
	Attachments     bool   `url:"attachments"`
	AttEncodingInfo bool   `url:"att_encoding_info"`
	InclusiveEnd    bool   `url:"inclusive_end"`
	Key             string `url:"key,omitempty"`
	Limit           int    `url:"limit,omitempty"`
	Reduce          bool   `url:"reduce"`
	Skip            int    `url:"skip"`
	Stale           string `url:"stale,omitempty"`
	StartKey        string `url:"startkey,comma,omitempty"`
	StartKeyDocId   string `url:"startkey_docid,omitempty"`
	UpdateSeq       bool   `url:"update_seq"`
}

// NewQueryParameters returns query parameters with default values
// http://docs.couchdb.org/en/1.6.1/api/ddoc/views.html#get--db-_design-ddoc-_view-view
// The problem is "reduce" for example. The default value is true.
// If we have a map/reduce function that has a reduce part everything works as expected.
// We'll get into trouble if we want to reuse this document without reduce.
// If we use the omitempty flag on the Reduce field it would get it's zero value false
// which would not be sent to the server.
func NewQueryParameters() QueryParameters {
	// reduce is the exception. the default would be true
	// but as have have more cases where we don't have a reduce function we set it false
	// set it to true if you really need it.
	return QueryParameters{
		Conflicts:       false,
		Descending:      false,
		Group:           false,
		IncludeDocs:     false,
		Attachments:     false,
		AttEncodingInfo: false,
		InclusiveEnd:    true,
		Reduce:          false,
		Skip:            0,
		UpdateSeq:       false,
	}
}

type ViewResponse struct {
	Offset    int   `json:"offset,omitempty"`
	Rows      []Row `json:"rows,omitempty"`
	TotalRows int   `json:"total_rows,omitempty"`
	UpdateSeq int   `json:"update_seq,omitempty"`
}

type Row struct {
	Id    string                 `json:"id"`
	Key   interface{}            `json:"key"`
	Value interface{}            `json:"value,omitempty"`
	Doc   map[string]interface{} `json:"doc,omitempty"`
}

// http://docs.couchdb.org/en/latest/api/database/bulk-api.html#post--db-_bulk_docs
type BulkDoc struct {
	AllOrNothing bool          `json:"all_or_nothing,omitempty"`
	Docs         []interface{} `json:"docs"`
	NewEdits     bool          `json:"new_edits,omitempty"`
}

// http://docs.couchdb.org/en/latest/api/server/authn.html#cookie-authentication
type Credentials struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type PostSessionResponse struct {
	Ok    bool
	Name  string
	Roles []string
}

type User struct {
	Document
	DerivedKey     string   `json:"derived_key,omitempty"`
	Name           string   `json:"name,omitempty"`
	Roles          []string `json:"roles"`
	Password       string   `json:"password,omitempty"`     // plain text password when creating the user
	PasswordSha    string   `json:"password_sha,omitempty"` // hashed password when requesting user information
	PasswordScheme string   `json:"password_scheme,omitempty"`
	Salt           string   `json:"salt,omitempty"`
	Type           string   `json:"type,omitempty"`
	Iterations     int      `json:"iterations,omitempty"`
}

func NewUser(name, password string, roles []string) User {
	user := User{
		Document: Document{
			Id: "org.couchdb.user:" + name,
		},
		DerivedKey:     "",
		Name:           name,
		Roles:          roles,
		Password:       password,
		PasswordSha:    "",
		PasswordScheme: "",
		Salt:           "",
		Type:           "user",
	}
	return user
}

type GetSessionResponse struct {
	Info struct {
		Authenticated          string   `json:"authenticated"`
		AuthenticationDb       string   `json:"authentication_db"`
		AuthenticationHandlers []string `json:"authentication_handlers"`
	} `json:"info"`
	Ok          bool `json:"ok"`
	UserContext struct {
		Db    string   `json:"db"`
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
	} `json:"userCtx"`
}
