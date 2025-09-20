// Package sslkeylogfile provides a way to log SSL/TLS keys to a file for debugging purposes.
//
// Functions in this package are no-ops unless the SSLKEYLOGFILE environment variable
// is set to a valid filename. If SSLKEYLOGFILE is set, the package will create (or append to)
// the specified file and write SSL/TLS keys to it in a format compatible with Wireshark and
// other tools that support SSLKEYLOGFILE.
//
// To support multiple concurrent writers (e.g., multiple HTTP clients), a file per tls.Config
// is created based on the SSLKEYLOGFILE pattern, appending a sequence number to the filename
// if necessary (e.g., sslkeylogfile, sslkeylogfile.1, sslkeylogfile.2, etc.).
//
// The package provides functions to create tls.Config and http.Transport instances
// with KeyLogWriter set appropriately, as well as a function to patch http.DefaultTransport.
//
// Example usage:
//
//	import (
//	    "crypto/tls"
//	    "net/http"
//	    "os"
//
//	    "github.com/fiam/sslkeylogfile"
//	)
//
//	func main() {
//	    // Option 1: Create a tls.Config with KeyLogWriter set
//	    tlsConfig, err := sslkeylogfile.NewTLSConfig()
//	    if err != nil {
//	        panic(err)
//	    }
//	    _ = tlsConfig // use tlsConfig in your TLS clients
//
//	    // Option 2: Create an http.Transport with TLSClientConfig.KeyLogWriter set
//	    transport, err := sslkeylogfile.NewTransport()
//	    if err != nil {
//	        panic(err)
//	    }
//	    client := &http.Client{Transport: transport}
//	    _ = client // use client for HTTP requests
//
//	    // Option 3: Patch http.DefaultTransport to log TLS keys
//	    if err := sslkeylogfile.PatchDefaultTransport(); err != nil {
//	        panic(err)
//	    }
//	    resp, err := http.Get("https://www.example.com/")
//	    if err != nil {
//	        panic(err)
//	    }
//	    defer resp.Body.Close()
//	    // process response...
//	}
//
// Note: Ensure that the SSLKEYLOGFILE environment variable is set before running your application.
// For example, in a Unix-like shell:
//
//	export SSLKEYLOGFILE=/path/to/your/sslkeylogfile.log
//
// Be cautious when using this package in production environments, as logging SSL/TLS keys
// can expose sensitive information.
package sslkeylogfile
