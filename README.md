# Build Cleaner

Google Cloud function for cleaning up Cloud Build artifact storage after successful deployment. Artifacts are only used during build/deployment. Cloud Build won't clean them up automatically and lifecycle setting in Cloud Storage won't let us delete the artifact immediately.

## Deployment

- Google Cloud Pub/Sub API enabled.

- Create Pub/Sub topic "cloud-builds" if not existing.

- Create a subscription that "push" message to the the function endpoint. Set acknowledge deadline to 260 seconds or higher due to the gap between build completion and deployment completion.

- Grant the service worker build-cleaner uses the "Cloud Functions Viewer" and "Storage Object Admin" permissions.

- The function requires ~20mb memory to run.

- Add environment variables required (shown in .env.example) to the function.
