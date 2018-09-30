package main

import "github.com/mongodb/mongo-go-driver/mongo"
import "context"
import "fmt"
import "github.com/mongodb/mongo-go-driver/mongo/clientopt"


func main() {
        option := clientopt.Auth(clientopt.Credential{AuthSource:"web_cache_db", Username:"web_cache_service",Password:"password" } )
        client, err := mongo.NewClientWithOptions("mongodb://web_cache_service@127.0.0.1:27017",option)
        if err != nil { fmt.Println(err) }
        err = client.Connect(context.TODO())
        if err != nil { fmt.Println(err) }
        collection := client.Database("web_cache_db").Collection("qux")
        res, err := collection.InsertOne(context.Background(), map[string]string{"hello": "world"})
        if err != nil { fmt.Println(err) }
        id := res.InsertedID
        fmt.Println(id)
}

