curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-flash-exp",
    "messages": [
      {
        "role": "developer",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello, I am olivier!"
      },
      {
        "role": "assistant",
        "content": "Hello olivier!"
      },
      {
        "role": "user",
        "content": "What is my name?"
      }
]
  }'
