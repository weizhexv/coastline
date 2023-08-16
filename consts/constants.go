package consts

const (
	DomainPlatform = "PLATFORM"
	DomainSupplier = "SUPPLIER"
	DomainBusiness = "BUSINESS"
	DomainEmployee = "EMPLOYEE"

	//headers need acquire from uc
	HeaderDomain    = "domain"
	HeaderDomainId  = "domain-id"
	HeaderRole      = "role"
	HeaderUid       = "uid"
	HeaderOpId      = "op-id"
	HeaderPoweredBy = "Powered-By"
	HeaderRemoteIP  = "remote-ip"

	//transparent headers below (need pass from client to upstream)
	HeaderTraceId     = "trace-id"
	HeaderToken       = "token"
	HeaderLang        = "lang"
	HeaderContentType = "Content-Type"
)
