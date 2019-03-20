package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	// "github.com/globalsign/mgo"
	// "gopkg.in/mgo.v2/bson"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	Name  string
	Grade string
}

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

	url := fmt.Sprintf("mongodb://localhost:%v", port)
	log.Println(url)
	opt := options.Client().ApplyURI(url)
	if len(rootPath) > 0 {
		tlsConfig := getTLSConfig()
		opt.TLSConfig = tlsConfig
	}
	ctx, _ := context.WithTimeout(context.Background(), 8*time.Second)
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		log.Fatal(err)
	}

	// if err = client.Ping(ctx, nil); err != nil {
	// 	log.Fatal(err)
	// }

	coll := client.Database("test").Collection("people")

	// record := &Person{Name: "Bob", Grade: "T3"}

	// coll.InsertOne(ctx, record)

	// var records []struct {
	// 	Name  string
	// 	Grade string
	// }
	log.Println("Querying...")

	cur, err := coll.Find(ctx, bson.M{"name": "Bob"})
	if err != nil {
		log.Fatal("record not found")
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}
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
