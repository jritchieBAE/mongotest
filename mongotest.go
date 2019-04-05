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
)

var (
	port          int
	rootPath      string
	certPath      string
	keyPath       string
	resetDatabase bool
)

func getFlags() {
	flag.IntVar(&port, "p", 27017, "mongodb port")
	flag.StringVar(&rootPath, "root", "", "path to root certificate")
	flag.StringVar(&certPath, "cert", "", "path to client certificate")
	flag.StringVar(&keyPath, "key", "", "path to client key")
	flag.BoolVar(&resetDatabase, "r", false, "Reset the test data")
	flag.Parse()
}

func main() {
	getFlags()

	url := fmt.Sprintf("localhost:%v", port)

	info := &mgo.DialInfo{
		Addrs: []string{url},
	}

	if certPath != "" {
		info.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), getTLSConfig())
		}
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	coll := session.DB("test").C("people")

	if resetDatabase {
		coll.DropCollection()
		coll = session.DB("test").C("people")
	}

	if isCollectionEmpty(coll) {
		log.Println("Populating dummy data")
		insertDummyData(coll)

	}

	var records []Person

	query := &struct {
		Tags string
	}{
		Tags: "HR",
	}

	err = coll.Find(query).All(&records)
	if err != nil {
		log.Fatal(err)
	}
	if len(records) == 0 {
		fmt.Println("no records found")
		return
	}
	for _, p := range records {
		fmt.Printf("%+v\n", p)
	}

}

type Person struct {
	Name        string
	Tags        []string
	ContactInfo ContactInfoStruct
}

type ContactInfoStruct struct {
	Phone string
	Tags  []string
}

func insertDummyData(coll *mgo.Collection) {
	record := []*Person{
		&Person{
			Name: "Pritam",
			Tags: []string{"HR", "IT"},
			ContactInfo: ContactInfoStruct{
				Phone: "01234 567890",
				Tags:  []string{"MAN", "HR"},
			},
		},
		&Person{
			Name: "James",
			Tags: []string{"HR", "IT"},
			ContactInfo: ContactInfoStruct{
				Phone: "01234 01234 567890",
				Tags:  []string{"MAN", "IT"},
			},
		},
		&Person{
			Name: "Russell",
			Tags: []string{"IT"},
			ContactInfo: ContactInfoStruct{
				Phone: "01234 567890",
				Tags:  []string{"MAN", "HR"},
			},
		},
	}
	for _, p := range record {
		coll.Insert(p)
	}
}

func isCollectionEmpty(coll *mgo.Collection) bool {
	docCount, err := coll.Count()
	if err != nil {
		log.Println(err)
		return true
	}
	return docCount == 0
}

func getTLSConfig() *tls.Config {
	defer log.Println("TLS Configured")
	return &tls.Config{
		RootCAs: func() *x509.CertPool {
			if len(rootPath) > 0 {
				log.Println("Loading root certificate")
				ca, err := ioutil.ReadFile(rootPath)
				if err != nil {
					panic("Failed to load root certificate")
				}
				rootCAs := x509.NewCertPool()
				if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
					panic("Failed to parse root certificate")
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
