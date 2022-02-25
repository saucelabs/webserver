// Content-Type MIME of the most common data formats.
//
// SEE: https://github.com/golang/go/issues/31572

package webserver

const (
	MIMEHTML              = "text/html"
	MIMEJSON              = "application/json"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEYAML              = "application/x-yaml"
)
