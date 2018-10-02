package diskclient

import (
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	//"github.com/mongodb/mongo-go-driver/bson/objectid"
	"context"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

type DiskClient struct {
}

var (
	initFlag = false
	cacheCapacity int64
	collection *mongo.Collection
	diskClient DiskClient
)

type Page struct {
        Url     string `json:"url" bson:"url"`
        Size int64 `json:"size" bson:"size"`
        Safe bool `json:"safe" bson:"safe"`
        TimesUsed int64 `json:"timesUsed" bson:"timesUsed"`
        Timestamp int64 `json:"timestamp" bson:"timestamp"` // Unix timestamp
        Images  []string  `json:"images" bson:"images"`
        Links   []string  `json:"links"  bson:"links"`
        Scripts []string  `json:"scripts" bson:"scripts"`
        Html    string    `json:"html" bson:"html"`
}

type DC interface {
	// returns 1 if success else -1
	AddPage(page Page) (flag int)

	// returns nil if success an error if fail
	DeletePage(url string) (flag int)

	// returns page if exists else nil
	GetPage(url string) (p Page, err error)

	// refreshes timestamp for a page to current time
	RefreshPageTimestamp(url string) (timestamp int64, err error)

	MarkSafe(url string) (err error)

	MarkUnsafe(url string) (err error)

	GetAllPages() (pa []Page, err error)
}

func Initialize(CacheCapacity int64) (client DC, err error) {
        if !initFlag {
                option := clientopt.Auth(clientopt.Credential{AuthSource:"web_cache_db", Username:"web_cache_service",Password:"password" })
		client, err := mongo.NewClientWithOptions("mongodb://web_cache_service@127.0.0.1:27017/web_cache_db",option)
		if err != nil {
			return nil,err
		}
		if err != nil { fmt.Println(err) }
		err = client.Connect(context.TODO())
		collection = client.Database("web_cache_db").Collection("pages")
		indexModel := mongo.IndexModel{ Keys:bson.NewDocument(bson.EC.String("url", "text")), Options: mongo.NewIndexOptionsBuilder().Unique(true).Build()}
                collection.Indexes().CreateOne(context.Background(), indexModel)
                cacheCapacity = CacheCapacity
                initFlag = true
		diskClient = DiskClient{}
                return diskClient, nil
        } else {
        	    // pass already initialized diskClient
                return diskClient, nil
        }
}

func (dc DiskClient) AddPage(page Page) (flag int) {
	docs := bson.NewDocument(
                                bson.EC.String("url", page.Url),
                                bson.EC.Int64("size", page.Size),
                                bson.EC.Boolean("safe", page.Safe),
                                bson.EC.Int64("timesUsed", page.TimesUsed),
                                bson.EC.Int64("timestamp", page.Timestamp),
                                bson.EC.Array("images", toBsonArray(page.Images)),
                                bson.EC.Array("links", toBsonArray(page.Links)),
                                bson.EC.Array("scripts", toBsonArray(page.Scripts)),
                                bson.EC.String("html", page.Html),
                        )
	_, err := collection.InsertOne(context.Background(), docs)
	if err != nil {
		fmt.Println(page.Url)
		fmt.Println(err)
		return -1
	}
	return 1
}

func toBsonArray(arr []string) (bsonArr *(bson.Array)){
	bsonArr = bson.NewArray()
	for _, str := range arr {
		bsonArr.Append(bson.VC.String(str))
        }
	return bsonArr
}

func (dc DiskClient) DeletePage(url string) (flag int){
	res ,err := collection.DeleteOne(context.Background(), bson.NewDocument(bson.EC.String("url", url)))
        if err != nil {
		fmt.Println(err)
		return -1
	} else if res.DeletedCount == 0 {
		return -1
	}
	return 1
}

func (dc DiskClient) GetPage(url string) (pa Page, err error){
	result := Page{}
	err = collection.FindOne(context.Background(), bson.NewDocument(bson.EC.String("url", url))).Decode(&result)
        if err != nil {
		fmt.Println(err)
		return result, err
	}
	return result , nil
}

// TODO: test this implementation of getting all pages
func (dc DiskClient) GetAllPages() (pa []Page, err error){
	var result []Page
	cursor, err := collection.Find(
		context.Background(),
		bson.NewDocument(
			bson.EC.SubDocumentFromElements("pages",
				bson.EC.ArrayFromElements("$all"),
			),
		))
	cursor.Decode(&result)
	if err != nil {
		fmt.Println(err)
		return result, err
	}
	return result , nil
}

func (dc DiskClient) RefreshPageTimestamp(url string) (timestamp int64, err error){
	timestamp = time.Now().Unix()
	result, err := collection.UpdateOne(
			context.Background(),
			bson.NewDocument(
				bson.EC.String("url", url),
			),
			bson.NewDocument(
				bson.EC.SubDocumentFromElements("$set",
					bson.EC.Int64("timestamp", timestamp),
				),
			),
		)
	fmt.Println(result)
	if err != nil {return -1,err}
	return timestamp,nil
}

func (dc DiskClient) MarkSafe(url string) (err error){
	result, err := collection.UpdateOne(
		context.Background(),
		bson.NewDocument(
			bson.EC.String("url", url),
		),
		bson.NewDocument(
			bson.EC.SubDocumentFromElements("$set",
				bson.EC.Boolean("safe", true),
			),
		),
	)
	fmt.Println(result)
	if err != nil {return err}
	return nil
}

func (dc DiskClient) MarkUnsafe(url string) (err error){
	result, err := collection.UpdateOne(
		context.Background(),
		bson.NewDocument(
			bson.EC.String("url", url),
		),
		bson.NewDocument(
			bson.EC.SubDocumentFromElements("$set",
				bson.EC.Boolean("safe", false),
			),
		),
	)
	fmt.Println(result)
	if err != nil {return err}
	return nil
}