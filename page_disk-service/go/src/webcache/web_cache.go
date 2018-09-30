package main

import "github.com/mongodb/mongo-go-driver/mongo"
import "github.com/mongodb/mongo-go-driver/bson"
import "context"
import "fmt"
import "github.com/mongodb/mongo-go-driver/mongo/clientopt"
import "time"

type Page struct {
	Url	string `json:"url" bson:"url"`
	Timestamp uint64 `json:"timestamp" bson:"timestamp"`
	Images  []string  `json:"images" bson:"images"`
	Links	[]string  `json:"links"  bson:"links"`
	Scripts []string  `json:"scripts" bson:"scripts"`
	Html	string    `json:"html" bson:"html"`
}


func main() {
        option := clientopt.Auth(clientopt.Credential{AuthSource:"web_cache_db", Username:"web_cache_service",Password:"password" } )
        client, err := mongo.NewClientWithOptions("mongodb://web_cache_service@127.0.0.1:27017/web_cache_db",option)
        if err != nil { fmt.Println(err) }
        err = client.Connect(context.TODO())
        if err != nil { fmt.Println(err) }
        collection := client.Database("web_cache_db").Collection("12345")
	arr := bson.NewArray()
	arr.Append(
            bson.VC.String("!!!"),
            bson.VC.String("???"),
            bson.VC.String("@@@"),
	)

	docs := bson.NewDocument(
				bson.EC.String("url", "journal23"),
				bson.EC.Int64("timestamp", time.Now().Unix()),
				bson.EC.Array("images",
					arr,
				),
				bson.EC.ArrayFromElements("links",
					bson.VC.String("123"),
					bson.VC.String("45"),
				),
				bson.EC.ArrayFromElements("scripts",
                                        bson.VC.String("fsa"),
                                        bson.VC.String("fgg"),
                                ),
				bson.EC.String("html", "<html><html>"),
			)


	fmt.Println(docs)
	index := mongo.NewIndexOptionsBuilder().Unique(true)
	indexModel := mongo.IndexModel{ Keys:bson.NewDocument(bson.EC.String("url", "text")), Options: index.Build()}
	str, errMsg := collection.Indexes().CreateOne(context.Background(), indexModel)
	fmt.Println(str)
	fmt.Println(errMsg)

        res, err := collection.InsertOne(context.Background(), docs)
	if err != nil { fmt.Println(err) }
	result := Page{}
	//filter := map[string][]string{"hello": []string{"john"}}
	err = collection.FindOne(context.Background(), bson.NewDocument(bson.EC.String("url", "journal23"))).Decode(&result)
        if err != nil { fmt.Println(err) }
	fmt.Println(res)
        fmt.Println("\n\n")
	fmt.Println(result)
}

