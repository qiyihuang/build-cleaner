# Build Cleaner

Google Cloud function for cleaning up Cloud Build artifact storage after successful deployment.

## Deployment

- Google Cloud Pub/Sub API enabled.

- Create Pub/Sub topic "cloud-builds" if not existing.

- The function needs to have "storage.objects.delete" permission to delete objects from Google Cloud Storage.

- Add environment variables required (shown in .env.example) to the function.
