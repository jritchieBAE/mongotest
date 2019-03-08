package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/globalsign/mgo"
	"gopkg.in/mgo.v2/bson"
)

var (
	port int
)

func getFlags() {
	flag.IntVar(&port, "p", 27017, "mongodb port")
	flag.Parse()
}

func main() {

	getFlags()
	fmt.Println(port)

	session, err := mgo.Dial(fmt.Sprintf("localhost:%v", port))
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

	err = coll.Find(bson.M{}).All(&records)
	if err != nil {
		log.Fatal("record not found")
	}
	fmt.Println(records)

}
