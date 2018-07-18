package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"crypto/tls"
	"crypto/x509"
)

var (
	token        = flag.String("token", os.Getenv("TOKEN"), "The Service Account token")
	promSvc      = flag.String("prometheus-svc", os.Getenv("PROMETHEUS_SERVICE"), "The Prometheus service URL")
	skipInsecure = flag.String("skip-insecure", os.Getenv("SKIP_INSECURE_VERIFY"), "Useful if your CA is not signed by an Authority")
	client       *http.Client
)

type logWriter struct {
}

// Formatting the logger interface according to customer needs: feel free to edit
func (writer *logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02 15:05:05.000-0700") + " INFO [proxy] (prometheus) " + string(bytes))
}

// Handling the decorated request with custom Bearer token and returning the response:
// just a simple proxy.
// TODO: handle redirects and headers
func proxy(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", *promSvc, r.RequestURI), r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	w.WriteHeader(resp.StatusCode)
	w.Write([]byte(string(body)))
}

func init() {
	flag.Parse()

	if *token == "" {
		panic("Missing bearer token: exiting")
	}
	if *promSvc == "" {
		panic("Missing Prometheus service: exiting")
	}

	client = initClient()

	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

func initClient() *http.Client {
	// Load CA cert
	caCert, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	if parseOk := caCertPool.AppendCertsFromPEM(caCert); !parseOk {
		panic("Error parsing service account CA certificate")
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: *skipInsecure != "",
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{Transport: transport}
}

// Setup logger interface and provide a simple validation: if everything is fine start serving 127.0.0.1:9090
// (Prometheus standard port and binding on loopback due to Pod network share according to Ambassador pattern)
func main() {
	log.Println("Serving for " + *promSvc)

	http.HandleFunc("/", proxy)
	// TODO: enabling listening only on loopback
	log.Println(http.ListenAndServe("127.0.0.1:9090", nil))
}
