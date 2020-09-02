# drobomon
## Basic health and status monitoring of a Drobo NAS

Calls the Drobo NAS status port and reports status back as a REST API.
	By default runs service on port 5000.
	Provides two endpoints:  
	GET /v1/drobomon/health  
	GET /v1/drobomon/status

	/health returns a JSON status report and HTTP 200 response for healthy or warning states, HTTP 500 for error. If you see anything wrong, suggest plugging in the standard Drobo Admin UI for diagnosis.
	/status returns a JSON object of a select subset of the Drobo status fields. Could be used for a Javascript dashboard...
  
 Intended use is for monitoring a Drobo behind a firewall that blocks Drobo Admin UI.
 
 ## Warning
 Current project status is "it works for me". Has been tested on precisely one Drobo 5N with limited failure modes.  
 Many thanks to https://github.com/droboports/droboports.github.io/wiki/NASD-XML-format for help in decoding the Drobo XML data.
