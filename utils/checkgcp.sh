# gcloud auth application-default login

PROJECT_ID=$(gcloud config get-value project)

# Replace with a valid model name and prompt.  This example uses a text generation model.
MODEL_ID="gemini-2.0-flash"
PROMPT="Write a short poem about the ocean."
REGION="us-central1" # Replace with your region

RESPONSE=$(curl \
  -H "Authorization: Bearer $(gcloud auth application-default print-access-token)" \
  -H "Content-Type: application/json" \
  -d "{
  \"instances\": [
    {\"prompt\": \"${PROMPT}\"}
  ],
  \"parameters\": {
    \"temperature\": 0.2,
    \"maxOutputTokens\": 256
  }
}" \
  "https://${REGION}-aiplatform.googleapis.com/v1/projects/${PROJECT_ID}/locations/${REGION}/endpoints/${MODEL_ID}:predict")

echo $RESPONSE

# Check the response for errors.  A successful call will return a JSON response
# containing the prediction.  A permission error will return a JSON response
# with an "error" field.
if echo $RESPONSE | jq '.error' &>/dev/null; then
  echo "API call failed. Check the response for details."
else
  echo "API call successful."
fi
