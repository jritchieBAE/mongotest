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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	port          int
	rootPath      string
	certPath      string
	keyPath       string
	ctx, _        = context.WithTimeout(context.Background(), 8*time.Second)
	client        *mongo.Client
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
	url := fmt.Sprintf("mongodb://localhost:%v", port)

	client = getMongoClient(url)
	coll := client.Database("test").Collection("people")

	if resetDatabase {
		coll.DeleteMany(ctx, bson.M{})
	}

	if isCollectionEmpty(coll) {
		log.Println("Populating dummy data")
		insertDummyData(coll)
	}

	userTags2 := bson.A{"HR", "IT", "MAN"}
	query := bson.M{}

	mongopl := mongo.Pipeline{
		bson.D{
			{"$match", query}},
		bson.D{
			{"$redact", bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$gt": bson.A{bson.M{"$size": bson.M{"$setIntersection": bson.A{"$tags", userTags2}}}, 0}},
					"then": "$$DESCEND",
					"else": "$$PRUNE"}}}},
	}

	cur, err := coll.Aggregate(ctx, mongopl, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	printResults(cur)
}

func insertDummyData(coll *mongo.Collection) {
	record := []interface{}{
		bson.D{
			{"name", "Pritam"},
			{"tags", bson.A{"HR", "IT"}},
			{"contact info", bson.A{
				bson.D{
					{"phone", "01234 567890"},
					{"tags", bson.A{"MAN", "HR"}},
				},
				bson.D{
					{"email", "pritam@bae.com"},
					{"tags", bson.A{"IT"}},
				}}}},
		bson.D{
			{"name", "James"},
			{"tags", bson.A{"HR", "IT"}},
			{"contact info", bson.A{
				bson.D{
					{"phone", "01234 567890"},
					{"tags", bson.A{"HR", "MAN"}},
				},
				bson.D{
					{"email", "james@bae.com"},
					{"tags", bson.A{"IT"}}}}}},
		bson.D{
			{"name", "Russ"},
			{"tags", bson.A{"HR", "IT"}},
			{"contact info", bson.A{
				bson.D{
					{"phone", "01234 567890"},
					{"tags", bson.A{"HR", "MAN"}},
				},
				bson.D{
					{"email", "russell@bae.com"},
					{"tags", bson.A{"IT", "MAN"}},
				}}}},
	}
	coll.InsertMany(ctx, record, nil)
}

func isCollectionEmpty(coll *mongo.Collection) bool {
	docCount, err := coll.CountDocuments(ctx, bson.M{}, nil)
	if err != nil {
		log.Println(err)
		return true
	}
	return docCount == 0
}

func printResults(cur *mongo.Cursor) {
	noRecords := true
	for cur.Next(ctx) {
		noRecords = false
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v\n", result)
	}
	if noRecords {
		fmt.Println("No matches.")
	}
}

func getMongoClient(url string) *mongo.Client {
	opt := options.Client().ApplyURI(url)
	if len(rootPath) > 0 {
		tlsConfig := getTLSConfig()
		opt.TLSConfig = tlsConfig
	}
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	return client
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
