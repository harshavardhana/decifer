package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/vulcand/oxy/forward"
)

type responseWriter struct {
	S cipher.Stream
	W http.ResponseWriter
}

func (r responseWriter) Header() http.Header {
	r.W.Header().Del("Content-Length")
	return r.W.Header()
}

func (r responseWriter) WriteHeader(i int) {
	r.W.WriteHeader(i)
}

func (r responseWriter) Write(src []byte) (n int, err error) {
	r.S.XORKeyStream(src, src)
	n, err = r.W.Write(src)
	if n != len(src) {
		if err == nil {
			// should never happen
			err = io.ErrShortWrite
		}
	}
	return
}

// To log the request headers and body to Stdout.
type decipher struct {
	key []byte
	iv  []byte
	h   http.Handler
}

const slashSeparator = "/"

func hasObject(path string) bool {
	return len(strings.Split(path, slashSeparator)) > 2
}

// Verify if request has AWS PreSign Version '4'.
func isRequestPresignedSignatureV4(u *url.URL) bool {
	_, ok := u.Query()["X-Amz-Credential"]
	return ok
}

func isPresignedGet(r *http.Request) bool {
	return r.Method == "GET" && hasObject(r.URL.Path) && isRequestPresignedSignatureV4(r.URL)
}

func (l decipher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if isPresignedGet(r) {
		w.Header().Del("Content-Length")
		block, err := aes.NewCipher(l.key)
		if err != nil {
			panic(err)
		}
		// If the key is unique for each ciphertext, then it's ok to use a zero IV.
		var iv [aes.BlockSize]byte
		stream := cipher.NewOFB(block, iv[:])
		l.h.ServeHTTP(responseWriter{stream, w}, r)
	} else {
		l.h.ServeHTTP(w, r)
	}
}

// To forward the request to the address specified with -f
type forwarder struct {
	scheme string
	host   string
	h      http.Handler
}

func (f forwarder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Scheme = f.scheme
	r.URL.Host = f.host
	f.h.ServeHTTP(w, r)
}

func main() {
	listenAddr := flag.String("l", ":8000", "listen address")
	fwdAddr := flag.String("f", "localhost:9000", "forward address")

	keyFile := flag.String("master-key", "master.key", "master AES key")
	sslCert := flag.String("ssl-cert", "", "certificate")
	sslKey := flag.String("ssl-key", "", "private key")
	flag.Parse()

	key, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		panic(err)
	}

	iv, err := base64.StdEncoding.DecodeString("nAsVgD8i3ujDs585L1XWiA==")
	if err != nil {
		panic(err)
	}

	fwd, _ := forward.New(
		forward.PassHostHeader(true),
	)
	server := &http.Server{
		Addr: *listenAddr,
		Handler: decipher{
			h:   forwarder{"http", *fwdAddr, fwd},
			key: key,
			iv:  iv,
		},
	}

	if *sslCert != "" && *sslKey != "" {
		fmt.Println(server.ListenAndServeTLS(*sslCert, *sslKey))
	} else {
		fmt.Println(server.ListenAndServe())
	}
}
