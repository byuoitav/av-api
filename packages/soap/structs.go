package soap

// SoapEnvelope wraps XML SOAP requests
type SoapEnvelope struct {
	XMLName struct{} `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SoapBody
}

// SoapBody is where The XML body goes
type SoapBody struct {
	XMLName  struct{} `xml:"Body"`
	Contents []byte   `xml:",innerxml"`
}
