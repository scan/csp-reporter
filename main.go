package main

type CSPReport struct {
    BlockedURI string `json:"blocked-uri"`
    Disposition string `json:"disposition"`
    DocumentURI string `json:"document-uri"`
    EffectiveDirective string `json:"effective-directive"`
	OriginalPolicy string `json:"original-policy"`
	Referrer string `json:"referrer"`
	ScriptSample string `json:"script-sample"`
	StatusCode string `json:"status-code"`
	ViolatedDirective string `json:"violated-directive"`
}