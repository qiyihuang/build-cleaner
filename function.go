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
	"google.golang.org/api/cloudfunctions/v1"
	"google.golang.org/api/iterator"
)

var webhookClient *messenger.Client
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

func (bs BuildStatus) string() string {
	return []string{"SUCCESS", "FAILURE", "CANCELLED", "TIMEOUT", "FAILED"}[bs]
}

func Clean(w http.ResponseWriter, r *http.Request) {
	var m pubsubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		internalError(w, err, "Failed to decode request body.")
		return
	}

	desc, color := notifyParams(m)
	if err := notify(desc, color); err != nil {
		internalError(w, err, "notify: ")
		return
	}

	isLast, err := waitDeploy()
	if err != nil {
		internalError(w, err, "waitDeploy: ")
	}
	if !isLast {
		notify("One resource deployed, waiting for other to complete.", GREEN)
		// Return 200 without cleaning up the bucket, wait for the last resource finished deploying to do that.
		return
	}

	if err := cleanBuckets(); err != nil {
		log.Println("deleteBuckets: ", err)
		if err := notify("Delete bucket failed, please check not and delete buckets manually.", RED); err != nil {
			log.Println("notify: ", err)
		}
		w.WriteHeader(500)
		return
	}

	if err := notify("Cloud Build artifact buckets cleaned.", GREEN); err != nil {
		internalError(w, err, "notify: ")
	}
}

func notifyParams(m pubsubMessage) (string, int) {
	status := m.Message.Attributes.Status
	desc := "Build status: " + status + "."
	var color int
	if status == Success.string() {
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

	clt, err := client()
	if err != nil {
		return err
	}
	_, err = clt.Send(msgs)
	if err != nil {
		return err
	}
	return nil
}

func waitDeploy() (isLast bool, err error) {
	// Since there's a delay between build completion and resource deployment completion,
	// startCount should at least be 1.
	startCount, err := countDeployInProgress()
	if err != nil {
		return
	}
	// Check if there's a function finish deploying every 10 seconds
	var count uint
	for {
		time.Sleep(10 * time.Second)
		if count, err = countDeployInProgress(); count < startCount {
			if err != nil {
				return
			}
			break
		}
	}
	return count == 0, nil
}

func countDeployInProgress() (uint, error) {
	service, err := cloudfunctions.NewService(context.Background())
	if err != nil {
		return 0, err
	}
	parent := "projects/" + os.Getenv("PROJECT_NAME") + "/locations/-"
	resp, err := service.Projects.Locations.Functions.List(parent).Do()
	if err != nil {
		return 0, err
	}

	var inProgressCount uint
	for _, fn := range resp.Functions {
		if fn.Status == "DEPLOY_IN_PROGRESS" {
			inProgressCount += 1
		}
	}
	return inProgressCount, nil
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

func client() (*messenger.Client, error) {
	if webhookClient == nil {
		var err error
		webhookClient, err = messenger.NewClient(http.DefaultClient, os.Getenv("DISCORD_WEBHOOK_URL"))
		if err != nil {
			return nil, err
		}
	}
	return webhookClient, nil
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
	log.Println(msg, err.Error())
	w.WriteHeader(500)
}
