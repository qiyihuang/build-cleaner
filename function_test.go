package function

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildStatusString(t *testing.T) {
	t.Run("Returns correct status", func(t *testing.T) {
		require.Equal(t, "SUCCESS", Success.string(), "Status string incorrect")
		require.Equal(t, "FAILURE", Failure.string(), "Status string incorrect")
		require.Equal(t, "CANCELLED", Cancelled.string(), "Status string incorrect")
		require.Equal(t, "TIMEOUT", Timeout.string(), "Status string incorrect")
		require.Equal(t, "FAILED", Failed.string(), "Status string incorrect")
	})
}

func TestNotifyParams(t *testing.T) {
	t.Run("Green", func(t *testing.T) {
		m := pubsubMessage{
			Message: message{
				Attributes: attributes{
					Status: "SUCCESS",
				},
			},
		}

		desc, color := notifyParams(m)

		require.Equal(t, "Build status: SUCCESS.", desc, "Incorrect description")
		require.Equal(t, GREEN, color, "Incorrect color")
	})

	t.Run("Red", func(t *testing.T) {
		msgs := []pubsubMessage{
			{
				Message: message{
					Attributes: attributes{
						Status: "FAILURE",
					},
				},
			},
			{
				Message: message{
					Attributes: attributes{
						Status: "CANCELLED",
					},
				},
			},
			{
				Message: message{
					Attributes: attributes{
						Status: "TIMEOUT",
					},
				},
			},
			{
				Message: message{
					Attributes: attributes{
						Status: "FAILED",
					},
				},
			},
		}

		for _, m := range msgs {
			desc, color := notifyParams(m)
			require.Equal(t, "Build status: "+m.Message.Attributes.Status+".", desc, "Incorrect description")
			require.Equal(t, RED, color, "Incorrect color")
		}
	})
}

func TestClient(t *testing.T) {
	t.Run("httpClient is nil", func(t *testing.T) {
		httpClient = nil

		c := client()

		require.Equal(t, http.DefaultClient, c, "Incorrect httpClient")
	})

	t.Run("httpClient exists", func(t *testing.T) {
		httpClient = &http.Client{Timeout: 1 * time.Second}

		c := client()

		require.NotEqual(t, http.DefaultClient, c, "Incorrect httpClient")
	})
}

// Todo check how to set up google creds for GitHub.
// Todo using storage.NewClient requires proper creds.
// func TestBucket(t *testing.T) {
// 	t.Run("bucketHandle is nil", func(t *testing.T) {
// 		bucketHandle = nil
// 		// c, _ := storage.NewClient(context.Background())
// 		n := "name"
// 		os.Setenv("ARTIFACT_BUCKET_NAME", n)
// 		defer os.Unsetenv("ARTIFACT_BUCKET_NAME")
// 		// expectedBucket := c.Bucket(n)

// 		b, err := bucket()

// 		// require.Equal(t, expectedBucket, b, "Incorrect bucketHandle")
// 		require.Equal(t, nil, err, "Incorrect bucketHandle")
// 		require.NotEqual(t, nil, b, "Incorrect bucketHandle")
// 	})
// }
