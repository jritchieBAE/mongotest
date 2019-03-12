package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

var (
	port     int
	rootPath string
	certPath string
	keyPath  string
)

func getFlags() {
	flag.IntVar(&port, "p", 27017, "mongodb port")
	flag.StringVar(&rootPath, "root", "", "path to root certificate")
	flag.StringVar(&certPath, "cert", "", "path to client certificate")
	flag.StringVar(&keyPath, "key", "", "path to client key")
	flag.Parse()
}

func main() {

	getFlags()

	dialInfo := &mgo.DialInfo{
		Addrs: []string{fmt.Sprintf("localhost:%v", port)},
	}
	if len(rootPath) > 0 {
		tlsConfig := getTLSConfig()
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), tlsConfig)
		}
	}

	log.Println("Making session")
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	coll := session.DB("test").C("people")

	// record := &struct {
	// 	Name  string
	// 	Grade string
	// }{"Bob", "T3"}

	// coll.Insert(record)

	var records []struct {
		Name  string
		Grade string
	}
	log.Println("Querying...")

	err = coll.Find(bson.M{"grade": "T3"}).All(&records)
	if err != nil {
		log.Fatal("record not found")
	}
	fmt.Println(records)

}

func getTLSConfig() *tls.Config {

	return &tls.Config{
		RootCAs: func() *x509.CertPool {
			if len(rootPath) > 0 {
				log.Println("Loading root certificate")
				ca, err := ioutil.ReadFile(rootPath)
				if err != nil {
					return nil
				}
				rootCAs := x509.NewCertPool()
				if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
					panic("Failed to parse CA certificate")
				}
				return rootCAs
			}
			return nil
		}(),
		Certificates: func() []tls.Certificate {
			if len(certPath) > 0 {

				log.Println("Loading certificate for mTLS")
				cert, err := tls.LoadX509KeyPair(certPath, keyPath)
				if err != nil {
					panic("Failed to load client certificate")
				}
				return []tls.Certificate{cert}
			}
			return nil
		}(),
	}
}
