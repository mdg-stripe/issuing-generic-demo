#!/bin/bash
export PROJECT_ID=""
export CODE_STAGE_BUCKET="${PROJECT_ID}_code_stage_bucket"
gsutil mb gs://${CODE_STAGE_BUCKET}
echo "==="
echo " * If you get an error message saying the bucket already exists,"
echo "   and you previously created the bucket, ignore it. That's fine."
echo " * If you get an error message saying you don't have permissions on that bucket,"
echo "   DON'T ignore it; either fix the permissions, or create a new bucket."
echo "==="

gcloud functions deploy RealTimeAuthWebhook \
    --region=us-central1 \
    --allow-unauthenticated \
    --entry-point=HandleWebhook \
    --runtime=go116 \
    --security-level=secure-always \
    --source=`pwd` \
    --stage-bucket=${CODE_STAGE_BUCKET} \
    --env-vars-file=env.yaml \
    --trigger-http 