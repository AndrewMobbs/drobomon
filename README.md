# drobomon
## Basic health and status monitoring of a Drobo NAS

Calls the Drobo NAS status port and reports status back as a REST API.
	By default runs service on port 5000.
	Provides two endpoints:
	/v1/drobomon/health
	/v1/drobomon/status

	/health returns a HTTP 200 response for healthy or warning states, HTTP 500 for error
	/status returns a JSON object of the Drobo status
  
 Intended use is for monitoring a Drobo behind a firewall that blocks Drobo Admin UI.
