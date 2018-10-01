package diskclient

import(
	"os"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
	"context"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
	"errors"
)

type DiskClient struct {
}

var (
        initFlag = false
	collection *mongo.Collection
)

type Page struct {
        Url     string `json:"url" bson:"url"`
        Timestamp int64 `json:"timestamp" bson:"timestamp"` // Unix timestamp
        Images  []string  `json:"images" bson:"images"`
        Links   []string  `json:"links"  bson:"links"`
        Scripts []string  `json:"scripts" bson:"scripts"`
        Html    string    `json:"html" bson:"html"`
}

type DC interface {
	// returns 1 is success else -1
	AddPage(page Page) (flag int)

	// returns nil if success an error if fail
	DeletePage(url string) (flag int)

	// returns page if exists else nil
	GetPage(url string) (p Page, err error)

	// refreshes timestamp for a page to current time
	RefreshPageTimestamp(url string) (timestamp int64, err error)
}

func Initialize() (diskClient DC, err error) {
        if !initFlag {
                option := clientopt.Auth(clientopt.Credential{AuthSource:"web_cache_db", Username:"web_cache_service",Password:"password" })
		hp := os.Getenv("DATABASE_HOST") + ":" +os.Getenv("DATABASE_PORT")
		client, err := mongo.NewClientWithOptions("mongodb://web_cache_service@"+ hp +"/web_cache_db",option)
		if err != nil {
			return nil,err
		}
		if err != nil { fmt.Println(err) }
		err = client.Connect(context.TODO())
		collection = client.Database("web_cache_db").Collection("pages")
		indexModel := mongo.IndexModel{ Keys:bson.NewDocument(bson.EC.String("url", "text")), Options: mongo.NewIndexOptionsBuilder().Unique(true).Build()}
                collection.Indexes().CreateOne(context.Background(), indexModel)
                initFlag = true
		diskClient = DiskClient{}
                return diskClient, nil
        } else {
                return nil, errors.New("Mutiple invocations on Initialize not allowed")
        }
}

func (dc DiskClient) AddPage(page Page) (flag int) {
	docs := bson.NewDocument(
                                bson.EC.String("url", page.Url),
                                bson.EC.Int64("timestamp", page.Timestamp),
                                bson.EC.Array("images", toBsonArray(page.Images)),
                                bson.EC.Array("links", toBsonArray(page.Links)),
                                bson.EC.Array("scripts", toBsonArray(page.Scripts)),
                                bson.EC.String("html", page.Html),
                        )
	_, err := collection.InsertOne(context.Background(), docs)
	if err != nil {
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

func (dc DiskClient) GetPage(url string) (p Page, err error){
	result := Page{}
	err = collection.FindOne(context.Background(), bson.NewDocument(bson.EC.String("url", url))).Decode(&result)
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
