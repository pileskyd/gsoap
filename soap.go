package gsoap

import (
	"encoding/json"
	"encoding/xml"
	"log"
)

const (
	SoapVersion11 = "1.1"
	SoapVersion12 = "1.2"

	SoapContentType11 = "text/xml; charset=\"utf-8\""
	SoapContentType12 = "application/soap+xml; charset=\"utf-8\""

	NamespaceSoap11 = "http://schemas.xmlsoap.org/soap/envelope/"
	NamespaceSoap12 = "http://www.w3.org/2003/05/soap-envelope"
)

// Verbose be verbose
var Verbose = false

func l(m ...interface{}) {
	if Verbose {
		log.Println(m...)
	}
}

func LogJSON(v interface{}) {
	if Verbose {
		jsonData, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			log.Println("Could not log json...")
			return
		}
		log.Println(string(jsonData))
	}
}

// Envelope type
type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Header  Header
	Body    Body
}

// Header type
type Header struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`

	Header interface{}
}

// Body type
type Body struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`

	Fault               *Fault      `xml:",omitempty"`
	Content             interface{} `xml:",omitempty"`
	SOAPBodyContentType string      `xml:"-"`
}

// Fault type
type Fault struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

// UnmarshalXML implement xml.Unmarshaler
func (b *Body) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if b.Content == nil {
		return xml.UnmarshalError("Content must be a pointer to a struct")
	}

	var consumed bool

	for {
		token, err := d.Token()
		if err != nil {
			return err
		}
		if token == nil {
			break
		}
		switch se := token.(type) {
		case xml.StartElement:
			if consumed {
				return xml.UnmarshalError("Found multiple elements inside SOAP body; not wrapped-document/literal WS-I compliant")
			}
			if se.Name.Space == "http://schemas.xmlsoap.org/soap/envelope/" && se.Name.Local == "Fault" {
				b.Fault = &Fault{}
				b.Content = nil

				err = d.DecodeElement(b.Fault, &se)
				if err != nil {
					return err
				}
				consumed = true
			} else {
				b.SOAPBodyContentType = se.Name.Local
				err = d.DecodeElement(b.Content, &se)
				if err != nil {
					return err
				}
				consumed = true
			}
		case xml.EndElement:
			return nil
		}
	}
	return nil
}

func (f *Fault) Error() string {
	return f.String
}
