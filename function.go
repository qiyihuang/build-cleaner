package function

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/qiyihuang/messenger"
	"google.golang.org/api/iterator"
)

var httpClient *http.Client
var bucketHandle *storage.BucketHandle

const (
	GREEN int = 5763719
	RED   int = 15548997
)

type pubsubMessage struct {
	Message message `json:"message"`
}

type message struct {
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	Status string `json:"status"`
}

type BuildStatus uint8

const (
	Success BuildStatus = iota
	Failure
	Cancelled
	Timeout
	Failed
)

func (bs BuildStatus) String() string {
	return []string{"SUCCESS", "FAILURE", "CANCELLED", "TIMEOUT", "FAILED"}[bs]
}

func Clean(w http.ResponseWriter, r *http.Request) {
	var m pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		internalError(w, err, "Failed to decode request body.")
		return
	}

	desc, color := notifyParams(m)
	log.Println(desc)
	if err := notify(desc, color); err != nil {
		internalError(w, err, "notify: ")
		return
	}

	// There's a delay between build completion and function being 'active'.
	time.Sleep(2 * time.Minute)
	if err := cleanBuckets(); err != nil {
		log.Println("deleteBuckets: ", err)
		if err := notify("Delete bucket failed, please check not and delete buckets manually.", RED); err != nil {
			log.Println("notify: ", err)
		}
		w.WriteHeader(500)
		return
	}

	if err := notify("Cloud Build source and artifact buckets deleted.", GREEN); err != nil {
		internalError(w, err, "notify: ")
	}
}

func notifyParams(m pubsubMessage) (string, int) {
	status := m.Message.Attributes.Status
	desc := "Build status: " + status + "."
	var color int
	if status == Success.String() {
		color = GREEN
	} else {
		color = RED
	}
	return desc, color
}

func notify(description string, color int) error {
	msgs := []messenger.Message{{
		Username: os.Getenv("DISCORD_WEBHOOK_USERNAME"),
		Embeds:   []messenger.Embed{{Title: "Google Cloud Build", Description: description, Color: color}},
	}}

	req, err := messenger.NewRequest(client(), os.Getenv("DISCORD_WEBHOOK_URL"), msgs)
	if err != nil {
		return err
	}

	_, err = req.Send()
	if err != nil {
		return err
	}
	return nil
}

func cleanBuckets() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bkt, err := bucket()
	if err != nil {
		return err
	}
	objIt := bkt.Objects(ctx, nil)
	names, err := objectNames(objIt)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, name := range names {
		wg.Add(1)
		go deleteObject(ctx, &wg, name, bkt)
	}
	wg.Wait()
	return nil
}

func objectNames(objIt *storage.ObjectIterator) ([]string, error) {
	var names []string
	for {
		objAttrs, err := objIt.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		names = append(names, objAttrs.Name)
	}
	return names, nil
}

func deleteObject(ctx context.Context, wg *sync.WaitGroup, name string, bkt *storage.BucketHandle) {
	defer wg.Done()

	err := bkt.Object(name).Delete(ctx)
	if err != nil {
		log.Println(err)
	}
}

func client() *http.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return httpClient
}

func bucket() (*storage.BucketHandle, error) {
	if bucketHandle == nil {
		client, err := storage.NewClient(context.Background())
		if err != nil {
			return nil, err
		}
		bucketHandle = client.Bucket(os.Getenv("ARTIFACT_BUCKET_NAME"))
	}
	return bucketHandle, nil
}

func internalError(w http.ResponseWriter, err error, msg string) {
	log.Println(msg, err)
	w.WriteHeader(500)
}
